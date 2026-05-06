import { useState } from 'react'
import { useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Order, PlaceOrderRequest, PortfolioPosition } from '../api/types'
import styles from './OrderForm.module.css'

interface Props {
  symbol: string
  assetName?: string
  myPosition?: PortfolioPosition | null
}

export default function OrderForm({ symbol, assetName, myPosition }: Props) {
  const [side, setSide] = useState<0 | 1>(1)       // 1=buy, 0=sell
  const [type, setType] = useState<'limit' | 'market'>('limit')
  const [qty, setQty] = useState('')
  const [price, setPrice] = useState('')
  const [error, setError] = useState('')
  const [success, setSuccess] = useState('')
  const qc = useQueryClient()

  const mutation = useMutation({
    mutationFn: (req: PlaceOrderRequest) => api.post<Order>('/api/orders', req),
    onSuccess: () => {
      setError('')
      setSuccess('Order placed!')
      setQty('')
      setPrice('')
      qc.invalidateQueries({ queryKey: ['portfolio'] })
      qc.invalidateQueries({ queryKey: ['orders'] })
      setTimeout(() => setSuccess(''), 3000)
    },
    onError: (err: Error) => {
      setError(err.message)
    },
  })

  const maxUnits = 5
  const longQty = myPosition ? parseFloat(myPosition.qty) : 0
  const sellQty = parseFloat(qty) || 0
  const isShortWarning = side === 0 && type === 'limit' && sellQty > longQty
  const canBuy = Math.max(0, maxUnits - Math.max(longQty, 0))

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    setError('')
    setSuccess('')
    const req: PlaceOrderRequest = {
      asset_symbol: symbol,
      side,
      type,
      qty,
    }
    if (type === 'limit') {
      if (!price) { setError('Price required for limit orders'); return }
      req.price = price
    }
    mutation.mutate(req)
  }

  return (
    <form onSubmit={handleSubmit} className={styles.form}>
      {/* Buy / Sell toggle */}
      <div className={styles.tabs}>
        <button
          type="button"
          className={`${styles.tab} ${side === 1 ? styles.tabBuy : ''}`}
          onClick={() => setSide(1)}
        >Buy</button>
        <button
          type="button"
          className={`${styles.tab} ${side === 0 ? styles.tabSell : ''}`}
          onClick={() => setSide(0)}
        >Sell</button>
      </div>

      {/* Limit / Market */}
      <div className={styles.typeRow}>
        <button
          type="button"
          className={`${styles.typeBtn} ${type === 'limit' ? styles.typeBtnActive : ''}`}
          onClick={() => setType('limit')}
        >Limit</button>
        <button
          type="button"
          className={`${styles.typeBtn} ${type === 'market' ? styles.typeBtnActive : ''}`}
          onClick={() => setType('market')}
        >Market</button>
      </div>

      <div className={styles.field}>
        <label>Quantity</label>
        <input
          type="number"
          step="1"
          min="1"
          max={side === 1 ? canBuy : undefined}
          placeholder="0"
          value={qty}
          onChange={e => setQty(e.target.value)}
          required
        />
      </div>

      {type === 'limit' && (
        <div className={styles.field}>
          <label>Price</label>
          <input
            type="number"
            step="0.01"
            min="0.01"
            placeholder="0.00"
            value={price}
            onChange={e => setPrice(e.target.value)}
            required={type === 'limit'}
          />
        </div>
      )}

      {isShortWarning && (
        <div className={styles.shortWarn}>
          This will open a short position of ~{(sellQty - Math.max(longQty, 0)).toFixed(0)} units.
          Collateral: {((sellQty - Math.max(longQty, 0)) * (parseFloat(price) || 0)).toFixed(2)} coins reserved.
        </div>
      )}

      {side === 1 && (
        <div style={{ fontSize: 12, marginBottom: 4 }}>
          {canBuy === 0
            ? <span className="text-red">Position limit reached (max 5 units)</span>
            : <span className="text-muted">{canBuy} unit{canBuy !== 1 ? 's' : ''} remaining (max 5)</span>
          }
        </div>
      )}

      {error && <p className="text-red" style={{ fontSize: 13 }}>{error}</p>}
      {success && <p className="text-green" style={{ fontSize: 13 }}>{success}</p>}

      <button
        type="submit"
        className={side === 1 ? 'btn-green' : 'btn-red'}
        disabled={mutation.isPending || (side === 1 && canBuy === 0)}
        style={{ width: '100%', padding: '10px' }}
      >
        {mutation.isPending ? 'Placing...' : side === 1 ? `Buy ${assetName ?? symbol}` : `Sell ${assetName ?? symbol}`}
      </button>
    </form>
  )
}
