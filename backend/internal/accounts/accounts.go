package accounts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stakestock/backend/internal/models"
)

// GetOrCreatePosition returns the position row, creating it with qty=0 if it doesn't exist.
func GetOrCreatePosition(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID) (*models.Position, error) {
	_, err := tx.Exec(ctx, `
		INSERT INTO positions (user_id, asset_id, qty, locked_qty)
		VALUES ($1, $2, 0, 0)
		ON CONFLICT (user_id, asset_id) DO NOTHING
	`, userID, assetID)
	if err != nil {
		return nil, fmt.Errorf("upsert position: %w", err)
	}

	var p models.Position
	err = tx.QueryRow(ctx, `
		SELECT user_id, asset_id, qty, locked_qty FROM positions WHERE user_id=$1 AND asset_id=$2
	`, userID, assetID).Scan(&p.UserID, &p.AssetID, &p.Qty, &p.LockedQty)
	if err != nil {
		return nil, fmt.Errorf("get position: %w", err)
	}
	return &p, nil
}

// ReserveCashForBuy locks cash for a buy order. Returns error if insufficient.
func ReserveCashForBuy(ctx context.Context, tx pgx.Tx, userID uuid.UUID, amount decimal.Decimal) error {
	tag, err := tx.Exec(ctx, `
		UPDATE balances
		SET cash_locked = cash_locked + $2
		WHERE user_id = $1 AND cash - cash_locked >= $2
	`, userID, amount)
	if err != nil {
		return fmt.Errorf("reserve cash: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient free cash balance")
	}
	return nil
}

// ReservePositionForSell locks existing long units for a covered sell.
func ReservePositionForSell(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, qty decimal.Decimal) error {
	tag, err := tx.Exec(ctx, `
		UPDATE positions
		SET locked_qty = locked_qty + $3
		WHERE user_id = $1 AND asset_id = $2 AND qty - locked_qty >= $3
	`, userID, assetID, qty)
	if err != nil {
		return fmt.Errorf("reserve position: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient available position")
	}
	return nil
}

// ReserveCashForShort locks cash collateral equal to qty*price for a short sell.
func ReserveCashForShort(ctx context.Context, tx pgx.Tx, userID uuid.UUID, collateral decimal.Decimal) error {
	tag, err := tx.Exec(ctx, `
		UPDATE balances
		SET cash_locked = cash_locked + $2
		WHERE user_id = $1 AND cash - cash_locked >= $2
	`, userID, collateral)
	if err != nil {
		return fmt.Errorf("reserve short collateral: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return fmt.Errorf("insufficient free cash for short collateral")
	}
	return nil
}

// ReleaseCashReservation releases previously locked cash (cancel or fill).
func ReleaseCashReservation(ctx context.Context, tx pgx.Tx, userID uuid.UUID, amount decimal.Decimal) error {
	_, err := tx.Exec(ctx, `
		UPDATE balances SET cash_locked = GREATEST(0, cash_locked - $2) WHERE user_id = $1
	`, userID, amount)
	return err
}

// ReleasePositionReservation releases previously locked position units.
func ReleasePositionReservation(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, qty decimal.Decimal) error {
	_, err := tx.Exec(ctx, `
		UPDATE positions SET locked_qty = GREATEST(0, locked_qty - $3) WHERE user_id=$1 AND asset_id=$2
	`, userID, assetID, qty)
	return err
}

// ApplyTradeBuy credits position and debits cash for the buyer.
func ApplyTradeBuy(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, qty, price, reservedPrice decimal.Decimal) error {
	// Increase position
	_, err := tx.Exec(ctx, `
		UPDATE positions SET qty = qty + $3 WHERE user_id=$1 AND asset_id=$2
	`, userID, assetID, qty)
	if err != nil {
		return fmt.Errorf("position buy: %w", err)
	}
	// Pay cash at trade price, release reserved at reserved price
	tradeCost := qty.Mul(price)
	reservedAmount := qty.Mul(reservedPrice)
	// cash -= tradeCost, cash_locked -= reservedAmount
	_, err = tx.Exec(ctx, `
		UPDATE balances
		SET cash = cash - $2, cash_locked = GREATEST(0, cash_locked - $3)
		WHERE user_id = $1
	`, userID, tradeCost, reservedAmount)
	return err
}

// ApplyTradeSellCovered credits cash and decrements position for a covered sell.
func ApplyTradeSellCovered(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, qty, price decimal.Decimal) error {
	proceeds := qty.Mul(price)
	_, err := tx.Exec(ctx, `
		UPDATE positions SET qty = qty - $3, locked_qty = GREATEST(0, locked_qty - $3)
		WHERE user_id=$1 AND asset_id=$2
	`, userID, assetID, qty)
	if err != nil {
		return fmt.Errorf("position covered sell: %w", err)
	}
	_, err = tx.Exec(ctx, `
		UPDATE balances SET cash = cash + $2 WHERE user_id=$1
	`, userID, proceeds)
	return err
}

// ApplyTradeSellShort credits cash and decrements position (goes negative) for a short sell.
// The collateral was already locked on order placement. On fill we:
//   - reduce position by qty (can go negative)
//   - credit proceeds to cash
//   - release the SHORT COLLATERAL from cash_locked (qty * reservedPrice)
func ApplyTradeSellShort(ctx context.Context, tx pgx.Tx, userID, assetID uuid.UUID, qty, price, reservedPrice decimal.Decimal) error {
	proceeds := qty.Mul(price)
	collateral := qty.Mul(reservedPrice)

	_, err := tx.Exec(ctx, `
		UPDATE positions SET qty = qty - $3 WHERE user_id=$1 AND asset_id=$2
	`, userID, assetID, qty)
	if err != nil {
		return fmt.Errorf("position short sell: %w", err)
	}
	_, err = tx.Exec(ctx, `
		UPDATE balances SET cash = cash + $2, cash_locked = GREATEST(0, cash_locked - $3)
		WHERE user_id=$1
	`, userID, proceeds, collateral)
	return err
}

// GetBalance returns the current balance for a user.
func GetBalance(ctx context.Context, db *pgxpool.Pool, userID uuid.UUID) (*models.Balance, error) {
	var b models.Balance
	err := db.QueryRow(ctx, `
		SELECT user_id, cash, cash_locked, special_coins FROM balances WHERE user_id=$1
	`, userID).Scan(&b.UserID, &b.Cash, &b.CashLocked, &b.SpecialCoins)
	if err != nil {
		return nil, fmt.Errorf("get balance: %w", err)
	}
	return &b, nil
}
