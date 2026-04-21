import { Link, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import { api } from '../api/client'
import styles from './NavBar.module.css'

export default function NavBar() {
  const { user, balance, clearAuth } = useAuthStore()
  const navigate = useNavigate()

  const handleLogout = async () => {
    await api.post('/auth/logout', {}).catch(() => {})
    clearAuth()
    navigate('/login')
  }

  const freeCash = balance
    ? (parseFloat(balance.cash) - parseFloat(balance.cash_locked)).toFixed(2)
    : '—'

  return (
    <nav className={styles.nav}>
      <div className={styles.left}>
        <Link to="/" className={styles.brand}>StakeStock</Link>
        <Link to="/" className={styles.link}>Markets</Link>
        <Link to="/portfolio" className={styles.link}>Portfolio</Link>
        <Link to="/history" className={styles.link}>History</Link>
      </div>
      <div className={styles.right}>
        <span className={styles.balance}>
          <span className="text-muted">Free:</span> <strong>{freeCash}</strong> coins
        </span>
        <span className={styles.username}>{user?.username}</span>
        <button className="btn-ghost" onClick={handleLogout}>Logout</button>
      </div>
    </nav>
  )
}
