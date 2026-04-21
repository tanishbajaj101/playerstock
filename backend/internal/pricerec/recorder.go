package pricerec

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Recorder samples the last-traded price for every asset every 30 minutes
// and writes it to price_snapshots. Run as a goroutine via Run(ctx).
type Recorder struct{ pool *pgxpool.Pool }

func New(pool *pgxpool.Pool) *Recorder { return &Recorder{pool: pool} }

func (r *Recorder) Run(ctx context.Context) {
	t := time.NewTicker(30 * time.Minute)
	defer t.Stop()
	log.Println("pricerec: started")
	for {
		select {
		case <-ctx.Done():
			log.Println("pricerec: stopped")
			return
		case tick := <-t.C:
			r.record(ctx, tick.UTC().Truncate(30*time.Minute))
		}
	}
}

func (r *Recorder) record(ctx context.Context, ts time.Time) {
	_, err := r.pool.Exec(ctx, `
		INSERT INTO price_snapshots (asset_id, ts, price)
		SELECT a.id, $1,
			(SELECT t.price FROM trades t
			 WHERE t.asset_id = a.id AND t.created_at <= $1
			 ORDER BY t.created_at DESC LIMIT 1)
		FROM assets a
		WHERE (SELECT t.price FROM trades t
		       WHERE t.asset_id = a.id AND t.created_at <= $1
		       ORDER BY t.created_at DESC LIMIT 1) IS NOT NULL
		ON CONFLICT (asset_id, ts) DO NOTHING
	`, ts)
	if err != nil {
		log.Printf("pricerec: record error: %v", err)
	}
}
