import { useQuery } from '@tanstack/react-query'
import { api } from '../api/client'
import type { Trade } from '../api/types'
import { useAuthStore } from '../store/auth'
import styles from './History.module.css'

export default function HistoryPage() {
  const { user } = useAuthStore()
  const { data: trades, isLoading } = useQuery<Trade[]>({
    queryKey: ['history'],
    queryFn: () => api.get<Trade[]>('/api/history'),
  })

  return (
    <div className={styles.page}>
      <h2 className={styles.heading}>Trade History</h2>
      {isLoading ? (
        <p className="text-muted">Loading...</p>
      ) : !trades || trades.length === 0 ? (
        <p className="text-muted">No trades yet. Go make some!</p>
      ) : (
        <table>
          <thead>
            <tr>
              <th>Time</th>
              <th>Side</th>
              <th>Price</th>
              <th>Quantity</th>
              <th>Total</th>
            </tr>
          </thead>
          <tbody>
            {trades.map(t => {
              const isBuyer = t.buy_user_id === user?.id
              return (
                <tr key={t.id}>
                  <td className="text-muted">{new Date(t.created_at).toLocaleString()}</td>
                  <td className={isBuyer ? 'text-green' : 'text-red'}>
                    {isBuyer ? 'Buy' : 'Sell'}
                  </td>
                  <td>{parseFloat(t.price).toFixed(2)}</td>
                  <td>{parseFloat(t.qty).toFixed(4)}</td>
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
