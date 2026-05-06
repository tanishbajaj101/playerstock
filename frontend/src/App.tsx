import { useEffect } from 'react'
import { BrowserRouter, Routes, Route, Navigate, useLocation } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from './api/client'
import type { MeResponse } from './api/types'
import { useAuthStore } from './store/auth'
import LoginPage from './pages/Login'
import OnboardingPage from './pages/Onboarding'
import WelcomePage from './pages/Welcome'
import DashboardPage from './pages/Dashboard'
import AssetPage from './pages/Asset'
import AssetsPage from './pages/Assets'
import OrdersPage from './pages/Orders'
import NavBar from './components/NavBar'

function AuthGuard({ children }: { children: React.ReactNode }) {
  const { data, isLoading, isError } = useQuery<MeResponse>({
    queryKey: ['me'],
    queryFn: () => api.get<MeResponse>('/api/me'),
    retry: false,
  })
  const { setAuth } = useAuthStore()
  const location = useLocation()

  useEffect(() => {
    if (data) setAuth(data.user, data.balance, data.needs_onboarding)
  }, [data, setAuth])

  if (isLoading) return <div style={{ padding: 24, color: 'var(--text-muted)' }}>Loading...</div>
  if (isError || !data) return <Navigate to="/login" replace />
  if (data.needs_onboarding) return <Navigate to="/onboarding" replace />
  if (data.needs_welcome && location.pathname !== '/welcome') return <Navigate to="/welcome" replace />
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
                <Route path="/assets" element={<AssetsPage />} />
                <Route path="/orders" element={<OrdersPage />} />
                <Route path="/welcome" element={<WelcomePage />} />
              </Routes>
            </AuthGuard>
          }
        />
      </Routes>
    </BrowserRouter>
  )
}
