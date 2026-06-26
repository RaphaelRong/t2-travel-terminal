import { useEffect, useState } from 'react'
import { Navigate, Outlet } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import { useTenantStore } from '../store/tenantStore'
import { useAuthStore } from '../store/authStore'
import type { UserRole } from '../lib/auth'

interface Subscription {
  id: string
  name: string
  slug?: string
  plan_id?: string
  plan_name: string
  effective_role: UserRole
  role: string
}

export function SubscriptionRoute() {
  const { currentTenant, setCurrentTenant } = useTenantStore()
  const setRole = useAuthStore((s) => s.setRole)
  const [checked, setChecked] = useState(false)

  const { data, isLoading, error } = useQuery({
    queryKey: ['subscriptions'],
    queryFn: async () => {
      const res = await api.get<{ subscriptions: Subscription[] }>('/tenants')
      return res.data.subscriptions
    },
  })

  useEffect(() => {
    if (data === undefined && !isLoading) {
      // wait for next render
      return
    }
    if (data && !currentTenant) {
      if (data.length > 0) {
        const selected = data[0]
        setCurrentTenant(selected)
        setRole(selected.effective_role)
      }
    }
    setChecked(true)
  }, [data, isLoading, currentTenant, setCurrentTenant, setRole])

  if (isLoading || !checked) {
    return (
      <div className="flex h-64 items-center justify-center">
        <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex h-64 items-center justify-center">
        <p className="font-mono text-sm text-obsidian-negative">
          Failed to load subscription: {(error as { message?: string }).message}
        </p>
      </div>
    )
  }

  const hasSubscription = data && data.length > 0
  if (!hasSubscription) {
    return <Navigate to="/plans" replace />
  }

  return <Outlet />
}
