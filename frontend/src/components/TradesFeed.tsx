import type { Trade } from '../api/types'
import styles from './TradesFeed.module.css'

interface Props {
  trades: Trade[]
  myUserId?: string
}

export default function TradesFeed({ trades, myUserId }: Props) {
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
          return (
            <div
              key={t.id}
              className={`${styles.row} ${isBuyer ? styles.myBuy : isSeller ? styles.mySell : ''}`}
            >
              <span className="text-green">{parseFloat(t.price).toFixed(2)}</span>
              <span>{parseFloat(t.qty).toFixed(4)}</span>
              <span className="text-muted">{new Date(t.created_at).toLocaleTimeString()}</span>
            </div>
          )
        })
      )}
    </div>
  )
}
