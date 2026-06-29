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

const planFeatureCatalog: Record<string, { tagline: string; features: string[]; highlight?: string }> = {
  'free trial': {
    tagline: 'Explore the terminal with a lightweight personal workspace.',
    features: [
      '1 active workspace for evaluation',
      'Basic Agent Playground conversations',
      'Public/system project browsing',
      'Community-level support',
    ],
  },
  basic: {
    tagline: 'Run real team workflows with project context and shared data sources.',
    features: [
      'Team workspace and member collaboration',
      'Project data sources and capability management',
      'LLM profile configuration for agents',
      'Standard support for operational rollout',
    ],
    highlight: 'Recommended',
  },
  advanced: {
    tagline: 'Scale agent operations with enterprise data, governance, and support.',
    features: [
      'Advanced Agent workflows and premium capabilities',
      'Larger project and integration capacity',
      'Tenant governance for production teams',
      'Priority support for enterprise deployment',
    ],
  },
}

function getPlanPresentation(plan: Plan) {
  const key = plan.name.toLowerCase()
  return (
    planFeatureCatalog[key] || {
      tagline: plan.description || 'A flexible plan for T2 Travel Terminal.',
      features: [
        'Agent workspace access',
        'Project context management',
        'Subscription-based team access',
        'T2 platform updates',
      ],
    }
  )
}

function formatPrice(pricing: Pricing) {
  return `${pricing.price} ${pricing.currency}`
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

  const handleSubscribe = (plan: Plan, pricingID: string) => {
    if (!pricingID) return
    setSelectedPlanId(plan.id)
    setSelectedPricingId(pricingID)
    createMutation.mutate({
      plan_id: plan.id,
      pricing_id: pricingID,
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
      <div className="mx-auto max-w-6xl space-y-8 py-8">
        <div className="text-center">
          <h1 className="font-mono text-3xl font-bold tracking-tight text-obsidian-text-primary">
            Choose a Plan
          </h1>
          <p className="mt-2 font-mono text-sm text-obsidian-text-secondary">
            Please choose a subscription plan to continue using T2 Travel Terminal.
          </p>
        </div>

        <section className="grid items-stretch gap-4 lg:grid-cols-3">
          {plans?.map((plan) => {
            const presentation = getPlanPresentation(plan)
            const selectedPricingForPlan =
              selectedPlanId === plan.id ? selectedPricingId || plan.pricing[0]?.id || '' : plan.pricing[0]?.id || ''
            return (
              <article
                key={plan.id}
                className={`flex h-full min-h-[520px] flex-col border bg-obsidian-surface p-6 transition-colors ${
                  selectedPlanId === plan.id
                    ? 'border-obsidian-accent'
                    : 'border-obsidian-border-dim hover:border-obsidian-border-med'
                }`}
              >
                <div className="min-h-[138px] border-b border-obsidian-border-dim pb-5">
                  <div className="mb-3 flex min-h-[24px] items-center justify-between gap-3">
                    <p className="font-mono text-xs uppercase tracking-wider text-obsidian-accent">
                      {presentation.highlight || 'Plan'}
                    </p>
                    {selectedPlanId === plan.id && (
                      <span className="border border-obsidian-accent px-2 py-1 font-mono text-[10px] uppercase tracking-wider text-obsidian-accent">
                        Selected
                      </span>
                    )}
                  </div>
                  <h2 className="font-mono text-2xl font-bold text-obsidian-text-primary">{plan.name}</h2>
                  <p className="mt-3 font-mono text-sm leading-6 text-obsidian-text-secondary">
                    {presentation.tagline}
                  </p>
                </div>

                <ul className="min-h-[176px] flex-1 space-y-3 py-5">
                  {presentation.features.map((feature) => (
                    <li key={feature} className="flex gap-3 font-mono text-sm leading-6 text-obsidian-text-secondary">
                      <span className="mt-2 h-1.5 w-1.5 shrink-0 bg-obsidian-accent" />
                      <span>{feature}</span>
                    </li>
                  ))}
                </ul>

                <div className="space-y-3 border-t border-obsidian-border-dim pt-5">
                  <p className="font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
                    Billing duration
                  </p>
                  <div className="space-y-2">
                    {plan.pricing.map((pp) => (
                      <label
                        key={pp.id}
                        className={`flex cursor-pointer items-center justify-between border px-3 py-3 font-mono text-sm transition-colors ${
                          selectedPricingForPlan === pp.id
                            ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-text-primary'
                            : 'border-obsidian-border-dim text-obsidian-text-secondary hover:border-obsidian-border-med'
                        }`}
                      >
                        <span className="flex items-center gap-2">
                          <input
                            type="radio"
                            name={`pricing-${plan.id}`}
                            checked={selectedPricingForPlan === pp.id}
                            onChange={() => {
                              setSelectedPlanId(plan.id)
                              setSelectedPricingId(pp.id)
                            }}
                            className="accent-obsidian-accent"
                          />
                          {pp.duration_months} months
                        </span>
                        <span className="text-obsidian-text-primary">{formatPrice(pp)}</span>
                      </label>
                    ))}
                  </div>

                  <button
                    type="button"
                    onClick={() => handleSubscribe(plan, selectedPricingForPlan)}
                    disabled={createMutation.isPending || !selectedPricingForPlan}
                    className="w-full border border-obsidian-accent bg-obsidian-accent px-4 py-3 font-mono text-sm font-semibold text-white transition-colors hover:bg-obsidian-accent/90 disabled:cursor-not-allowed disabled:opacity-50"
                  >
                    {createMutation.isPending && selectedPlanId === plan.id
                      ? 'Subscribing...'
                      : `Choose ${plan.name}`}
                  </button>
                </div>
              </article>
            )
          })}
        </section>

        <section>
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
        <div className="grid items-stretch gap-4 lg:grid-cols-3">
          {plans?.map((plan) => {
            const presentation = getPlanPresentation(plan)
            const pricingID =
              selectedPlanId === plan.id ? selectedPricingId || plan.pricing[0]?.id || '' : plan.pricing[0]?.id || ''
            return (
              <article
                key={plan.id}
                className="flex h-full min-h-[430px] flex-col border border-obsidian-border-dim bg-obsidian-base p-5"
              >
                <div className="min-h-[112px]">
                  <p className="font-mono text-xs uppercase tracking-wider text-obsidian-accent">
                    {presentation.highlight || 'Plan'}
                  </p>
                  <h3 className="mt-2 font-mono text-xl font-bold text-obsidian-text-primary">{plan.name}</h3>
                  <p className="mt-2 font-mono text-xs leading-5 text-obsidian-text-secondary">
                    {presentation.tagline}
                  </p>
                </div>
                <div className="flex-1 space-y-2 py-4">
                  {plan.pricing.map((pp) => (
                    <label
                      key={pp.id}
                      className={`flex cursor-pointer items-center justify-between border px-3 py-2 font-mono text-xs ${
                        pricingID === pp.id
                          ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-text-primary'
                          : 'border-obsidian-border-dim text-obsidian-text-secondary'
                      }`}
                    >
                      <span className="flex items-center gap-2">
                        <input
                          type="radio"
                          name={`existing-pricing-${plan.id}`}
                          checked={pricingID === pp.id}
                          onChange={() => {
                            setSelectedPlanId(plan.id)
                            setSelectedPricingId(pp.id)
                          }}
                          className="accent-obsidian-accent"
                        />
                        {pp.duration_months} months
                      </span>
                      <span>{formatPrice(pp)}</span>
                    </label>
                  ))}
                </div>
                <button
                  type="button"
                  onClick={() => handleSubscribe(plan, pricingID)}
                  disabled={createMutation.isPending || !pricingID}
                  className="w-full border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
                >
                  Subscribe
                </button>
              </article>
            )
          })}
        </div>
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
