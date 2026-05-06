import { useEffect, useState } from 'react'

type Status = 'checking' | 'ready' | 'unavailable'

const POLL_INTERVAL = 2000
const MAX_ATTEMPTS = 15

export default function BackendGate({ children }: { children: React.ReactNode }) {
  const [status, setStatus] = useState<Status>('checking')
  const [attempts, setAttempts] = useState(0)

  useEffect(() => {
    let cancelled = false
    let timer: ReturnType<typeof setTimeout>

    async function check() {
      try {
        const res = await fetch('/healthz', { cache: 'no-store' })
        if (!cancelled && res.ok) {
          setStatus('ready')
          return
        }
      } catch {
        // backend not up yet
      }

      if (cancelled) return

      setAttempts(a => {
        const next = a + 1
        if (next >= MAX_ATTEMPTS) {
          setStatus('unavailable')
        } else {
          timer = setTimeout(check, POLL_INTERVAL)
        }
        return next
      })
    }

    check()
    return () => {
      cancelled = true
      clearTimeout(timer)
    }
  }, [])

  if (status === 'ready') return <>{children}</>

  return (
    <div style={{ display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', height: '100vh', gap: 12, color: 'var(--text-muted)' }}>
      {status === 'checking' ? (
        <>
          <div className="spinner" />
          <span>Connecting to server…</span>
        </>
      ) : (
        <span>Server unavailable. <button onClick={() => { setStatus('checking'); setAttempts(0) }}>Retry</button></span>
      )}
    </div>
  )
}
