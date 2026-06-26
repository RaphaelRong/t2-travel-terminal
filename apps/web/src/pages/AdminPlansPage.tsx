import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'

interface Pricing {
  id: string
  duration_months: number
  price: number
  currency: string
  status: string
}

interface Plan {
  id: string
  name: string
  description: string
  status: string
  pricing: Pricing[]
}

export function AdminPlansPage() {
  const queryClient = useQueryClient()
  const [editingPlan, setEditingPlan] = useState<Plan | null>(null)
  const [planForm, setPlanForm] = useState({ name: '', description: '', status: 'active' })
  const [pricingForm, setPricingForm] = useState({ plan_id: '', duration_months: 1, price: 0, currency: 'USD' })

  const { data, isLoading } = useQuery({
    queryKey: ['admin-plans'],
    queryFn: async () => {
      const res = await api.get<{ plans: Plan[] }>('/admin/plans')
      return res.data.plans
    },
  })

  const createPlan = useMutation({
    mutationFn: (payload: { name: string; description: string; status: string }) =>
      api.post('/admin/plans', payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-plans'] })
      queryClient.invalidateQueries({ queryKey: ['plans'] })
      setPlanForm({ name: '', description: '', status: 'active' })
    },
  })

  const updatePlan = useMutation({
    mutationFn: (payload: { id: string; data: Partial<Plan> }) =>
      api.put(`/admin/plans/${payload.id}`, payload.data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-plans'] })
      queryClient.invalidateQueries({ queryKey: ['plans'] })
      setEditingPlan(null)
    },
  })

  const deletePlan = useMutation({
    mutationFn: (id: string) => api.delete(`/admin/plans/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-plans'] })
      queryClient.invalidateQueries({ queryKey: ['plans'] })
    },
  })

  const createPricing = useMutation({
    mutationFn: (payload: { plan_id: string; duration_months: number; price: number; currency: string }) =>
      api.post(`/admin/plans/${payload.plan_id}/pricing`, payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-plans'] })
      queryClient.invalidateQueries({ queryKey: ['plans'] })
      setPricingForm({ plan_id: '', duration_months: 1, price: 0, currency: 'USD' })
    },
  })

  const updatePricing = useMutation({
    mutationFn: (payload: { plan_id: string; pricing_id: string; status: string }) =>
      api.put(`/admin/plans/${payload.plan_id}/pricing/${payload.pricing_id}`, { status: payload.status }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-plans'] })
      queryClient.invalidateQueries({ queryKey: ['plans'] })
    },
  })

  const deletePricing = useMutation({
    mutationFn: (payload: { plan_id: string; pricing_id: string }) =>
      api.delete(`/admin/plans/${payload.plan_id}/pricing/${payload.pricing_id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-plans'] })
      queryClient.invalidateQueries({ queryKey: ['plans'] })
    },
  })

  const handlePlanSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (editingPlan) {
      updatePlan.mutate({ id: editingPlan.id, data: planForm })
    } else {
      createPlan.mutate(planForm)
    }
  }

  const handlePricingSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!pricingForm.plan_id) return
    createPricing.mutate(pricingForm)
  }

  if (isLoading) return <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>

  return (
    <div className="space-y-8">
      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          {editingPlan ? 'Edit Plan' : 'Create Plan'}
        </h2>
        <form onSubmit={handlePlanSubmit} className="space-y-4">
          <input
            type="text"
            placeholder="Plan name"
            value={planForm.name}
            onChange={(e) => setPlanForm({ ...planForm, name: e.target.value })}
            className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          />
          <textarea
            placeholder="Description"
            value={planForm.description}
            onChange={(e) => setPlanForm({ ...planForm, description: e.target.value })}
            className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          />
          <select
            value={planForm.status}
            onChange={(e) => setPlanForm({ ...planForm, status: e.target.value })}
            className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          >
            <option value="active">Active</option>
            <option value="inactive">Inactive</option>
          </select>
          <div className="flex gap-2">
            <button
              type="submit"
              disabled={createPlan.isPending || updatePlan.isPending}
              className="border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
            >
              {editingPlan ? 'Update' : 'Create'}
            </button>
            {editingPlan && (
              <button
                type="button"
                onClick={() => {
                  setEditingPlan(null)
                  setPlanForm({ name: '', description: '', status: 'active' })
                }}
                className="border border-obsidian-border-dim bg-obsidian-base px-4 py-2 font-mono text-sm text-obsidian-text-secondary transition-colors hover:border-obsidian-border-med"
              >
                Cancel
              </button>
            )}
          </div>
        </form>
      </section>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Add Pricing
        </h2>
        <form onSubmit={handlePricingSubmit} className="grid gap-4 sm:grid-cols-5">
          <select
            value={pricingForm.plan_id}
            onChange={(e) => setPricingForm({ ...pricingForm, plan_id: e.target.value })}
            className="border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          >
            <option value="">Select plan</option>
            {data?.map((p) => (
              <option key={p.id} value={p.id}>
                {p.name}
              </option>
            ))}
          </select>
          <input
            type="number"
            min={1}
            placeholder="Months"
            value={pricingForm.duration_months}
            onChange={(e) => setPricingForm({ ...pricingForm, duration_months: parseInt(e.target.value) || 0 })}
            className="border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          />
          <input
            type="number"
            min={0}
            step="0.01"
            placeholder="Price"
            value={pricingForm.price}
            onChange={(e) => setPricingForm({ ...pricingForm, price: parseFloat(e.target.value) || 0 })}
            className="border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          />
          <input
            type="text"
            placeholder="Currency"
            value={pricingForm.currency}
            onChange={(e) => setPricingForm({ ...pricingForm, currency: e.target.value })}
            className="border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          />
          <button
            type="submit"
            disabled={createPricing.isPending || !pricingForm.plan_id}
            className="border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            Add Pricing
          </button>
        </form>
      </section>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Plans
        </h2>
        <div className="space-y-4">
          {data?.map((plan) => (
            <div key={plan.id} className="border border-obsidian-border-dim p-4">
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="font-mono text-lg font-bold text-obsidian-text-primary">
                    {plan.name}
                    <span
                      className={`ml-2 inline-block rounded px-2 py-0.5 font-mono text-[10px] uppercase ${
                        plan.status === 'active'
                          ? 'bg-obsidian-positive-dim text-obsidian-positive'
                          : 'bg-obsidian-negative-dim text-obsidian-negative'
                      }`}
                    >
                      {plan.status}
                    </span>
                  </h3>
                  <p className="font-mono text-sm text-obsidian-text-secondary">{plan.description}</p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => {
                      setEditingPlan(plan)
                      setPlanForm({ name: plan.name, description: plan.description, status: plan.status })
                    }}
                    className="border border-obsidian-border-dim bg-obsidian-base px-3 py-1 font-mono text-xs text-obsidian-text-secondary transition-colors hover:border-obsidian-border-med"
                  >
                    Edit
                  </button>
                  <button
                    onClick={() => deletePlan.mutate(plan.id)}
                    disabled={deletePlan.isPending}
                    className="border border-obsidian-negative-dim bg-obsidian-negative-dim/20 px-3 py-1 font-mono text-xs text-obsidian-negative transition-colors hover:bg-obsidian-negative-dim/40 disabled:opacity-50"
                  >
                    Delete
                  </button>
                </div>
              </div>

              <div className="mt-4">
                <h4 className="mb-2 font-mono text-xs font-semibold uppercase text-obsidian-text-secondary">
                  Pricing
                </h4>
                {plan.pricing.length === 0 ? (
                  <p className="font-mono text-xs text-obsidian-text-tertiary">No pricing yet.</p>
                ) : (
                  <div className="grid gap-2 sm:grid-cols-2 lg:grid-cols-3">
                    {plan.pricing.map((pp) => (
                      <div
                        key={pp.id}
                        className={`flex items-center justify-between border px-3 py-2 font-mono text-xs ${
                          pp.status === 'active'
                            ? 'border-obsidian-border-dim'
                            : 'border-obsidian-negative-dim opacity-60'
                        }`}
                      >
                        <span className="text-obsidian-text-secondary">
                          {pp.duration_months}mo · {pp.price} {pp.currency}
                        </span>
                        <div className="flex gap-2">
                          <button
                            onClick={() =>
                              updatePricing.mutate({
                                plan_id: plan.id,
                                pricing_id: pp.id,
                                status: pp.status === 'active' ? 'inactive' : 'active',
                              })
                            }
                            className="text-obsidian-accent hover:underline"
                          >
                            {pp.status === 'active' ? 'Deactivate' : 'Activate'}
                          </button>
                          <button
                            onClick={() => deletePricing.mutate({ plan_id: plan.id, pricing_id: pp.id })}
                            className="text-obsidian-negative hover:underline"
                          >
                            Delete
                          </button>
                        </div>
                      </div>
                    ))}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </section>
    </div>
  )
}
