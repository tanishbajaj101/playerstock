import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { api } from '../api/client'
import { useQueryClient } from '@tanstack/react-query'
import styles from './Onboarding.module.css'

export default function OnboardingPage() {
  const [username, setUsername] = useState('')
  const [error, setError] = useState('')
  const [loading, setLoading] = useState(false)
  const navigate = useNavigate()
  const qc = useQueryClient()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!username.trim()) return
    setLoading(true)
    setError('')
    try {
      const res = await api.post<{ username: string; starter_assets: unknown[] }>('/api/me/username', { username: username.trim() })
      await qc.invalidateQueries({ queryKey: ['me'] })
      navigate('/welcome', { state: { assets: res.starter_assets } })
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to set username')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className={styles.container}>
      <div className={styles.card}>
        <h2>Choose your username</h2>
        <p className="text-muted">This will be your unique trader identity on StakeStock.</p>
        <form onSubmit={handleSubmit} className={styles.form}>
          <input
            type="text"
            placeholder="cooltrader42"
            value={username}
            onChange={e => setUsername(e.target.value)}
            maxLength={30}
            pattern="[a-zA-Z0-9_]+"
            title="Letters, numbers and underscores only"
            required
          />
          {error && <p className="text-red">{error}</p>}
          <button type="submit" className="btn-primary" disabled={loading}>
            {loading ? 'Saving...' : 'Start Trading'}
          </button>
        </form>
      </div>
    </div>
  )
}
