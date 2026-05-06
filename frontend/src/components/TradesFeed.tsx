import { useState, useEffect, useRef } from 'react'
import type { Trade } from '../api/types'
import styles from './TradesFeed.module.css'

interface Props {
  trades: Trade[]
  myUserId?: string
}

export default function TradesFeed({ trades, myUserId }: Props) {
  const [flashIds, setFlashIds] = useState<Set<string>>(new Set())
  const prevIdsRef = useRef<Set<string>>(new Set())

  useEffect(() => {
    const newIds = trades.map(t => t.id).filter(id => !prevIdsRef.current.has(id))
    prevIdsRef.current = new Set(trades.map(t => t.id))
    if (newIds.length === 0) return
    setFlashIds(new Set(newIds))
    const timer = setTimeout(() => setFlashIds(new Set()), 800)
    return () => clearTimeout(timer)
  }, [trades])

  return (
    <div className={styles.feed}>
      <div className={styles.header}>
        <span>Price</span>
        <span>Qty</span>
        <span>Time</span>
      </div>
      {trades.length === 0 ? (
        <p className={styles.empty}>No trades yet</p>
      ) : (
        trades.map(t => {
          const isBuyer = myUserId && t.buy_user_id === myUserId
          const isSeller = myUserId && t.sell_user_id === myUserId
          const isNew = flashIds.has(t.id)
          return (
            <div
              key={t.id}
              className={[
                styles.row,
                isBuyer ? styles.myBuy : isSeller ? styles.mySell : '',
                isNew ? styles.flash : '',
              ].filter(Boolean).join(' ')}
            >
              <span className="text-green">{parseFloat(t.price).toFixed(2)}</span>
              <span>{parseFloat(t.qty).toFixed(0)}</span>
              <span className="text-muted">{new Date(t.created_at).toLocaleTimeString()}</span>
            </div>
          )
        })
      )}
    </div>
  )
}
