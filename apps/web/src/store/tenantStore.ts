import { create } from 'zustand'

interface Pricing {
  id: string
  duration_months: number
  price: number
  currency: string
}

interface Tenant {
  id: string
  name: string
  slug?: string
  plan_id?: string
  plan_name?: string
  pricing?: Pricing
  pricing_id?: string
  role: string
  subscribed_at?: string
  expires_at?: string
  auto_renew?: boolean
  status?: string
}

interface TenantState {
  currentTenant: Tenant | null
  setCurrentTenant: (tenant: Tenant | null) => void
}

export const useTenantStore = create<TenantState>((set) => ({
  currentTenant: null,
  setCurrentTenant: (tenant) => set({ currentTenant: tenant }),
}))
