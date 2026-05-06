import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { useMemo, useState } from 'react'
import { api } from '../api/client'
import type { PortfolioPosition, PortfolioResponse } from '../api/types'
import styles from './Holdings.module.css'

function Avatar({ src, name }: { src?: string | null; name: string }) {
  const [failed, setFailed] = useState(false)
  const isGeneric = !src || src.endsWith('icon512.png')
  if (failed || isGeneric) {
    return <div className={styles.avatarFallback}>{name.trim()[0]?.toUpperCase() ?? '?'}</div>
  }
  return <img src={src} alt="" className={styles.avatar} onError={() => setFailed(true)} />
}

function formatCoins(value: number | null) {
  if (value == null || Number.isNaN(value)) return '--'
  return value.toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  })
}

function positionValue(position: PortfolioPosition) {
  const qty = parseFloat(position.qty)
  const price = position.last_price != null ? parseFloat(position.last_price) : null
  return price != null ? qty * price : null
}

export default function HoldingsPage() {
  const { data, isLoading } = useQuery<PortfolioResponse>({
    queryKey: ['portfolio'],
    queryFn: () => api.get<PortfolioResponse>('/api/portfolio'),
    refetchInterval: 10_000,
  })

  const positions = useMemo(() => {
    return [...(data?.positions ?? [])]
      .filter(p => parseFloat(p.qty) > 0)
      .sort((a, b) => (positionValue(b) ?? 0) - (positionValue(a) ?? 0))
  }, [data?.positions])

  const summary = useMemo(() => {
    const totalValue = positions.reduce((sum, p) => sum + (positionValue(p) ?? 0), 0)
    const totalPnl = positions.reduce((sum, p) => {
      const pnl = p.unrealised_pnl != null ? parseFloat(p.unrealised_pnl) : 0
      return sum + pnl
    }, 0)

    return { totalValue, totalPnl }
  }, [positions])

  return (
    <div className={styles.page}>
      <div className={styles.header}>
        <div>
          <p className={styles.eyebrow}>Portfolio</p>
          <h2 className={styles.heading}>Holdings</h2>
        </div>
        <Link to="/assets" className={styles.browseLink}>Browse assets</Link>
      </div>

      {isLoading ? (
        <div className={styles.loadingGrid} aria-label="Loading holdings">
          <div className={styles.loadingCard} />
          <div className={styles.loadingCard} />
          <div className={styles.loadingCard} />
        </div>
      ) : positions.length === 0 ? (
        <div className={styles.empty}>
          <div className={styles.emptyIcon}>0</div>
          <h3 className={styles.emptyTitle}>No holdings yet</h3>
          <p className={styles.emptyText}>Start by buying a player asset. Your quantities, live value, and unrealised P/L will show here.</p>
          <Link to="/assets" className={styles.primaryAction}>Browse assets</Link>
        </div>
      ) : (
        <>
          <section className={styles.summaryGrid} aria-label="Portfolio summary">
            <div className={styles.metricCard}>
              <span className={styles.metricLabel}>Portfolio value</span>
              <strong className={styles.metricValue}>{formatCoins(summary.totalValue)}</strong>
            </div>
            <div className={styles.metricCard}>
              <span className={styles.metricLabel}>Unrealised P/L</span>
              <strong className={summary.totalPnl >= 0 ? styles.metricPositive : styles.metricNegative}>
                {summary.totalPnl >= 0 ? '+' : ''}{formatCoins(summary.totalPnl)}
              </strong>
            </div>
          </section>

          <section className={styles.holdingsPanel}>
            <div className={styles.tableHeader}>
              <span>Asset</span>
              <span>Qty</span>
              <span>Last price</span>
              <span>Value</span>
              <span>P/L</span>
            </div>
            <div className={styles.list}>
              {positions.map(p => {
                const qty = parseFloat(p.qty)
                const locked = parseFloat(p.locked_qty)
                const price = p.last_price != null ? parseFloat(p.last_price) : null
                const value = positionValue(p)
                const pnl = p.unrealised_pnl != null ? parseFloat(p.unrealised_pnl) : null

                return (
                  <Link to={`/asset/${p.asset.symbol}`} key={p.asset_id} className={styles.row}>
                    <div className={styles.assetCell}>
                      <Avatar src={p.asset.player_img} name={p.asset.name} />
                      <div className={styles.info}>
                        <span className={styles.name}>{p.asset.name}</span>
                        <span className={styles.team}>
                          {p.asset.team_logo && <img src={p.asset.team_logo} alt="" className={styles.teamLogo} />}
                          {p.asset.team}
                        </span>
                      </div>
                    </div>
                    <div className={styles.qty} data-label="Qty">
                      <span className={styles.qtyValue}>{formatCoins(qty).replace(/\.00$/, '')}</span>
                      {locked > 0 && <span className={styles.locked}>{formatCoins(locked).replace(/\.00$/, '')} locked</span>}
                    </div>
                    <div className={styles.priceCol} data-label="Last price">
                      <span className={styles.price}>{formatCoins(price)}</span>
                      <span className={styles.subValue}>per unit</span>
                    </div>
                    <div className={styles.valueCol} data-label="Value">
                      <span className={styles.value}>{formatCoins(value)}</span>
                    </div>
                    <div className={styles.pnlCol} data-label="P/L">
                      {pnl != null ? (
                        <span className={pnl >= 0 ? styles.pnlPositive : styles.pnlNegative}>
                          {pnl >= 0 ? '+' : ''}{formatCoins(pnl)}
                        </span>
                      ) : (
                        <span className={styles.missingValue}>--</span>
                      )}
                    </div>
                  </Link>
                )
              })}
            </div>
          </section>
        </>
      )}
    </div>
  )
}
