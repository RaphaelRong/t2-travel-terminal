import { create } from 'zustand'
import {
  getToken,
  removeToken,
  setToken,
  getSuperAdmin,
  setSuperAdmin as setStoredSuperAdmin,
  removeSuperAdmin,
  getRole,
  setRole as setStoredRole,
  removeRole,
  type UserRole,
} from '../lib/auth'

interface AuthState {
  token: string | null
  isAuthenticated: boolean
  isSuperAdmin: boolean
  role: UserRole | null
  initialized: boolean
  setToken: (token: string) => void
  setSuperAdmin: (isSuperAdmin: boolean) => void
  setRole: (role: UserRole) => void
  logout: () => void
  init: () => void
}

export const useAuthStore = create<AuthState>((set) => ({
  token: null,
  isAuthenticated: false,
  isSuperAdmin: false,
  role: null,
  initialized: false,
  setToken: (token) => {
    setToken(token)
    set({ token, isAuthenticated: true })
  },
  setSuperAdmin: (isSuperAdmin) => {
    setStoredSuperAdmin(isSuperAdmin)
    set({ isSuperAdmin })
  },
  setRole: (role) => {
    setStoredRole(role)
    set({ role })
  },
  logout: () => {
    removeToken()
    removeSuperAdmin()
    removeRole()
    set({ token: null, isAuthenticated: false, isSuperAdmin: false, role: null })
  },
  init: () => {
    const token = getToken()
    const isSuperAdmin = getSuperAdmin()
    const role = getRole()
    set({
      token,
      isAuthenticated: !!token,
      isSuperAdmin,
      role,
      initialized: true,
    })
  },
}))
