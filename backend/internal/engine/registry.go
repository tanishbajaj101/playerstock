package engine

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stakestock/backend/internal/models"
	ob "github.com/stakestock/backend/internal/orderbook"
)

type Registry struct {
	books map[string]*AssetBook // keyed by symbol
}

func NewRegistry() *Registry {
	return &Registry{books: make(map[string]*AssetBook)}
}

// Build loads all assets and rehydrates open limit orders, then starts each worker.
func (r *Registry) Build(ctx context.Context, pool *pgxpool.Pool) error {
	// Load assets
	rows, err := pool.Query(ctx, `SELECT id, symbol FROM assets ORDER BY symbol`)
	if err != nil {
		return fmt.Errorf("load assets: %w", err)
	}
	defer rows.Close()

	var assets []struct {
		id     uuid.UUID
		symbol string
	}
	for rows.Next() {
		var a struct {
			id     uuid.UUID
			symbol string
		}
		if err := rows.Scan(&a.id, &a.symbol); err != nil {
			return fmt.Errorf("scan asset: %w", err)
		}
		assets = append(assets, a)
	}
	rows.Close()

	for _, a := range assets {
		ab := newAssetBook(a.id, a.symbol)
		r.books[a.symbol] = ab
	}

	// Rehydrate open/partial limit orders
	if err := rehydrate(ctx, pool, r.books); err != nil {
		return fmt.Errorf("rehydrate: %w", err)
	}

	// Start worker goroutines
	for _, ab := range r.books {
		go ab.run(ctx)
	}

	log.Printf("engine: registry loaded %d assets", len(r.books))
	return nil
}

func (r *Registry) Get(symbol string) (*AssetBook, bool) {
	ab, ok := r.books[symbol]
	return ab, ok
}

func (r *Registry) All() map[string]*AssetBook {
	return r.books
}

func rehydrate(ctx context.Context, pool *pgxpool.Pool, books map[string]*AssetBook) error {
	rows, err := pool.Query(ctx, `
		SELECT o.id, o.asset_id, o.side, o.qty, o.filled_qty, o.price, a.symbol
		FROM orders o
		JOIN assets a ON a.id = o.asset_id
		WHERE o.status IN ('open','partial') AND o.type = 'limit' AND o.price IS NOT NULL
		ORDER BY o.created_at ASC
	`)
	if err != nil {
		return fmt.Errorf("query open orders: %w", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var (
			id        uuid.UUID
			assetID   uuid.UUID
			side      int16
			qty       decimal.Decimal
			filledQty decimal.Decimal
			price     decimal.Decimal
			symbol    string
		)
		if err := rows.Scan(&id, &assetID, &side, &qty, &filledQty, &price, &symbol); err != nil {
			return fmt.Errorf("scan open order: %w", err)
		}

		ab, ok := books[symbol]
		if !ok {
			continue
		}

		remaining := qty.Sub(filledQty)
		if remaining.LessThanOrEqual(decimal.Zero) {
			continue
		}

		obSide := ob.Side(side)
		if err := ab.RehydrateLimit(obSide, id.String(), remaining, price); err != nil {
			log.Printf("rehydrate order %s: %v (skipping)", id, err)
		} else {
			count++
		}
	}

	log.Printf("engine: rehydrated %d open limit orders", count)
	return nil
}

func SideFromModel(s models.OrderSide) ob.Side {
	if s == models.SideBuy {
		return ob.Buy
	}
	return ob.Sell
}
