package models

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type User struct {
	ID        uuid.UUID `json:"id"`
	GoogleSub string    `json:"-"`
	Email     string    `json:"email"`
	Username  *string   `json:"username"`
	CreatedAt time.Time `json:"created_at"`
}

type Balance struct {
	UserID       uuid.UUID       `json:"user_id"`
	Cash         decimal.Decimal `json:"cash"`
	CashLocked   decimal.Decimal `json:"cash_locked"`
	SpecialCoins int             `json:"special_coins"`
}

type Asset struct {
	ID          uuid.UUID `json:"id"`
	Symbol      string    `json:"symbol"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Nationality *string   `json:"nationality"`
	Role        *string   `json:"role"`
}

type Position struct {
	UserID    uuid.UUID       `json:"user_id"`
	AssetID   uuid.UUID       `json:"asset_id"`
	Qty       decimal.Decimal `json:"qty"`
	LockedQty decimal.Decimal `json:"locked_qty"`
}

type OrderSide int16

const (
	SideSell OrderSide = 0
	SideBuy  OrderSide = 1
)

type OrderType string

const (
	OrderTypeLimit  OrderType = "limit"
	OrderTypeMarket OrderType = "market"
)

type OrderStatus string

const (
	OrderStatusOpen      OrderStatus = "open"
	OrderStatusPartial   OrderStatus = "partial"
	OrderStatusFilled    OrderStatus = "filled"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusRejected  OrderStatus = "rejected"
)

type Order struct {
	ID        uuid.UUID        `json:"id"`
	UserID    uuid.UUID        `json:"user_id"`
	AssetID   uuid.UUID        `json:"asset_id"`
	Side      OrderSide        `json:"side"`
	Type      OrderType        `json:"type"`
	Qty       decimal.Decimal  `json:"qty"`
	FilledQty decimal.Decimal  `json:"filled_qty"`
	Price     *decimal.Decimal `json:"price"`
	Status    OrderStatus      `json:"status"`
	IsShort   bool             `json:"is_short"`
	CreatedAt time.Time        `json:"created_at"`
	UpdatedAt time.Time        `json:"updated_at"`
	Asset     *Asset           `json:"asset,omitempty"`
}

type Trade struct {
	ID           uuid.UUID       `json:"id"`
	AssetID      uuid.UUID       `json:"asset_id"`
	BuyOrderID   uuid.UUID       `json:"buy_order_id"`
	SellOrderID  uuid.UUID       `json:"sell_order_id"`
	BuyUserID    uuid.UUID       `json:"buy_user_id"`
	SellUserID   uuid.UUID       `json:"sell_user_id"`
	Qty          decimal.Decimal `json:"qty"`
	Price        decimal.Decimal `json:"price"`
	CreatedAt    time.Time       `json:"created_at"`
	Asset        *Asset          `json:"asset,omitempty"`
}

// AssetWithPrice is returned by the assets list and detail endpoints
type AssetWithPrice struct {
	Asset
	LastPrice       *decimal.Decimal `json:"last_price"`
	Price24hAgo     *decimal.Decimal `json:"price_24h_ago"`
	ChangePct       *decimal.Decimal `json:"change_pct"`
	Volume24h       decimal.Decimal  `json:"volume_24h"`
	SupplyUsed      int              `json:"supply_used"`
	SpecialCoinUsed bool             `json:"special_coin_used"`
}

// PortfolioPosition includes mark-to-market PnL
type PortfolioPosition struct {
	Position
	Asset     Asset            `json:"asset"`
	LastPrice *decimal.Decimal `json:"last_price"`
	UnrealPnL *decimal.Decimal `json:"unrealised_pnl"`
}

// Events published to Redis
type TradeEvent struct {
	AssetSymbol string          `json:"asset_symbol"`
	Trade       Trade           `json:"trade"`
}

type DepthEvent struct {
	AssetSymbol string           `json:"asset_symbol"`
	Bids        []PriceLevel     `json:"bids"`
	Asks        []PriceLevel     `json:"asks"`
}

type PriceLevel struct {
	Price    decimal.Decimal `json:"price"`
	Quantity decimal.Decimal `json:"quantity"`
}

type OrderUpdateEvent struct {
	UserID string `json:"user_id"`
	Order  Order  `json:"order"`
}

// PricePoint is the wire format for chart data (float64 for JS charting libs).
type PricePoint struct {
	Ts    int64   `json:"ts"`    // Unix milliseconds
	Price float64 `json:"price"`
}
