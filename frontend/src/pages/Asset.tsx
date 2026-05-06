import { useState, useEffect, useCallback } from 'react'
import { useParams, Link } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { AssetWithPrice, DepthResponse, Trade, PortfolioResponse } from '../api/types'
import { useAuthStore } from '../store/auth'
import { useWebSocket } from '../ws/useAssetStream'
import OrderBookLadder from '../components/OrderBookLadder'
import TradesFeed from '../components/TradesFeed'
import OrderForm from '../components/OrderForm'
import PriceChart from '../components/PriceChart'
import styles from './Asset.module.css'

function PlayerAvatar({ src, initials }: { src?: string | null; initials: string }) {
  const [failed, setFailed] = useState(false)
  const isGeneric = !src || src.endsWith('icon512.png')
  if (failed || isGeneric) {
    return (
      <div className={styles.playerPortrait}>
        <div className={styles.playerAvatarFallback}>{initials}</div>
      </div>
    )
  }
  return (
    <div className={styles.playerPortrait}>
      <img
        className={styles.playerAvatar}
        src={src}
        alt={initials}
        onError={() => setFailed(true)}
      />
    </div>
  )
}

export default function AssetPage() {
  const { symbol } = useParams<{ symbol: string }>()
  const { user } = useAuthStore()
  const qc = useQueryClient()

  const { data: asset } = useQuery<AssetWithPrice>({
    queryKey: ['asset', symbol],
    queryFn: () => api.get<AssetWithPrice>(`/api/assets/${symbol}`),
  })

  const [depth, setDepth] = useState<DepthResponse>({ bids: [], asks: [] })
  const [liveTrades, setLiveTrades] = useState<Trade[]>([])
  const [mobileTab, setMobileTab] = useState<'chart' | 'book' | 'trade'>('chart')

  const { data: initialDepth } = useQuery<DepthResponse>({
    queryKey: ['depth', symbol],
    queryFn: () => api.get<DepthResponse>(`/api/assets/${symbol}/depth`),
  })

  const { data: initialTrades } = useQuery<Trade[]>({
    queryKey: ['trades', symbol],
    queryFn: () => api.get<Trade[]>(`/api/assets/${symbol}/trades`),
  })

  useEffect(() => {
    if (initialDepth) setDepth({ bids: initialDepth.bids ?? [], asks: initialDepth.asks ?? [] })
  }, [initialDepth])

  useEffect(() => {
    if (initialTrades) setLiveTrades(initialTrades)
  }, [initialTrades])

  const onWsMessage = useCallback((channel: string, data: unknown) => {
    if (channel === `asset:${symbol}:depth`) {
      const d = data as DepthResponse
      setDepth({ bids: d.bids ?? [], asks: d.asks ?? [] })
    } else if (channel === `asset:${symbol}:trades`) {
      const batch = data as Trade[]
      setLiveTrades(prev => [...batch, ...prev].slice(0, 100))
      qc.invalidateQueries({ queryKey: ['asset', symbol] })
    } else if (user && channel === `user:${user.id}:orders`) {
      qc.invalidateQueries({ queryKey: ['portfolio'] })
    }
  }, [symbol, user, qc])

  const { subscribe } = useWebSocket(onWsMessage)

  useEffect(() => {
    if (!symbol || !user) return
    subscribe([
      `asset:${symbol}:depth`,
      `asset:${symbol}:trades`,
      `user:${user.id}:orders`,
    ])
  }, [symbol, user, subscribe])

  const { data: portfolio } = useQuery<PortfolioResponse>({
    queryKey: ['portfolio'],
    queryFn: () => api.get<PortfolioResponse>('/api/portfolio'),
  })

  const myPosition = portfolio?.positions.find(p => p.asset.symbol === symbol) ?? null
  const myOpenOrders = (portfolio?.open_orders ?? []).filter(o => o.asset_id === asset?.id)

  const cancelMutation = useMutation({
    mutationFn: (id: string) => api.delete<void>(`/api/orders/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['portfolio'] })
    },
  })

  const lastPrice = liveTrades[0]?.price ?? asset?.last_price

  const initials = (asset?.name ?? symbol ?? '?')
    .split(' ').slice(0, 2).map((w: string) => w[0]).join('').toUpperCase()

  const formattedDob = asset?.date_of_birth
    ? new Date(asset.date_of_birth).toLocaleDateString('en-US', { year: 'numeric', month: 'short', day: 'numeric' })
    : null

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div className={styles.playerProfile}>
          <PlayerAvatar src={asset?.player_img} initials={initials} />
          <div className={styles.playerMeta}>
            <div className={styles.symbolPill}>{symbol}</div>
            <h2 className={styles.assetName}>{asset?.name ?? symbol}</h2>
            <div className={styles.playerSubline}>
              {asset?.nationality && <span className="text-muted">{asset.nationality}</span>}
              {asset?.role && <span className="text-muted">· {asset.role}</span>}
              {formattedDob && <span className="text-muted">· Born {formattedDob}</span>}
            </div>
            {(asset?.batting_style || asset?.bowling_style) && (
              <div className={styles.styleBadges}>
                {asset.batting_style && <span className={styles.styleBadge}>{asset.batting_style}</span>}
                {asset.bowling_style && <span className={styles.styleBadge}>{asset.bowling_style}</span>}
              </div>
            )}
            {asset?.team && (
              <div className={styles.teamRow}>
                {asset.team_logo && (
                  <img src={asset.team_logo} alt={asset.team} className={styles.teamLogo} />
                )}
                <span className={styles.teamName}>{asset.team}</span>
              </div>
            )}
          </div>
        </div>
        <div className={styles.marketPanel}>
          <span className={styles.marketLabel}>Last trade price</span>
          <div className={styles.priceDisplay}>
            {lastPrice
              ? <span className={styles.bigPrice}>{parseFloat(lastPrice).toFixed(2)}</span>
              : <span className={styles.bigPriceMuted}>--</span>
            }
            <span className={styles.coinLabel}>coins</span>
          </div>
          {myPosition ? (
            <Link to="/portfolio" className={styles.quickPosition}>
              <span>Your position</span>
              <strong className={parseFloat(myPosition.qty) >= 0 ? 'text-green' : 'text-red'}>
                {parseFloat(myPosition.qty).toFixed(0)} qty
              </strong>
            </Link>
          ) : (
            <span className={styles.quickPositionMuted}>No position yet</span>
          )}
        </div>
      </div>

      <div className={styles.mobileTabs}>
        {(['chart', 'book', 'trade'] as const).map(tab => (
          <button
            key={tab}
            className={`${styles.mobileTab} ${mobileTab === tab ? styles.mobileTabActive : ''}`}
            onClick={() => setMobileTab(tab)}
          >
            {tab === 'chart' ? 'Chart' : tab === 'book' ? 'Order Book' : 'Trade'}
          </button>
        ))}
      </div>

      <div className={styles.layout}>
        {/* Left column: chart + trades */}
        <div className={`${styles.colLeft} ${mobileTab !== 'chart' ? styles.mobileHide : ''}`}>
          <div className="card" style={{ marginBottom: 16 }}>
            <div className={styles.colTitle}>Price History</div>
            <PriceChart symbol={symbol!} />
          </div>

          <div className={`card ${styles.tradesSection}`}>
            <div className={styles.colTitle}>Recent Trades</div>
            <TradesFeed trades={liveTrades} myUserId={user?.id} />

            {myOpenOrders.length > 0 && (
              <>
                <div className={styles.colTitle} style={{ marginTop: 16 }}>My Open Orders</div>
                <table>
                  <thead>
                    <tr>
                      <th>Side</th>
                      <th>Type</th>
                      <th>Qty</th>
                      <th>Price</th>
                      <th>Filled</th>
                      <th>Status</th>
                      <th></th>
                    </tr>
                  </thead>
                  <tbody>
                    {myOpenOrders.map(o => (
                      <tr key={o.id}>
                        <td className={o.side === 1 ? 'text-green' : 'text-red'}>
                          {o.side === 1 ? 'Buy' : 'Sell'}{o.is_short ? ' (short)' : ''}
                        </td>
                        <td>{o.type}</td>
                        <td>{parseFloat(o.qty).toFixed(0)}</td>
                        <td>{o.price ? parseFloat(o.price).toFixed(2) : 'MKT'}</td>
                        <td>{parseFloat(o.filled_qty).toFixed(0)}</td>
                        <td className={o.status === 'partial' ? 'text-yellow' : 'text-muted'}>{o.status}</td>
                        <td>
                          <button
                            className="btn-ghost"
                            style={{ padding: '2px 8px', fontSize: 12 }}
                            onClick={() => cancelMutation.mutate(o.id)}
                          >Cancel</button>
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </>
            )}
          </div>
        </div>

        {/* Right column: orderbook + order form (sticky) */}
        <div className={`${styles.colRight} ${mobileTab === 'chart' ? styles.mobileHide : ''}`}>
          <div className={styles.stickyPanel}>
            <div className={`card ${mobileTab === 'trade' ? styles.mobileHide : ''}`} style={{ marginBottom: 16 }}>
              <div className={styles.colTitle}>Order Book</div>
              <OrderBookLadder bids={depth.bids} asks={depth.asks} />
            </div>

            <div className={`card ${mobileTab === 'book' ? styles.mobileHide : ''}`}>
              {myPosition && (
                <Link to="/portfolio" className={styles.positionBanner}>
                  <span>
                    Holding:{' '}
                    <strong className={parseFloat(myPosition.qty) >= 0 ? 'text-green' : 'text-red'}>
                      {parseFloat(myPosition.qty).toFixed(0)} / 5 qty
                    </strong>
                    {parseFloat(myPosition.locked_qty) > 0 && (
                      <span className="text-muted"> ({parseFloat(myPosition.locked_qty).toFixed(0)} locked)</span>
                    )}
                  </span>
                  {myPosition.unrealised_pnl && (
                    <span className={parseFloat(myPosition.unrealised_pnl) >= 0 ? 'text-green' : 'text-red'}>
                      PnL: {parseFloat(myPosition.unrealised_pnl) >= 0 ? '+' : ''}
                      {parseFloat(myPosition.unrealised_pnl).toFixed(2)} coins
                    </span>
                  )}
                </Link>
              )}

              <div className={styles.colTitle}>Place Order</div>
              <OrderForm symbol={symbol!} assetName={asset?.name} myPosition={myPosition} />
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
