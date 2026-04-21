import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from './api/client'
import type { MeResponse } from './api/types'
import { useAuthStore } from './store/auth'
import LoginPage from './pages/Login'
import OnboardingPage from './pages/Onboarding'
import DashboardPage from './pages/Dashboard'
import AssetPage from './pages/Asset'
import PortfolioPage from './pages/Portfolio'
import HistoryPage from './pages/History'
import NavBar from './components/NavBar'

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { data, isLoading, isError } = useQuery<MeResponse>({
    queryKey: ['me'],
    queryFn: () => api.get<MeResponse>('/api/me'),
    retry: false,
  })
  const { setAuth } = useAuthStore()

  useEffect(() => {
    if (data) setAuth(data.user, data.balance, data.needs_onboarding)
  }, [data, setAuth])

  if (isLoading) return <div style={{ padding: 24, color: 'var(--text-muted)' }}>Loading...</div>
  if (isError || !data) return <Navigate to="/login" replace />
  if (data.needs_onboarding) return <Navigate to="/onboarding" replace />
  return <>{children}</>
}

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/onboarding" element={<OnboardingPage />} />
        <Route
          path="/*"
          element={
            <AuthGuard>
              <NavBar />
              <Routes>
                <Route path="/" element={<DashboardPage />} />
                <Route path="/asset/:symbol" element={<AssetPage />} />
                <Route path="/portfolio" element={<PortfolioPage />} />
                <Route path="/history" element={<HistoryPage />} />
              </Routes>
            </AuthGuard>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}
