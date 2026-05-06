import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Link } from 'react-router-dom'
import { api } from '../api/client'
import type { PortfolioResponse, Trade } from '../api/types'
import { useAuthStore } from '../store/auth'
import styles from './Orders.module.css'

export default function OrdersPage() {
  const qc = useQueryClient()
  const { user } = useAuthStore()

  const { data: portfolio, isLoading: portfolioLoading } = useQuery<PortfolioResponse>({
    queryKey: ['portfolio'],
    queryFn: () => api.get<PortfolioResponse>('/api/portfolio'),
  })

  const { data: trades, isLoading: tradesLoading } = useQuery<Trade[]>({
    queryKey: ['history'],
    queryFn: () => api.get<Trade[]>('/api/history'),
  })

  const cancelMutation = useMutation({
    mutationFn: (id: string) => api.delete<void>(`/api/orders/${id}`),
    onSuccess: () => qc.invalidateQueries({ queryKey: ['portfolio'] }),
  })

  const isLoading = portfolioLoading || tradesLoading

  if (isLoading) return <div className={styles.page}><p className="text-muted">Loading...</p></div>

  const { balance, open_orders } = portfolio!
  const freeCash = parseFloat(balance.cash) - parseFloat(balance.cash_locked)

  return (
    <div className={styles.page}>
      <h2 className={styles.heading}>Orders</h2>

      <div className={styles.balanceCard}>
        <div className={styles.balItem}>
          <span className="text-muted">Total Cash</span>
          <strong>{parseFloat(balance.cash).toFixed(2)}</strong>
        </div>
        <div className={styles.balItem}>
          <span className="text-muted">In Orders</span>
          <strong className="text-red">{parseFloat(balance.cash_locked).toFixed(2)}</strong>
        </div>
        <div className={styles.balItem}>
          <span className="text-muted">Free Cash</span>
          <strong className="text-green">{freeCash.toFixed(2)}</strong>
        </div>
      </div>

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
                <td>{parseFloat(o.qty).toFixed(0)}</td>
                <td>{parseFloat(o.filled_qty).toFixed(0)}</td>
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

      <h3 className={styles.subheading}>Trade History</h3>
      {!trades || trades.length === 0 ? (
        <p className="text-muted">No trades yet.</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Asset</th>
              <th>Side</th>
              <th>Price</th>
              <th>Qty</th>
              <th>Total</th>
            </tr>
          </thead>
          <tbody>
            {trades.map(t => {
              const isBuyer = t.buy_user_id === user?.id
              return (
                <tr key={t.id}>
                  <td className="text-muted">{new Date(t.created_at).toLocaleString()}</td>
                  <td>
                    {t.asset
                      ? <Link to={`/asset/${t.asset.symbol}`}>{t.asset.name}</Link>
                      : t.asset_id}
                  </td>
                  <td className={isBuyer ? 'text-green' : 'text-red'}>
                    {isBuyer ? 'Buy' : 'Sell'}
                  </td>
                  <td>{parseFloat(t.price).toFixed(2)}</td>
                  <td>{parseFloat(t.qty).toFixed(0)}</td>
                  <td>{(parseFloat(t.price) * parseFloat(t.qty)).toFixed(2)}</td>
                </tr>
              )
            })}
          </tbody>
        </table>
      )}
    </div>
  )
}
