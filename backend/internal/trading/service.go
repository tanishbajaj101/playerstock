package trading

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stakestock/backend/internal/accounts"
	"github.com/stakestock/backend/internal/engine"
	"github.com/stakestock/backend/internal/models"
	"github.com/stakestock/backend/internal/pubsub"
	ob "github.com/stakestock/backend/internal/orderbook"
)

type PlaceOrderRequest struct {
	AssetSymbol string           `json:"asset_symbol"`
	Side        models.OrderSide `json:"side"`
	Type        models.OrderType `json:"type"`
	Qty         decimal.Decimal  `json:"qty"`
	Price       *decimal.Decimal `json:"price"` // nil for market
}

type Service struct {
	pool     *pgxpool.Pool
	registry *engine.Registry
	bus      *pubsub.Bus
}

func NewService(pool *pgxpool.Pool, reg *engine.Registry, bus *pubsub.Bus) *Service {
	return &Service{pool: pool, registry: reg, bus: bus}
}

func (s *Service) PlaceOrder(ctx context.Context, userID uuid.UUID, req PlaceOrderRequest) (*models.Order, error) {
	if req.Qty.LessThanOrEqual(decimal.Zero) {
		return nil, fmt.Errorf("quantity must be positive")
	}
	if !req.Qty.Equal(req.Qty.Floor()) {
		return nil, fmt.Errorf("quantity must be a whole number")
	}
	if req.Qty.GreaterThan(decimal.NewFromInt(5)) {
		return nil, fmt.Errorf("quantity cannot exceed 5 per order")
	}
	if req.Type == models.OrderTypeLimit && req.Price == nil {
		return nil, fmt.Errorf("limit order requires a price")
	}
	if req.Type == models.OrderTypeMarket {
		req.Price = nil
	}

	// Resolve asset
	var assetID uuid.UUID
	err := s.pool.QueryRow(ctx, `SELECT id FROM assets WHERE symbol=$1`, req.AssetSymbol).Scan(&assetID)
	if err != nil {
		return nil, fmt.Errorf("asset not found: %s", req.AssetSymbol)
	}

	ab, ok := s.registry.Get(req.AssetSymbol)
	if !ok {
		return nil, fmt.Errorf("engine not available for %s", req.AssetSymbol)
	}

	orderID := uuid.New()

	// reservation state tracked across phases
	var (
		reservedCash     decimal.Decimal
		reservedPosition decimal.Decimal
		isShort          bool
	)

	// ---- Phase 1: Reservation transaction ----
	err = pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{IsoLevel: pgx.RepeatableRead}, func(tx pgx.Tx) error {
		pos, err := accounts.GetOrCreatePosition(ctx, tx, userID, assetID)
		if err != nil {
			return err
		}

		switch req.Side {
		case models.SideBuy:
			if pos.Qty.Add(req.Qty).GreaterThan(decimal.NewFromInt(5)) {
				remaining := decimal.NewFromInt(5).Sub(pos.Qty)
				if remaining.LessThanOrEqual(decimal.Zero) {
					return fmt.Errorf("position limit reached: you already hold 5 units of this asset")
				}
				return fmt.Errorf("position limit exceeded: you can buy at most %s more unit(s) of this asset", remaining.String())
			}
			if req.Type == models.OrderTypeLimit {
				reserve := req.Qty.Mul(*req.Price)
				if err := accounts.ReserveCashForBuy(ctx, tx, userID, reserve); err != nil {
					return err
				}
				reservedCash = reserve
			} else {
				// Market buy: estimate via engine
				r := ab.MarketPrice(ob.Buy, req.Qty)
				if r.Err != nil {
					return fmt.Errorf("cannot estimate market price: %w", r.Err)
				}
				if r.MarketPriceResult.LessThanOrEqual(decimal.Zero) {
					return fmt.Errorf("no liquidity in ask book")
				}
				if err := accounts.ReserveCashForBuy(ctx, tx, userID, r.MarketPriceResult); err != nil {
					return err
				}
				reservedCash = r.MarketPriceResult
			}

		case models.SideSell:
			available := pos.Qty.Sub(pos.LockedQty)
			if req.Type == models.OrderTypeMarket {
				if available.LessThan(req.Qty) {
					return fmt.Errorf("short sells must use limit orders; insufficient long position for market sell")
				}
				if err := accounts.ReservePositionForSell(ctx, tx, userID, assetID, req.Qty); err != nil {
					return err
				}
				reservedPosition = req.Qty
			} else {
				// Limit sell: split covered / short
				coveredQty := decimal.Max(decimal.Min(available, req.Qty), decimal.Zero)
				shortQty := req.Qty.Sub(coveredQty)

				if coveredQty.GreaterThan(decimal.Zero) {
					if err := accounts.ReservePositionForSell(ctx, tx, userID, assetID, coveredQty); err != nil {
						return err
					}
					reservedPosition = coveredQty
				}
				if shortQty.GreaterThan(decimal.Zero) {
					collateral := shortQty.Mul(*req.Price)
					if err := accounts.ReserveCashForShort(ctx, tx, userID, collateral); err != nil {
						return err
					}
					reservedCash = collateral
					isShort = true
				}
			}
		}

		now := time.Now().UTC()
		_, err = tx.Exec(ctx, `
			INSERT INTO orders (id, user_id, asset_id, side, type, qty, filled_qty, price, status, is_short, created_at, updated_at)
			VALUES ($1,$2,$3,$4,$5,$6,0,$7,'open',$8,$9,$9)
		`, orderID, userID, assetID, int16(req.Side), string(req.Type), req.Qty, req.Price, isShort, now)
		return err
	})
	if err != nil {
		return nil, fmt.Errorf("reservation failed: %w", err)
	}

	// ---- Phase 2: Call matching engine ----
	var reply engine.Reply
	obSide := engine.SideFromModel(req.Side)
	if req.Type == models.OrderTypeLimit {
		reply = ab.PlaceLimit(obSide, orderID.String(), req.Qty, *req.Price)
	} else {
		reply = ab.PlaceMarket(obSide, req.Qty)
	}

	if reply.Err != nil {
		// Compensate
		_ = s.compensate(ctx, orderID, userID, assetID, reservedCash, reservedPosition)
		return nil, fmt.Errorf("engine error: %w", reply.Err)
	}

	// ---- Phase 3: Settlement transaction ----
	order, err := s.settle(ctx, userID, assetID, orderID, req, reply, reservedCash, reservedPosition, isShort)
	if err != nil {
		return nil, fmt.Errorf("settlement: %w", err)
	}

	go s.publishDepth(ab)
	return order, nil
}

func (s *Service) compensate(ctx context.Context, orderID, userID, assetID uuid.UUID, cash, position decimal.Decimal) error {
	return pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if cash.GreaterThan(decimal.Zero) {
			_ = accounts.ReleaseCashReservation(ctx, tx, userID, cash)
		}
		if position.GreaterThan(decimal.Zero) {
			_ = accounts.ReleasePositionReservation(ctx, tx, userID, assetID, position)
		}
		_, err := tx.Exec(ctx, `UPDATE orders SET status='rejected', updated_at=now() WHERE id=$1`, orderID)
		return err
	})
}

func (s *Service) settle(
	ctx context.Context,
	userID, assetID, orderID uuid.UUID,
	req PlaceOrderRequest,
	reply engine.Reply,
	reservedCash, reservedPosition decimal.Decimal,
	isShort bool,
) (*models.Order, error) {
	var finalOrder *models.Order

	err := pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{IsoLevel: pgx.ReadCommitted}, func(tx pgx.Tx) error {
		var trades []models.Trade
		totalFilled := decimal.Zero
		remainingReservedPosition := reservedPosition

		// Process fully-matched resting orders
		for _, doneObOrder := range reply.Done {
			doneID, err := uuid.Parse(doneObOrder.ID())
			if err != nil || doneID == orderID {
				continue
			}

			var resting restingOrderRow
			if err := s.fetchRestingOrder(ctx, tx, doneID, &resting); err != nil {
				return err
			}

			tradeQty := resting.qty.Sub(resting.filledQty)
			trade := s.buildTrade(assetID, orderID, doneID, userID, resting.userID, req.Side, tradeQty, resting.price)
			trades = append(trades, trade)
			totalFilled = totalFilled.Add(tradeQty)

			if err := s.applyRestingFill(ctx, tx, resting, assetID, tradeQty, resting.price); err != nil {
				return err
			}
			if err := s.applyIncomingFill(ctx, tx, userID, assetID, req, trade, resting.price, &remainingReservedPosition, isShort); err != nil {
				return err
			}

			_, err = tx.Exec(ctx, `UPDATE orders SET status='filled', filled_qty=qty, updated_at=now() WHERE id=$1`, doneID)
			if err != nil {
				return err
			}
		}

		// Process partial resting order
		if reply.Partial != nil {
			partialID, err := uuid.Parse(reply.Partial.ID())
			if err == nil && partialID != orderID {
				var resting restingOrderRow
				if err := s.fetchRestingOrder(ctx, tx, partialID, &resting); err != nil {
					return err
				}

				tradeQty := reply.PartialQuantityProcessed
				trade := s.buildTrade(assetID, orderID, partialID, userID, resting.userID, req.Side, tradeQty, resting.price)
				trades = append(trades, trade)
				totalFilled = totalFilled.Add(tradeQty)

				if err := s.applyRestingFill(ctx, tx, resting, assetID, tradeQty, resting.price); err != nil {
					return err
				}
				if err := s.applyIncomingFill(ctx, tx, userID, assetID, req, trade, resting.price, &remainingReservedPosition, isShort); err != nil {
					return err
				}

				_, err = tx.Exec(ctx, `
					UPDATE orders SET status='partial', filled_qty=filled_qty+$2, updated_at=now() WHERE id=$1
				`, partialID, tradeQty)
				if err != nil {
					return err
				}
			}
		}

		// Insert trades
		for _, t := range trades {
			_, err := tx.Exec(ctx, `
				INSERT INTO trades (id, asset_id, buy_order_id, sell_order_id, buy_user_id, sell_user_id, qty, price, created_at)
				VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
			`, t.ID, t.AssetID, t.BuyOrderID, t.SellOrderID, t.BuyUserID, t.SellUserID, t.Qty, t.Price, t.CreatedAt)
			if err != nil {
				return fmt.Errorf("insert trade: %w", err)
			}
		}

		// Determine final status
		status := models.OrderStatusOpen
		if totalFilled.GreaterThanOrEqual(req.Qty) {
			status = models.OrderStatusFilled
		} else if totalFilled.GreaterThan(decimal.Zero) {
			status = models.OrderStatusPartial
		}

		// Release over-reservation for market buy
		if status == models.OrderStatusFilled && req.Type == models.OrderTypeMarket && req.Side == models.SideBuy {
			actualCost := decimal.Zero
			for _, t := range trades {
				actualCost = actualCost.Add(t.Qty.Mul(t.Price))
			}
			overReserved := reservedCash.Sub(actualCost)
			if overReserved.GreaterThan(decimal.Zero) {
				if err := accounts.ReleaseCashReservation(ctx, tx, userID, overReserved); err != nil {
					log.Printf("release over-reservation: %v", err)
				}
			}
		}

		_, err := tx.Exec(ctx, `UPDATE orders SET status=$2, filled_qty=$3, updated_at=now() WHERE id=$1`,
			orderID, string(status), totalFilled)
		if err != nil {
			return err
		}

		finalOrder = &models.Order{}
		var priceVal *decimal.Decimal
		err = tx.QueryRow(ctx, `
			SELECT id, user_id, asset_id, side, type, qty, filled_qty, price, status, is_short, created_at, updated_at
			FROM orders WHERE id=$1
		`, orderID).Scan(
			&finalOrder.ID, &finalOrder.UserID, &finalOrder.AssetID,
			&finalOrder.Side, &finalOrder.Type, &finalOrder.Qty, &finalOrder.FilledQty,
			&priceVal, &finalOrder.Status, &finalOrder.IsShort,
			&finalOrder.CreatedAt, &finalOrder.UpdatedAt,
		)
		finalOrder.Price = priceVal
		if err != nil {
			return err
		}

		// Async publish — one message per order, not per trade
		orderSnap := *finalOrder
		go func() {
			bgCtx := context.Background()
			if len(trades) > 0 {
				_ = s.bus.Publish(bgCtx, pubsub.AssetTradesChan(req.AssetSymbol), trades)
				_ = s.bus.CacheSet(bgCtx, fmt.Sprintf("asset:%s:lastTrade", req.AssetSymbol), trades[len(trades)-1].Price)
			}
			_ = s.bus.Publish(bgCtx, pubsub.UserOrdersChan(userID.String()), orderSnap)
		}()

		return nil
	})
	return finalOrder, err
}

type restingOrderRow struct {
	id        uuid.UUID
	userID    uuid.UUID
	qty       decimal.Decimal
	filledQty decimal.Decimal
	price     decimal.Decimal
	side      models.OrderSide
	isShort   bool
}

func (s *Service) fetchRestingOrder(ctx context.Context, tx pgx.Tx, id uuid.UUID, out *restingOrderRow) error {
	return tx.QueryRow(ctx, `
		SELECT id, user_id, qty, filled_qty, price, side, is_short FROM orders WHERE id=$1
	`, id).Scan(&out.id, &out.userID, &out.qty, &out.filledQty, &out.price, &out.side, &out.isShort)
}

func (s *Service) buildTrade(assetID, incomingID, restingID, incomingUser, restingUser uuid.UUID, incomingSide models.OrderSide, qty, price decimal.Decimal) models.Trade {
	var buyOrderID, sellOrderID, buyUserID, sellUserID uuid.UUID
	if incomingSide == models.SideBuy {
		buyOrderID, buyUserID = incomingID, incomingUser
		sellOrderID, sellUserID = restingID, restingUser
	} else {
		sellOrderID, sellUserID = incomingID, incomingUser
		buyOrderID, buyUserID = restingID, restingUser
	}
	return models.Trade{
		ID:          uuid.New(),
		AssetID:     assetID,
		BuyOrderID:  buyOrderID,
		SellOrderID: sellOrderID,
		BuyUserID:   buyUserID,
		SellUserID:  sellUserID,
		Qty:         qty,
		Price:       price,
		CreatedAt:   time.Now().UTC(),
	}
}

func (s *Service) applyRestingFill(ctx context.Context, tx pgx.Tx, resting restingOrderRow, assetID uuid.UUID, qty, price decimal.Decimal) error {
	if resting.side == models.SideBuy {
		return accounts.ApplyTradeBuy(ctx, tx, resting.userID, assetID, qty, price, price)
	}
	if resting.isShort {
		return accounts.ApplyTradeSellShort(ctx, tx, resting.userID, assetID, qty, price, price)
	}
	return accounts.ApplyTradeSellCovered(ctx, tx, resting.userID, assetID, qty, price)
}

func (s *Service) applyIncomingFill(
	ctx context.Context, tx pgx.Tx,
	userID, assetID uuid.UUID,
	req PlaceOrderRequest,
	trade models.Trade,
	tradePrice decimal.Decimal,
	remainingReservedPosition *decimal.Decimal,
	isShort bool,
) error {
	if req.Side == models.SideBuy {
		reservedPrice := tradePrice
		if req.Price != nil {
			reservedPrice = *req.Price
		}
		return accounts.ApplyTradeBuy(ctx, tx, userID, assetID, trade.Qty, tradePrice, reservedPrice)
	}

	// Sell side: consume covered first, then short
	coveredQty := decimal.Min(trade.Qty, *remainingReservedPosition)
	shortQty := trade.Qty.Sub(coveredQty)
	*remainingReservedPosition = remainingReservedPosition.Sub(coveredQty)

	if coveredQty.GreaterThan(decimal.Zero) {
		if err := accounts.ApplyTradeSellCovered(ctx, tx, userID, assetID, coveredQty, tradePrice); err != nil {
			return err
		}
	}
	if shortQty.GreaterThan(decimal.Zero) {
		reservedPrice := tradePrice
		if req.Price != nil {
			reservedPrice = *req.Price
		}
		if err := accounts.ApplyTradeSellShort(ctx, tx, userID, assetID, shortQty, tradePrice, reservedPrice); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) CancelOrder(ctx context.Context, userID, orderID uuid.UUID) error {
	var order models.Order
	var priceVal *decimal.Decimal
	err := s.pool.QueryRow(ctx, `
		SELECT id, user_id, asset_id, side, type, qty, filled_qty, price, status, is_short
		FROM orders WHERE id=$1 AND user_id=$2
	`, orderID, userID).Scan(
		&order.ID, &order.UserID, &order.AssetID, &order.Side, &order.Type,
		&order.Qty, &order.FilledQty, &priceVal, &order.Status, &order.IsShort,
	)
	if err != nil {
		return fmt.Errorf("order not found or not yours")
	}
	order.Price = priceVal

	if order.Status != models.OrderStatusOpen && order.Status != models.OrderStatusPartial {
		return fmt.Errorf("order cannot be cancelled (status: %s)", order.Status)
	}

	var symbol string
	if err := s.pool.QueryRow(ctx, `SELECT symbol FROM assets WHERE id=$1`, order.AssetID).Scan(&symbol); err != nil {
		return err
	}

	ab, ok := s.registry.Get(symbol)
	if !ok {
		return fmt.Errorf("engine not available")
	}

	_ = ab.Cancel(orderID.String()) // best-effort; may already be filled

	remaining := order.Qty.Sub(order.FilledQty)

	return pgx.BeginTxFunc(ctx, s.pool, pgx.TxOptions{}, func(tx pgx.Tx) error {
		if order.Side == models.SideBuy && order.Price != nil {
			_ = accounts.ReleaseCashReservation(ctx, tx, userID, remaining.Mul(*order.Price))
		} else if order.Side == models.SideSell {
			if order.IsShort && order.Price != nil {
				_ = accounts.ReleaseCashReservation(ctx, tx, userID, remaining.Mul(*order.Price))
			} else {
				_ = accounts.ReleasePositionReservation(ctx, tx, userID, order.AssetID, remaining)
			}
		}
		_, err := tx.Exec(ctx, `UPDATE orders SET status='cancelled', updated_at=now() WHERE id=$1`, orderID)
		if err != nil {
			return err
		}
		order.Status = models.OrderStatusCancelled
		go func() {
			bgCtx := context.Background()
			_ = s.bus.Publish(bgCtx, pubsub.UserOrdersChan(userID.String()), order)
		}()
		return nil
	})
}

func (s *Service) publishDepth(ab *engine.AssetBook) {
	r := ab.Depth()
	event := models.DepthEvent{AssetSymbol: ab.Symbol}
	for _, l := range r.Bids {
		event.Bids = append(event.Bids, models.PriceLevel{Price: l.Price, Quantity: l.Quantity})
	}
	for _, l := range r.Asks {
		event.Asks = append(event.Asks, models.PriceLevel{Price: l.Price, Quantity: l.Quantity})
	}
	bgCtx := context.Background()
	_ = s.bus.Publish(bgCtx, pubsub.AssetDepthChan(ab.Symbol), event)
	_ = s.bus.CacheSet(bgCtx, fmt.Sprintf("asset:%s:depth", ab.Symbol), event)
}
