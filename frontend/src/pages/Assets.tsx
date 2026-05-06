import { useState, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import * as Flags from 'country-flag-icons/react/3x2'
import { api } from '../api/client'
import type { AssetWithPrice, PortfolioResponse } from '../api/types'
import styles from './Assets.module.css'

const NATIONALITY_TO_CODE: Record<string, string> = {
  'India': 'IN',
  'Australia': 'AU',
  'England': 'GB',
  'South Africa': 'ZA',
  'New Zealand': 'NZ',
  'Sri Lanka': 'LK',
  'Bangladesh': 'BD',
  'Afghanistan': 'AF',
  'Pakistan': 'PK',
  'Zimbabwe': 'ZW',
  'Ireland': 'IE',
  'West Indies': 'TT',
  'Netherlands': 'NL',
  'Scotland': 'GB',
}

export default function AssetsPage() {
  const [selectedTeam, setSelectedTeam] = useState<string | null>(null)

  const { data: assets, isLoading: assetsLoading } = useQuery<AssetWithPrice[]>({
    queryKey: ['assets'],
    queryFn: () => api.get<AssetWithPrice[]>('/api/assets'),
  })

  const { data: portfolio } = useQuery<PortfolioResponse>({
    queryKey: ['portfolio'],
    queryFn: () => api.get<PortfolioResponse>('/api/portfolio'),
  })

  const teams = useMemo(() => {
    if (!assets) return []
    const seen = new Set<string>()
    return assets.filter(a => {
      if (!a.team || !a.team_logo || seen.has(a.team)) return false
      seen.add(a.team)
      return true
    })
  }, [assets])

  const filtered = useMemo(() =>
    selectedTeam ? (assets ?? []).filter(a => a.team === selectedTeam) : (assets ?? []),
    [assets, selectedTeam]
  )

  if (assetsLoading) return <div className={styles.page}><p className="text-muted">Loading...</p></div>

  const positionMap = new Map(
    (portfolio?.positions ?? []).map(p => [p.asset_id, p])
  )

  return (
    <div className={styles.page}>
      <div className={styles.teamFilter}>
        {teams.map(t => (
          <button
            key={t.team}
            className={`${styles.teamFilterBtn} ${selectedTeam === t.team ? styles.teamFilterBtnActive : ''}`}
            onClick={() => setSelectedTeam(prev => prev === t.team ? null : t.team)}
          >
            <img src={t.team_logo!} alt={t.team} className={styles.teamFilterLogo} title={t.team} />
          </button>
        ))}
      </div>

      <div className={styles.grid}>
        {filtered.map(asset => {
          const position = positionMap.get(asset.id)
          const qty = position ? parseFloat(position.qty) : 0
          const price = asset.last_price ? parseFloat(asset.last_price) : null
          const value = price !== null && qty !== 0 ? qty * price : null
          const change = asset.change_pct ? parseFloat(asset.change_pct) : null
          const countryCode = asset.nationality ? NATIONALITY_TO_CODE[asset.nationality] : null
          const FlagComp = countryCode ? Flags[countryCode as keyof typeof Flags] : null

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
                <div className={styles.cardIcons}>
                  {asset.team_logo && (
                    <img src={asset.team_logo} alt={asset.team} className={styles.smallTeamLogo} title={asset.team} />
                  )}
                  {FlagComp && (
                    <FlagComp className={styles.flagIcon} title={asset.nationality ?? ''} />
                  )}
                </div>
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
