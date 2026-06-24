import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { useTenantStore } from '../store/tenantStore'

interface Plan {
  id: string
  name: string
  description: string
  price_monthly: number
  features: string[]
}

export function PlansPage() {
  const { currentTenant, setCurrentTenant } = useTenantStore()
  const queryClient = useQueryClient()

  const { data: plans, isLoading } = useQuery({
    queryKey: ['plans'],
    queryFn: async () => {
      const res = await api.get<{ plans: Plan[] }>('/plans')
      return res.data.plans
    },
  })

  const activateMutation = useMutation({
    mutationFn: (planId: string) =>
      api.put('/tenants/current/plan', { plan_id: planId }),
    onSuccess: (_, planId) => {
      queryClient.invalidateQueries({ queryKey: ['tenants'] })
      if (currentTenant) {
        setCurrentTenant({ ...currentTenant, plan_id: planId })
      }
    },
  })

  if (!currentTenant) {
    return (
      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <p className="font-mono text-sm text-obsidian-text-secondary">
          Please select or create a workspace first.
        </p>
      </section>
    )
  }

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
          <span className="text-obsidian-accent">&gt;</span> Plans
        </h1>
        <p className="mt-1 font-mono text-sm text-obsidian-text-secondary">
          Current plan: <span className="text-obsidian-text-primary">{currentTenant.plan_id}</span>
        </p>
      </div>

      {isLoading ? (
        <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>
      ) : (
        <div className="grid gap-4 md:grid-cols-3">
          {plans?.map((plan) => (
            <div
              key={plan.id}
              className={`border bg-obsidian-surface p-6 transition-colors ${
                currentTenant.plan_id === plan.id
                  ? 'border-obsidian-accent'
                  : 'border-obsidian-border-dim hover:border-obsidian-border-med'
              }`}
            >
              <h2 className="font-mono text-lg font-bold text-obsidian-text-primary">{plan.name}</h2>
              <p className="mt-1 font-mono text-xs text-obsidian-text-secondary">{plan.description}</p>
              <p className="mt-4 font-mono text-2xl font-bold text-obsidian-text-primary">
                ${plan.price_monthly}
                <span className="text-sm font-normal text-obsidian-text-tertiary">/mo</span>
              </p>
              <ul className="mt-4 space-y-2 font-mono text-xs text-obsidian-text-secondary">
                {plan.features.map((f, i) => (
                  <li key={i} className="flex items-start gap-2">
                    <span className="text-obsidian-accent">&gt;</span>
                    <span>{f}</span>
                  </li>
                ))}
              </ul>
              <button
                onClick={() => activateMutation.mutate(plan.id)}
                disabled={currentTenant.plan_id === plan.id || activateMutation.isPending}
                className="mt-6 w-full border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
              >
                {currentTenant.plan_id === plan.id ? 'Current Plan' : 'Activate'}
              </button>
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
