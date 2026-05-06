import { Link, useLocation, useNavigate } from 'react-router-dom'
import { useAuthStore } from '../store/auth'
import { api } from '../api/client'
import styles from './NavBar.module.css'

function IconMarkets() {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" width="20" height="20">
      <path d="M2 11a1 1 0 011-1h2a1 1 0 011 1v5a1 1 0 01-1 1H3a1 1 0 01-1-1v-5zM8 7a1 1 0 011-1h2a1 1 0 011 1v9a1 1 0 01-1 1H9a1 1 0 01-1-1V7zM14 4a1 1 0 011-1h2a1 1 0 011 1v12a1 1 0 01-1 1h-2a1 1 0 01-1-1V4z" />
    </svg>
  )
}

function IconAssets() {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" width="20" height="20">
      <path d="M5 3a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2V5a2 2 0 00-2-2H5zM5 11a2 2 0 00-2 2v2a2 2 0 002 2h2a2 2 0 002-2v-2a2 2 0 00-2-2H5zM11 5a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V5zM11 13a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
    </svg>
  )
}

function IconOrders() {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" width="20" height="20">
      <path fillRule="evenodd" d="M3 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1zm0 4a1 1 0 011-1h12a1 1 0 110 2H4a1 1 0 01-1-1z" clipRule="evenodd" />
    </svg>
  )
}

function IconWallet() {
  return (
    <svg viewBox="0 0 20 20" fill="currentColor" width="20" height="20">
      <path d="M4 4a2 2 0 00-2 2v1h16V6a2 2 0 00-2-2H4z" />
      <path fillRule="evenodd" d="M18 9H2v5a2 2 0 002 2h12a2 2 0 002-2V9zM4 13a1 1 0 011-1h1a1 1 0 110 2H5a1 1 0 01-1-1zm5-1a1 1 0 100 2h1a1 1 0 100-2H9z" clipRule="evenodd" />
    </svg>
  )
}

const NAV_LINKS = [
  { to: '/', label: 'Markets', Icon: IconMarkets },
  { to: '/assets', label: 'Assets', Icon: IconAssets },
  { to: '/orders', label: 'Orders', Icon: IconOrders },
] as const

export default function NavBar() {
  const { user, balance, clearAuth } = useAuthStore()
  const navigate = useNavigate()
  const location = useLocation()

  const handleLogout = async () => {
    await api.post('/auth/logout', {}).catch(() => {})
    clearAuth()
    navigate('/login')
  }

  const freeCash = balance
    ? (parseFloat(balance.cash) - parseFloat(balance.cash_locked)).toFixed(2)
    : '—'

  const isActive = (to: string) =>
    to === '/' ? location.pathname === '/' : location.pathname.startsWith(to)

  return (
    <>
      <nav className={styles.nav}>
        <div className={styles.left}>
          <Link to="/" className={styles.brand}>
            Stake<span>Stock</span>
          </Link>
          {NAV_LINKS.map(({ to, label }) => (
            <Link
              key={to}
              to={to}
              className={`${styles.link} ${isActive(to) ? styles.linkActive : ''}`}
            >
              {label}
            </Link>
          ))}
        </div>
        <div className={styles.right}>
          <div className={styles.balanceChip}>
            <span className={styles.balanceValue}>{freeCash}</span>
            <span className={styles.balanceLabel}>coins</span>
          </div>
          <span className={styles.username}>{user?.username}</span>
          <button className="btn-ghost" onClick={handleLogout} style={{ fontSize: 13, padding: '4px 12px' }}>
            Logout
          </button>
        </div>
      </nav>

      <nav className={styles.bottomNav}>
        {NAV_LINKS.map(({ to, label, Icon }) => (
          <Link
            key={to}
            to={to}
            className={`${styles.tab} ${isActive(to) ? styles.tabActive : ''}`}
          >
            <span className={styles.tabIcon}><Icon /></span>
            <span className={styles.tabLabel}>{label}</span>
          </Link>
        ))}
        <div className={styles.tab}>
          <span className={styles.tabIcon}><IconWallet /></span>
          <span className={styles.tabLabel}>{freeCash}c</span>
        </div>
      </nav>
    </>
  )
}
