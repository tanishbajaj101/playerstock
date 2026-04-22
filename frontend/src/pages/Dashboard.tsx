import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { AssetWithPrice, PricePoint } from '../api/types'
import Sparkline from '../components/Sparkline'
import styles from './Dashboard.module.css'

type SortKey = 'name' | 'last_price' | 'change_pct' | 'volume_24h'
type ViewMode = 'grid' | 'table'

function CardAvatar({ src, name }: { src?: string | null; name: string }) {
  const [failed, setFailed] = useState(false)
  const isGeneric = !src || src.endsWith('icon512.png')
  const initial = name.trim()[0]?.toUpperCase() ?? '?'
  if (failed || isGeneric) {
    return <div className={styles.cardAvatarFallback}>{initial}</div>
  }
  return (
    <img
      src={src}
      alt=""
      className={styles.cardAvatar}
      onError={() => setFailed(true)}
    />
  )
}

function useWatchlist() {
  const [watchlist, setWatchlist] = useState<Set<string>>(() => {
    try {
      const stored = localStorage.getItem('ss_watchlist')
      return stored ? new Set(JSON.parse(stored)) : new Set()
    } catch { return new Set() }
  })

  const toggle = (id: string) => {
    setWatchlist(prev => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      localStorage.setItem('ss_watchlist', JSON.stringify([...next]))
      return next
    })
  }

  return { watchlist, toggle }
}

function changePctDisplay(pct: string | null) {
  if (pct == null) return null
  const n = parseFloat(pct)
  return (
    <span className={n >= 0 ? 'text-green' : 'text-red'} style={{ fontSize: 12, marginLeft: 6 }}>
      {n >= 0 ? '+' : ''}{n.toFixed(2)}%
    </span>
  )
}

export default function DashboardPage() {
  const { data: assets, isLoading } = useQuery<AssetWithPrice[]>({
    queryKey: ['assets'],
    queryFn: () => api.get<AssetWithPrice[]>('/api/assets'),
  })

  const { data: charts } = useQuery<Record<string, PricePoint[]>>({
    queryKey: ['charts', '24h'],
    queryFn: () => api.get<Record<string, PricePoint[]>>('/api/charts?tf=24h'),
    staleTime: 30 * 60 * 1000,
  })

  const [view, setView] = useState<ViewMode>('grid')
  const [sortKey, setSortKey] = useState<SortKey>('name')
  const [sortAsc, setSortAsc] = useState(true)
  const { watchlist, toggle } = useWatchlist()

  const trending = useMemo(() =>
    [...(assets ?? [])]
      .filter(a => a.change_pct != null)
      .sort((a, b) => Math.abs(parseFloat(b.change_pct!)) - Math.abs(parseFloat(a.change_pct!)))
      .slice(0, 8),
    [assets],
  )

  const sorted = useMemo(() => {
    const list = [...(assets ?? [])]
    list.sort((a, b) => {
      const aPin = watchlist.has(a.id) ? 0 : 1
      const bPin = watchlist.has(b.id) ? 0 : 1
      if (aPin !== bPin) return aPin - bPin

      let av: number, bv: number
      switch (sortKey) {
        case 'name':
          return sortAsc ? a.name.localeCompare(b.name) : b.name.localeCompare(a.name)
        case 'last_price':
          av = a.last_price ? parseFloat(a.last_price) : 0
          bv = b.last_price ? parseFloat(b.last_price) : 0
          break
        case 'change_pct':
          av = a.change_pct ? parseFloat(a.change_pct) : 0
          bv = b.change_pct ? parseFloat(b.change_pct) : 0
          break
        case 'volume_24h':
          av = parseFloat(a.volume_24h)
          bv = parseFloat(b.volume_24h)
          break
        default:
          return 0
      }
      return sortAsc ? av - bv : bv - av
    })
    return list
  }, [assets, sortKey, sortAsc, watchlist])

  const handleSort = (key: SortKey) => {
    if (sortKey === key) setSortAsc(!sortAsc)
    else { setSortKey(key); setSortAsc(true) }
  }

  const sortArrow = (key: SortKey) =>
    sortKey === key ? (sortAsc ? ' ▲' : ' ▼') : ''

  return (
    <div className={styles.page}>
      {trending.length > 0 && (
        <section className={styles.trendingSection}>
          <div className={styles.trendingLabel}>Trending</div>
          <div className={styles.trendingStrip}>
            {trending.map(asset => (
              <Link to={`/asset/${asset.symbol}`} key={asset.id} className={styles.trendingCard}>
                <div className={styles.trendingName}>{asset.name}</div>
                <div className={styles.trendingPrice}>
                  {asset.last_price != null
                    ? parseFloat(asset.last_price).toFixed(2)
                    : '--'}
                </div>
                <div>{changePctDisplay(asset.change_pct)}</div>
              </Link>
            ))}
          </div>
        </section>
      )}

      <div className={styles.marketsHeader}>
        <h2 className={styles.heading}>Markets</h2>
        <div className={styles.viewToggle}>
          <button
            className={`${styles.viewBtn} ${view === 'grid' ? styles.viewActive : ''}`}
            onClick={() => setView('grid')}
            title="Grid view"
          >▦</button>
          <button
            className={`${styles.viewBtn} ${view === 'table' ? styles.viewActive : ''}`}
            onClick={() => setView('table')}
            title="Table view"
          >☰</button>
        </div>
      </div>

      {isLoading ? (
        <p className="text-muted">Loading markets...</p>
      ) : view === 'grid' ? (
        <div className={styles.grid}>
          {sorted.map(asset => (
            <div key={asset.id} className={styles.card}>
              <button
                className={`${styles.star} ${watchlist.has(asset.id) ? styles.starred : ''}`}
                onClick={e => { e.preventDefault(); toggle(asset.id) }}
                title={watchlist.has(asset.id) ? 'Unpin' : 'Pin to watchlist'}
              >★</button>
              <Link to={`/asset/${asset.symbol}`} className={styles.cardLink}>
                <div className={styles.cardTop}>
                  <CardAvatar src={asset.player_img} name={asset.name} />
                  <span className={styles.assetName}>{asset.name}</span>
                  <span className={styles.price}>
                    {asset.last_price != null
                      ? <span className="text-green">{parseFloat(asset.last_price).toFixed(2)}</span>
                      : <span className="text-muted">--</span>
                    }
                    {changePctDisplay(asset.change_pct)}
                  </span>
                </div>
                <div className={styles.sparkline}>
                  <Sparkline points={charts?.[asset.symbol]} />
                </div>
                <div className={styles.description}>{asset.description}</div>
                <div className={styles.meta}>
                  <span className="text-muted">24h vol: </span>
                  {parseFloat(asset.volume_24h).toFixed(2)}
                </div>
              </Link>
            </div>
          ))}
        </div>
      ) : (
        <table className={styles.table}>
          <thead>
            <tr>
              <th style={{ width: 32 }}></th>
              <th className={styles.sortable} onClick={() => handleSort('name')}>
                Name{sortArrow('name')}
              </th>
              <th>24h</th>
              <th className={styles.sortable} onClick={() => handleSort('last_price')}>
                Last{sortArrow('last_price')}
              </th>
              <th className={styles.sortable} onClick={() => handleSort('change_pct')}>
                Change%{sortArrow('change_pct')}
              </th>
              <th className={styles.sortable} onClick={() => handleSort('volume_24h')}>
                Vol 24h{sortArrow('volume_24h')}
              </th>
            </tr>
          </thead>
          <tbody>
            {sorted.map(asset => (
              <tr key={asset.id} className={styles.tableRow}>
                <td>
                  <button
                    className={`${styles.starSmall} ${watchlist.has(asset.id) ? styles.starred : ''}`}
                    onClick={() => toggle(asset.id)}
                  >★</button>
                </td>
                <td>
                  <Link to={`/asset/${asset.symbol}`}>{asset.name}</Link>
                </td>
                <td style={{ width: 120 }}>
                  <Sparkline points={charts?.[asset.symbol]} />
                </td>
                <td>
                  {asset.last_price != null
                    ? parseFloat(asset.last_price).toFixed(2)
                    : '--'}
                </td>
                <td>{changePctDisplay(asset.change_pct)}</td>
                <td>{parseFloat(asset.volume_24h).toFixed(2)}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
