import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { PortfolioResponse } from '../api/types'
import styles from './Portfolio.module.css'

export default function PortfolioPage() {
  const qc = useQueryClient()
  const { data, isLoading } = useQuery<PortfolioResponse>({
    queryKey: ['portfolio'],
    queryFn: () => api.get<PortfolioResponse>('/api/portfolio'),
  })

  const cancelMutation = useMutation({
    mutationFn: (id: string) => api.delete<void>(`/api/orders/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolio'] }),
  })

  if (isLoading) return <div className={styles.page}><p className="text-muted">Loading...</p></div>

  const { balance, positions, open_orders } = data!
  const freeCash = parseFloat(balance.cash) - parseFloat(balance.cash_locked)

  return (
    <div className={styles.page}>
      <h2 className={styles.heading}>Portfolio</h2>

      <div className={styles.balanceCard}>
        <div className={styles.balItem}>
          <span className="text-muted">Total Cash</span>
          <strong>{parseFloat(balance.cash).toFixed(2)}</strong>
        </div>
        <div className={styles.balItem}>
          <span className="text-muted">In Orders (Locked)</span>
          <strong className="text-red">{parseFloat(balance.cash_locked).toFixed(2)}</strong>
        </div>
        <div className={styles.balItem}>
          <span className="text-muted">Free Cash</span>
          <strong className="text-green">{freeCash.toFixed(2)}</strong>
        </div>
      </div>

      <h3 className={styles.subheading}>Positions</h3>
      {positions.length === 0 ? (
        <p className="text-muted">No open positions yet. <Link to="/">Go trade!</Link></p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Asset</th>
              <th>Quantity</th>
              <th>Locked</th>
              <th>Mark Price</th>
              <th>Unrealised P&L</th>
            </tr>
          </thead>
          <tbody>
            {positions.map(p => {
              const qty = parseFloat(p.qty)
              const pnl = p.unrealised_pnl ? parseFloat(p.unrealised_pnl) : null
              return (
                <tr key={p.asset_id}>
                  <td>
                    <Link to={`/asset/${p.asset.symbol}`}>{p.asset.name}</Link>
                  </td>
                  <td className={qty >= 0 ? 'text-green' : 'text-red'}>
                    {qty.toFixed(4)}
                    {qty < 0 && <span className="text-muted" style={{ fontSize: 11, marginLeft: 4 }}>(short)</span>}
                  </td>
                  <td className="text-muted">{parseFloat(p.locked_qty).toFixed(4)}</td>
                  <td>{p.last_price ? parseFloat(p.last_price).toFixed(2) : '--'}</td>
                  <td className={pnl === null ? '' : pnl >= 0 ? 'text-green' : 'text-red'}>
                    {pnl === null ? '--' : pnl.toFixed(2)}
                  </td>
                </tr>
              )
            })}
          </tbody>
        </table>
      )}

      <h3 className={styles.subheading}>Open Orders</h3>
      {open_orders.length === 0 ? (
        <p className="text-muted">No open orders.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Asset</th>
              <th>Side</th>
              <th>Type</th>
              <th>Qty</th>
              <th>Filled</th>
              <th>Price</th>
              <th>Status</th>
              <th></th>
            </tr>
          </thead>
          <tbody>
            {open_orders.map(o => (
              <tr key={o.id}>
                <td>
                  {o.asset
                    ? <Link to={`/asset/${o.asset.symbol}`}>{o.asset.name}</Link>
                    : o.asset_id.slice(0, 8)}
                </td>
                <td className={o.side === 1 ? 'text-green' : 'text-red'}>
                  {o.side === 1 ? 'Buy' : 'Sell'}{o.is_short ? ' (short)' : ''}
                </td>
                <td>{o.type}</td>
                <td>{parseFloat(o.qty).toFixed(4)}</td>
                <td>{parseFloat(o.filled_qty).toFixed(4)}</td>
                <td>{o.price ? parseFloat(o.price).toFixed(2) : 'MKT'}</td>
                <td><span className={`tag tag-${o.status}`}>{o.status}</span></td>
                <td>
                  <button
                    className="btn-ghost"
                    style={{ padding: '3px 10px', fontSize: 12 }}
                    onClick={() => cancelMutation.mutate(o.id)}
                    disabled={cancelMutation.isPending}
                  >Cancel</button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
