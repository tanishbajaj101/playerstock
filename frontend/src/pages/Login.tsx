import styles from './Login.module.css'

const API_BASE = import.meta.env.VITE_API_URL ?? ''

const players = [
  {
    name: 'Virat Kohli',
    symbol: 'VIRAT_KOHLI',
    team: 'Royal Challengers Bengaluru',
    role: 'Batsman',
    price: '428.50',
    change: '+8.4%',
    sentiment: 'Bullish demand',
    image: 'https://documents.iplt20.com/ipl/IPLHeadshot2026/2.png',
    logo: '/logos/Royal Challengers Bangalore.webp',
    accent: 'red',
  },
  {
    name: 'MS Dhoni',
    symbol: 'MS_DHONI',
    team: 'Chennai Super Kings',
    role: 'WK-Batsman',
    price: '391.25',
    change: '+5.7%',
    sentiment: 'High liquidity',
    image: 'https://documents.iplt20.com/ipl/IPLHeadshot2026/57.png',
    logo: '/logos/Chennai Super Kings.webp',
    accent: 'yellow',
  },
] as const

const features = [
  {
    label: 'Buy rising players',
    copy: 'Back players before the market catches up and sell when their value moves higher.',
  },
  {
    label: 'Shortsell hype',
    copy: 'Take a view on falling prices by shortselling player stocks when the market looks stretched.',
  },
  {
    label: 'Track your book',
    copy: 'Watch holdings, orders, live prices, charts, and open positions from one trading desk.',
  },
] as const

function GoogleIcon() {
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" aria-hidden="true">
      <path d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92c-.26 1.37-1.04 2.53-2.21 3.31v2.77h3.57c2.08-1.92 3.28-4.74 3.28-8.09z" fill="#4285F4"/>
      <path d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84C3.99 20.53 7.7 23 12 23z" fill="#34A853"/>
      <path d="M5.84 14.09c-.22-.66-.35-1.36-.35-2.09s.13-1.43.35-2.09V7.07H2.18C1.43 8.55 1 10.22 1 12s.43 3.45 1.18 4.93l3.66-2.84z" fill="#FBBC05"/>
      <path d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15C17.45 2.09 14.97 1 12 1 7.7 1 3.99 3.47 2.18 7.07l3.66 2.84c.87-2.6 3.3-4.53 6.16-4.53z" fill="#EA4335"/>
    </svg>
  )
}

function PlayerCard({ player }: { player: typeof players[number] }) {
  return (
    <article className={`${styles.playerCard} ${styles[player.accent]}`}>
      <div className={styles.playerTopline}>
        <span>{player.symbol}</span>
        <span className={styles.change}>{player.change}</span>
      </div>
      <div className={styles.playerVisual}>
        <img src={player.image} alt={player.name} className={styles.playerImage} />
      </div>
      <div className={styles.playerInfo}>
        <div>
          <h3>{player.name}</h3>
          <p>{player.role}</p>
        </div>
        <img src={player.logo} alt={player.team} className={styles.teamLogo} />
      </div>
      <div className={styles.priceRow}>
        <span>Current value</span>
        <strong>{player.price}</strong>
      </div>
      <div className={styles.actionRow}>
        <span>{player.sentiment}</span>
        <div>
          <button type="button" className={styles.buyButton}>Buy</button>
          <button type="button" className={styles.shortButton}>Short</button>
        </div>
      </div>
    </article>
  )
}

export default function LoginPage() {
  return (
    <main className={styles.page}>
      <nav className={styles.nav} aria-label="Welcome">
        <div className={styles.brand}>Stake<span>Stock</span></div>
        <a href={`${API_BASE}/auth/google/login`} className={styles.navCta}>
          Sign in
        </a>
      </nav>

      <section className={styles.hero}>
        <div className={styles.heroCopy}>
          <div className={styles.eyebrow}>IPL player stock market</div>
          <h1>Trade your read on every player before the next big move.</h1>
          <p>
            Buy IPL player stocks when you expect their value to rise, sell into stronger prices,
            or shortsell when you think the market is too high. It is virtual trading with in-game coins.
          </p>
          <div className={styles.ctaGroup}>
            <a href={`${API_BASE}/auth/google/login`} className={styles.googleBtn}>
              <GoogleIcon />
              Continue with Google
            </a>
            <span className={styles.bonus}>Create an account and get 10 free players instantly</span>
          </div>
        </div>

        <div className={styles.marketPreview} aria-label="Featured player markets">
          {players.map(player => <PlayerCard key={player.symbol} player={player} />)}
        </div>
      </section>

      <section className={styles.marketStrip} aria-label="Market activity">
        <div>
          <span>Live player prices</span>
          <strong>24h moves</strong>
        </div>
        <div>
          <span>Portfolio</span>
          <strong>Holdings and P&amp;L</strong>
        </div>
        <div>
          <span>Trading</span>
          <strong>Buy, sell, short</strong>
        </div>
        <div>
          <span>Teams</span>
          <strong>All IPL squads</strong>
        </div>
      </section>

      <section className={styles.features} aria-label="Features">
        {features.map(feature => (
          <article key={feature.label} className={styles.feature}>
            <h2>{feature.label}</h2>
            <p>{feature.copy}</p>
          </article>
        ))}
      </section>
    </main>
  )
}
