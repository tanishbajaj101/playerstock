import { create } from 'zustand'
import type { User, Balance } from '../api/types'

interface AuthState {
  user: User | null
  balance: Balance | null
  needsOnboarding: boolean
  setAuth: (user: User, balance: Balance, needsOnboarding: boolean) => void
  setBalance: (balance: Balance) => void
  clearAuth: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  user: null,
  balance: null,
  needsOnboarding: false,
  setAuth: (user, balance, needsOnboarding) => set({ user, balance, needsOnboarding }),
  setBalance: (balance) => set({ balance }),
  clearAuth: () => set({ user: null, balance: null, needsOnboarding: false }),
}))
