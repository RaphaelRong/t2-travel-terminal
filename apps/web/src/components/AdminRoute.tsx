import { Navigate, Outlet } from 'react-router-dom'
import { useAuthStore } from '../store/authStore'

export function AdminRoute() {
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const role = useAuthStore((s) => s.role)
  const initialized = useAuthStore((s) => s.initialized)

  if (!initialized) {
    return (
      <div className="flex h-screen items-center justify-center bg-obsidian-base">
        <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>
      </div>
    )
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />
  }

  // 仅允许后端确认的超级管理员进入管理后台
  if (role !== 'superadmin') {
    return <Navigate to="/" replace />
  }

  return <Outlet />
}
