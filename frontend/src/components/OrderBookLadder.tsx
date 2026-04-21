import type { PriceLevel } from '../api/types'
import styles from './OrderBookLadder.module.css'

interface Props {
  bids: PriceLevel[]
  asks: PriceLevel[]
}

function fmt(n: string) { return parseFloat(n).toFixed(2) }

export default function OrderBookLadder({ bids, asks }: Props) {
  const maxQty = Math.max(
    ...bids.map(l => parseFloat(l.quantity)),
    ...asks.map(l => parseFloat(l.quantity)),
    0.001
  )

  return (
    <div className={styles.book}>
      <div className={styles.header}>
        <span>Price</span>
        <span>Quantity</span>
      </div>

      {/* Asks (sorted ascending, displayed in reverse so best ask is closest to spread) */}
      <div className={styles.asks}>
        {[...asks].reverse().map((l, i) => (
          <div key={i} className={styles.row}>
            <div
              className={styles.fill}
              style={{ width: `${(parseFloat(l.quantity) / maxQty) * 100}%`, background: 'rgba(248,81,73,0.15)' }}
            />
            <span className="text-red">{fmt(l.price)}</span>
            <span>{fmt(l.quantity)}</span>
          </div>
        ))}
      </div>

      <div className={styles.spread}>
        {asks.length === 0 && bids.length === 0
          ? <span className="text-muted">No orders yet</span>
          : asks.length > 0 && bids.length > 0
            ? <span className="text-muted">
                Spread: {(parseFloat(asks[0].price) - parseFloat(bids[0].price)).toFixed(2)}
              </span>
            : null
        }
      </div>

      {/* Bids (sorted descending) */}
      <div className={styles.bids}>
        {bids.map((l, i) => (
          <div key={i} className={styles.row}>
            <div
              className={styles.fill}
              style={{ width: `${(parseFloat(l.quantity) / maxQty) * 100}%`, background: 'rgba(63,185,80,0.15)' }}
            />
            <span className="text-green">{fmt(l.price)}</span>
            <span>{fmt(l.quantity)}</span>
          </div>
        ))}
      </div>
    </div>
  )
}
