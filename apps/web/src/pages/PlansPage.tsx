import { useEffect, useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { useTenantStore } from '../store/tenantStore'
import { useAuthStore } from '../store/authStore'
import type { UserRole } from '../lib/auth'

interface Pricing {
  id: string
  duration_months: number
  price: number
  currency: string
}

interface Plan {
  id: string
  name: string
  description: string
  pricing: Pricing[]
}

interface Subscription {
  id: string
  name: string
  slug?: string
  plan_id?: string
  plan_name: string
  effective_role: UserRole
  pricing?: Pricing
  role: string
  subscribed_at?: string
  expires_at?: string
  auto_renew: boolean
  status: string
}

export function PlansPage() {
  const queryClient = useQueryClient()
  const { currentTenant, setCurrentTenant } = useTenantStore()
  const setRole = useAuthStore((s) => s.setRole)
  const [selectedPlanId, setSelectedPlanId] = useState('')
  const [selectedPricingId, setSelectedPricingId] = useState('')

  const {
    data: subscriptions,
    isLoading: subscriptionsLoading,
    error: subscriptionsError,
  } = useQuery({
    queryKey: ['subscriptions'],
    queryFn: async () => {
      const res = await api.get<{ subscriptions: Subscription[] }>('/tenants')
      return res.data.subscriptions
    },
  })

  const { data: plans, isLoading: plansLoading } = useQuery({
    queryKey: ['plans'],
    queryFn: async () => {
      const res = await api.get<{ plans: Plan[] }>('/plans')
      return res.data.plans
    },
  })

  useEffect(() => {
    if (subscriptions && subscriptions.length > 0 && !currentTenant) {
      const selected = subscriptions[0]
      setCurrentTenant(selected)
      setRole(selected.effective_role)
    }
  }, [subscriptions, currentTenant, setCurrentTenant, setRole])

  useEffect(() => {
    if (plans && plans.length > 0 && !selectedPlanId) {
      setSelectedPlanId(plans[0].id)
      if (plans[0].pricing.length > 0) {
        setSelectedPricingId(plans[0].pricing[0].id)
      }
    }
  }, [plans, selectedPlanId])

  const selectedPlan = plans?.find((p) => p.id === selectedPlanId)

  const createMutation = useMutation({
    mutationFn: (payload: {
      plan_id: string
      pricing_id: string
      auto_renew: boolean
    }) => api.post('/tenants', payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['subscriptions'] })
    },
  })

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedPlanId || !selectedPricingId) return
    createMutation.mutate({
      plan_id: selectedPlanId,
      pricing_id: selectedPricingId,
      auto_renew: false,
    })
  }

  const isLoading = subscriptionsLoading || plansLoading

  if (isLoading) {
    return <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>
  }

  if (subscriptionsError) {
    return (
      <div className="rounded border border-obsidian-negative-dim bg-obsidian-negative-dim/20 p-6">
        <p className="font-mono text-sm text-obsidian-negative">
          Failed to load subscriptions:{' '}
          {(subscriptionsError as { response?: { data?: { error?: string } }; message?: string })?.response?.data
            ?.error || (subscriptionsError as { message?: string }).message}
        </p>
      </div>
    )
  }

  const hasSubscription = subscriptions && subscriptions.length > 0

  if (!hasSubscription) {
    return (
      <div className="mx-auto max-w-2xl space-y-8 py-8">
        <div className="text-center">
          <h1 className="font-mono text-3xl font-bold tracking-tight text-obsidian-text-primary">
            Choose a Plan
          </h1>
          <p className="mt-2 font-mono text-sm text-obsidian-text-secondary">
            Please choose a subscription plan to continue using T2 Travel Terminal.
          </p>
        </div>

        <section className="border border-obsidian-accent bg-obsidian-surface p-8">
          <h2 className="mb-6 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
            Create Subscription
          </h2>
          <form onSubmit={handleCreate} className="space-y-6">
            <div>
              <label className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
                Plan
              </label>
              <select
                value={selectedPlanId}
                onChange={(e) => {
                  setSelectedPlanId(e.target.value)
                  const plan = plans?.find((p) => p.id === e.target.value)
                  setSelectedPricingId(plan?.pricing[0]?.id || '')
                }}
                className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
              >
                {plans?.map((p) => (
                  <option key={p.id} value={p.id}>
                    {p.name}
                  </option>
                ))}
              </select>
              {selectedPlan && (
                <p className="mt-1 font-mono text-xs text-obsidian-text-secondary">{selectedPlan.description}</p>
              )}
            </div>

            <div>
              <label className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
                Duration
              </label>
              <div className="grid gap-2 sm:grid-cols-2">
                {selectedPlan?.pricing.map((pp) => (
                  <label
                    key={pp.id}
                    className={`flex cursor-pointer items-center justify-between border px-4 py-3 font-mono text-sm transition-colors ${
                      selectedPricingId === pp.id
                        ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-text-primary'
                        : 'border-obsidian-border-dim text-obsidian-text-secondary hover:border-obsidian-border-med'
                    }`}
                  >
                    <span className="flex items-center gap-2">
                      <input
                        type="radio"
                        name="pricing"
                        checked={selectedPricingId === pp.id}
                        onChange={() => setSelectedPricingId(pp.id)}
                        className="accent-obsidian-accent"
                      />
                      {pp.duration_months} months
                    </span>
                    <span className="text-obsidian-text-primary">
                      {pp.price} {pp.currency}
                    </span>
                  </label>
                ))}
              </div>
            </div>

            <button
              type="submit"
              disabled={createMutation.isPending || !selectedPricingId}
              className="w-full border border-obsidian-accent bg-obsidian-accent px-4 py-3 font-mono text-sm font-semibold text-white transition-colors hover:bg-obsidian-accent/90 disabled:opacity-50"
            >
              {createMutation.isPending ? 'Subscribing...' : 'Subscribe & Continue'}
            </button>
          </form>
          {createMutation.isError && (
            <p className="mt-4 font-mono text-xs text-obsidian-negative">
              {(createMutation.error as { response?: { data?: { error?: string } } })?.response?.data?.error ||
                'Failed to create subscription'}
            </p>
          )}
        </section>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
        <span className="text-obsidian-accent">&gt;</span> Plans
      </h1>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Your Subscriptions
        </h2>
        <div className="space-y-2">
          {subscriptions?.map((s) => (
            <button
              key={s.id}
              onClick={() => {
                setCurrentTenant(s)
                setRole(s.effective_role)
              }}
              className={`flex w-full items-center justify-between border px-4 py-3 text-left font-mono text-sm transition-colors ${
                currentTenant?.id === s.id
                  ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-text-primary'
                  : 'border-obsidian-border-dim bg-obsidian-base text-obsidian-text-secondary hover:border-obsidian-border-med hover:text-obsidian-text-primary'
              }`}
            >
              <span>{s.name}</span>
              <span className="text-xs text-obsidian-text-tertiary">
                {s.plan_name} · {s.pricing ? `${s.pricing.duration_months}mo` : '-'} · {s.role}
              </span>
            </button>
          ))}
        </div>
      </section>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Subscribe to T2
        </h2>
        <form onSubmit={handleCreate} className="space-y-4">
          <div className="grid gap-4 sm:grid-cols-2">
            <select
              value={selectedPlanId}
              onChange={(e) => {
                setSelectedPlanId(e.target.value)
                const plan = plans?.find((p) => p.id === e.target.value)
                setSelectedPricingId(plan?.pricing[0]?.id || '')
              }}
              className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
            >
              {plans?.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>

            <select
              value={selectedPricingId}
              onChange={(e) => setSelectedPricingId(e.target.value)}
              className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
            >
              {selectedPlan?.pricing.map((pp) => (
                <option key={pp.id} value={pp.id}>
                  {pp.duration_months} months · {pp.price} {pp.currency}
                </option>
              ))}
            </select>
          </div>

          {selectedPlan && (
            <p className="font-mono text-xs text-obsidian-text-secondary">{selectedPlan.description}</p>
          )}

          <button
            type="submit"
            disabled={createMutation.isPending || !selectedPricingId}
            className="w-full border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            Subscribe
          </button>
        </form>
        {createMutation.isError && (
          <p className="mt-2 font-mono text-xs text-obsidian-negative">
            {(createMutation.error as { response?: { data?: { error?: string } } })?.response?.data?.error ||
              'Failed to create subscription'}
          </p>
        )}
      </section>
    </div>
  )
}
