import { Navigate, Outlet, useLocation } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

export function ProtectedRoute() {
  const location = useLocation()
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const initialized = useAuthStore((s) => s.initialized)

  if (!initialized) {
    return (
      <div className="flex h-screen items-center justify-center bg-obsidian-base">
        <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>
      </div>
    )
  }

  const redirect = encodeURIComponent(`${location.pathname}${location.search}`)

  return isAuthenticated ? <Outlet /> : <Navigate to={`/login?redirect=${redirect}`} replace />
}
