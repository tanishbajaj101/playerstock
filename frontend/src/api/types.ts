export interface User {
  id: string
  email: string
  username: string | null
  created_at: string
}

export interface Balance {
  user_id: string
  cash: string
  cash_locked: string
}

export interface MeResponse {
  user: User
  balance: Balance
  needs_onboarding: boolean
  needs_welcome: boolean
}

export interface Asset {
  id: string
  symbol: string
  name: string
  description: string
  nationality?: string | null
  role?: string | null
  date_of_birth?: string | null
  batting_style?: string | null
  bowling_style?: string | null
  player_img?: string | null
  team: string
  team_logo?: string | null
}

export interface AssetWithPrice extends Asset {
  last_price: string | null
  price_24h_ago: string | null
  change_pct: string | null
  volume_24h: string
}

export interface PriceLevel {
  price: string
  quantity: string
}

export interface DepthResponse {
  bids: PriceLevel[]
  asks: PriceLevel[]
}

export type OrderSide = 0 | 1   // 0=sell 1=buy
export type OrderType = 'limit' | 'market'
export type OrderStatus = 'open' | 'partial' | 'filled' | 'cancelled' | 'rejected'

export interface Order {
  id: string
  user_id: string
  asset_id: string
  side: OrderSide
  type: OrderType
  qty: string
  filled_qty: string
  price: string | null
  status: OrderStatus
  is_short: boolean
  created_at: string
  updated_at: string
  asset?: Asset
}

export interface Trade {
  id: string
  asset_id: string
  buy_order_id: string
  sell_order_id: string
  buy_user_id: string
  sell_user_id: string
  qty: string
  price: string
  created_at: string
  asset?: Asset
}

export interface PlaceOrderRequest {
  asset_symbol: string
  side: OrderSide
  type: OrderType
  qty: string
  price?: string
}

export interface PortfolioPosition {
  user_id: string
  asset_id: string
  qty: string
  locked_qty: string
  asset: Asset
  last_price: string | null
  unrealised_pnl: string | null
}

export interface PortfolioResponse {
  balance: Balance
  positions: PortfolioPosition[]
  open_orders: Order[]
}

export interface PricePoint {
  ts: number    // Unix milliseconds
  price: number
}

export type ChartTimeframe = '24h' | '7d' | '30d'
