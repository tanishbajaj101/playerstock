import { useState } from 'react'
import { useLocation, useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Asset, PortfolioResponse } from '../api/types'
import { useAuthStore } from '../store/auth'
import styles from './Welcome.module.css'

function AssetCard({ asset }: { asset: Asset }) {
  const [imgFailed, setImgFailed] = useState(false)
  const isGeneric = !asset.player_img || asset.player_img.endsWith('icon512.png')
  const initials = asset.name.split(' ').slice(0, 2).map(w => w[0]).join('').toUpperCase()

  return (
    <div className={styles.card}>
      <div className={styles.photo}>
        {!imgFailed && !isGeneric ? (
          <img src={asset.player_img!} alt={asset.name} onError={() => setImgFailed(true)} />
        ) : (
          <div className={styles.initials}>{initials}</div>
        )}
      </div>
      <div className={styles.cardBody}>
        <div className={styles.assetName}>{asset.name}</div>
        {asset.team && <div className={styles.tag}>{asset.team}</div>}
        {asset.nationality && <div className={styles.tag}>{asset.nationality}</div>}
      </div>
    </div>
  )
}

export default function WelcomePage() {
  const location = useLocation()
  const navigate = useNavigate()
  const qc = useQueryClient()
  const { user } = useAuthStore()

  const passedAssets: Asset[] | undefined = (location.state as { assets?: Asset[] })?.assets

  const { data: portfolio } = useQuery<PortfolioResponse>({
    queryKey: ['portfolio'],
    queryFn: () => api.get<PortfolioResponse>('/api/portfolio'),
    enabled: !passedAssets,
  })

  const assets: Asset[] = passedAssets ?? portfolio?.positions.map(p => p.asset) ?? []

  const ackMutation = useMutation({
    mutationFn: () => api.post<{ status: string }>('/api/me/starter-pack', {}),
    onSuccess: async () => {
      await qc.invalidateQueries({ queryKey: ['me'] })
      navigate('/')
    },
  })

  return (
    <div className={styles.page}>
      <div className={styles.inner}>
        <h1 className={styles.heading}>Welcome, {user?.username ?? ''}!</h1>
        <p className={styles.sub}>
          Here's your starter portfolio — 10 assets, 1 unit each. Get trading!
        </p>

        <div className={styles.grid}>
          {assets.map(a => <AssetCard key={a.id} asset={a} />)}
        </div>

        <button
          className={`btn-primary ${styles.cta}`}
          onClick={() => ackMutation.mutate()}
          disabled={ackMutation.isPending}
        >
          {ackMutation.isPending ? 'Loading...' : 'Start Trading →'}
        </button>
      </div>
    </div>
  )
}
