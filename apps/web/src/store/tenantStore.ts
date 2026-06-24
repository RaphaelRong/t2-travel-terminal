import { create } from 'zustand'

interface Tenant {
  id: string
  name: string
  slug?: string
  plan_id: string
  role: string
}

interface TenantState {
  currentTenant: Tenant | null
  setCurrentTenant: (tenant: Tenant | null) => void
}

export const useTenantStore = create<TenantState>((set) => ({
  currentTenant: null,
  setCurrentTenant: (tenant) => set({ currentTenant: tenant }),
}))
