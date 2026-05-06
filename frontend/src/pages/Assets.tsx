import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { AssetWithPrice, PortfolioResponse } from '../api/types'
import styles from './Assets.module.css'

export default function AssetsPage() {
  const { data: assets, isLoading: assetsLoading } = useQuery<AssetWithPrice[]>({
    queryKey: ['assets'],
    queryFn: () => api.get<AssetWithPrice[]>('/api/assets'),
  })

  const { data: portfolio } = useQuery<PortfolioResponse>({
    queryKey: ['portfolio'],
    queryFn: () => api.get<PortfolioResponse>('/api/portfolio'),
  })

  if (assetsLoading) return <div className={styles.page}><p className="text-muted">Loading...</p></div>

  const positionMap = new Map(
    (portfolio?.positions ?? []).map(p => [p.asset_id, p])
  )

  return (
    <div className={styles.page}>
      <h2 className={styles.heading}>Assets</h2>
      <div className={styles.grid}>
        {(assets ?? []).map(asset => {
          const position = positionMap.get(asset.id)
          const qty = position ? parseFloat(position.qty) : 0
          const price = asset.last_price ? parseFloat(asset.last_price) : null
          const value = price !== null && qty !== 0 ? qty * price : null
          const change = asset.change_pct ? parseFloat(asset.change_pct) : null

          return (
            <Link to={`/asset/${asset.symbol}`} key={asset.id} className={styles.card}>
              <div className={styles.photoWrap}>
                {asset.player_img
                  ? <img src={asset.player_img} alt={asset.name} className={styles.photo} />
                  : <div className={styles.photoFallback}>{asset.name[0]}</div>
                }
              </div>
              <div className={styles.info}>
                <div className={styles.name}>{asset.name}</div>
                {asset.team && <div className={styles.team}>{asset.team}</div>}
                {asset.nationality && <div className={styles.meta}>{asset.nationality}</div>}
              </div>
              <div className={styles.priceRow}>
                <span className={styles.price}>
                  {price !== null ? price.toFixed(2) : '--'}
                </span>
                {change !== null && (
                  <span className={change >= 0 ? 'text-green' : 'text-red'}>
                    {change >= 0 ? '+' : ''}{change.toFixed(2)}%
                  </span>
                )}
              </div>
              {qty !== 0 && (
                <div className={styles.holdings}>
                  <span className="text-muted">Held:</span>{' '}
                  <strong>{qty.toFixed(0)}</strong>
                  {value !== null && (
                    <span className="text-muted"> · {value.toFixed(2)}</span>
                  )}
                </div>
              )}
            </Link>
          )
        })}
      </div>
    </div>
  )
}
