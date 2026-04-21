package engine

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	ob "github.com/stakestock/backend/internal/orderbook"
)

type reqKind int

const (
	reqLimit reqKind = iota
	reqMarket
	reqCancel
	reqDepth
	reqMarketPrice
)

type engineReq struct {
	kind    reqKind
	orderID string
	side    ob.Side
	qty     decimal.Decimal
	price   decimal.Decimal
	replyCh chan Reply
}

// DepthLevel is a single price level in the order book depth.
type DepthLevel struct {
	Price    decimal.Decimal
	Quantity decimal.Decimal
}

// Reply is the result of an engine operation.
type Reply struct {
	Done                     []*ob.Order
	Partial                  *ob.Order
	PartialQuantityProcessed decimal.Decimal
	QuantityLeft             decimal.Decimal
	Bids                     []DepthLevel
	Asks                     []DepthLevel
	MarketPriceResult        decimal.Decimal
	Err                      error
}

// AssetBook wraps one orderbook.OrderBook and serializes access through a goroutine.
type AssetBook struct {
	AssetID uuid.UUID
	Symbol  string
	book    *ob.OrderBook
	inbox   chan engineReq
}

func newAssetBook(assetID uuid.UUID, symbol string) *AssetBook {
	return &AssetBook{
		AssetID: assetID,
		Symbol:  symbol,
		book:    ob.NewOrderBook(),
		inbox:   make(chan engineReq, 256),
	}
}

func (ab *AssetBook) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case req := <-ab.inbox:
			ab.handle(req)
		}
	}
}

func (ab *AssetBook) handle(req engineReq) {
	var reply Reply
	switch req.kind {
	case reqLimit:
		done, partial, partialQty, err := ab.book.ProcessLimitOrder(req.side, req.orderID, req.qty, req.price)
		reply = Reply{Done: done, Partial: partial, PartialQuantityProcessed: partialQty, Err: err}
	case reqMarket:
		done, partial, partialQty, qtyLeft, err := ab.book.ProcessMarketOrder(req.side, req.qty)
		reply = Reply{Done: done, Partial: partial, PartialQuantityProcessed: partialQty, QuantityLeft: qtyLeft, Err: err}
	case reqCancel:
		cancelled := ab.book.CancelOrder(req.orderID)
		if cancelled == nil {
			reply.Err = fmt.Errorf("order %s not found in book", req.orderID)
		}
	case reqDepth:
		asks, bids := ab.book.Depth()
		reply.Asks = toDepthLevels(asks)
		reply.Bids = toDepthLevels(bids)
	case reqMarketPrice:
		price, _, err := ab.book.CalculateMarketPrice(req.side, req.qty)
		reply = Reply{MarketPriceResult: price, Err: err}
	}
	req.replyCh <- reply
}

func toDepthLevels(levels []*ob.PriceLevel) []DepthLevel {
	out := make([]DepthLevel, len(levels))
	for i, l := range levels {
		out[i] = DepthLevel{Price: l.Price, Quantity: l.Quantity}
	}
	return out
}

func (ab *AssetBook) send(req engineReq) Reply {
	req.replyCh = make(chan Reply, 1)
	ab.inbox <- req
	return <-req.replyCh
}

// PlaceLimit places a limit order in the engine.
func (ab *AssetBook) PlaceLimit(side ob.Side, orderID string, qty, price decimal.Decimal) Reply {
	return ab.send(engineReq{kind: reqLimit, side: side, orderID: orderID, qty: qty, price: price})
}

// PlaceMarket places a market order in the engine.
func (ab *AssetBook) PlaceMarket(side ob.Side, qty decimal.Decimal) Reply {
	return ab.send(engineReq{kind: reqMarket, side: side, qty: qty})
}

// Cancel removes an order from the engine.
func (ab *AssetBook) Cancel(orderID string) Reply {
	return ab.send(engineReq{kind: reqCancel, orderID: orderID})
}

// Depth returns the current order book depth snapshot.
func (ab *AssetBook) Depth() Reply {
	return ab.send(engineReq{kind: reqDepth})
}

// MarketPrice estimates the total cost to fill qty at market.
func (ab *AssetBook) MarketPrice(side ob.Side, qty decimal.Decimal) Reply {
	return ab.send(engineReq{kind: reqMarketPrice, side: side, qty: qty})
}

// RehydrateLimit injects a resting limit order during startup, before run() starts.
func (ab *AssetBook) RehydrateLimit(side ob.Side, orderID string, qty, price decimal.Decimal) error {
	_, _, _, err := ab.book.ProcessLimitOrder(side, orderID, qty, price)
	return err
}
