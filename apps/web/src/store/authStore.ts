import { create } from 'zustand'
import { getToken, removeToken, setToken } from '../lib/auth'

interface AuthState {
  token: string | null
  isAuthenticated: boolean
  setToken: (token: string) => void
  logout: () => void
  init: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  isAuthenticated: false,
  setToken: (token) => {
    setToken(token)
    set({ token, isAuthenticated: true })
  },
  logout: () => {
    removeToken()
    set({ token: null, isAuthenticated: false })
  },
  init: () => {
    const token = getToken()
    set({ token, isAuthenticated: !!token })
  },
}))
