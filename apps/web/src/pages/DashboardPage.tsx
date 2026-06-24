import { useEffect, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { useTenantStore } from '../store/tenantStore'

interface Tenant {
  id: string
  name: string
  slug?: string
  plan_id: string
  role: string
}

export function DashboardPage() {
  const queryClient = useQueryClient()
  const { currentTenant, setCurrentTenant } = useTenantStore()
  const [newTenantName, setNewTenantName] = useState('')

  const { data: tenants, isLoading } = useQuery({
    queryKey: ['tenants'],
    queryFn: async () => {
      const res = await api.get<{ tenants: Tenant[] }>('/tenants')
      return res.data.tenants
    },
  })

  useEffect(() => {
    if (tenants && tenants.length > 0 && !currentTenant) {
      setCurrentTenant(tenants[0])
    }
  }, [tenants, currentTenant, setCurrentTenant])

  const createMutation = useMutation({
    mutationFn: (name: string) => api.post('/tenants', { name }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      setNewTenantName('')
    },
  })

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newTenantName.trim()) return
    createMutation.mutate(newTenantName.trim())
  }

  if (isLoading) return <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>

  return (
    <div className="space-y-6">
      <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
        <span className="text-obsidian-accent">&gt;</span> Dashboard
      </h1>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Your Workspaces
        </h2>
        {tenants?.length === 0 && <p className="font-mono text-sm text-obsidian-text-secondary">No workspaces yet.</p>}
        <div className="space-y-2">
          {tenants?.map((t) => (
            <button
              key={t.id}
              onClick={() => setCurrentTenant(t)}
              className={`flex w-full items-center justify-between border px-4 py-3 text-left font-mono text-sm transition-colors ${
                currentTenant?.id === t.id
                  ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-text-primary'
                  : 'border-obsidian-border-dim bg-obsidian-base text-obsidian-text-secondary hover:border-obsidian-border-med hover:text-obsidian-text-primary'
              }`}
            >
              <span>{t.name}</span>
              <span className="text-xs text-obsidian-text-tertiary">
                {t.plan_id} · {t.role}
              </span>
            </button>
          ))}
        </div>
      </section>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Create Workspace
        </h2>
        <form onSubmit={handleCreate} className="flex gap-2">
          <input
            type="text"
            value={newTenantName}
            onChange={(e) => setNewTenantName(e.target.value)}
            placeholder="Workspace name"
            className="flex-1 border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
          />
          <button
            type="submit"
            disabled={createMutation.isPending}
            className="border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            Create
          </button>
        </form>
        {createMutation.isError && (
          <p className="mt-2 font-mono text-xs text-obsidian-negative">
            {(createMutation.error as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Failed to create workspace'}
          </p>
        )}
      </section>
    </div>
  )
}
