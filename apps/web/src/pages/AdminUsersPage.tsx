import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'

interface SubscriptionSummary {
  id?: string
  name?: string
  plan_name?: string
  duration_months?: number
  price?: number
  currency?: string
  subscribed_at?: string
  expires_at?: string
  status?: string
}

interface User {
  id: string
  email: string
  name?: string
  email_verified: boolean
  is_superadmin: boolean
  created_at: string
  subscriptions: SubscriptionSummary[]
}

export function AdminUsersPage() {
  const { data, isLoading } = useQuery({
    queryKey: ['admin-users'],
    queryFn: async () => {
      const res = await api.get<{ users: User[] }>('/admin/users')
      return res.data.users
    },
  })

  if (isLoading) return <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>

  return (
    <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
      <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
        Users
      </h2>

      {data?.length === 0 && <p className="font-mono text-sm text-obsidian-text-secondary">No users yet.</p>}

      <div className="overflow-x-auto">
        <table className="w-full text-left font-mono text-sm">
          <thead>
            <tr className="border-b border-obsidian-border-dim">
              <th className="py-2 text-xs font-semibold uppercase tracking-wider text-obsidian-text-secondary">Email</th>
              <th className="py-2 text-xs font-semibold uppercase tracking-wider text-obsidian-text-secondary">Name</th>
              <th className="py-2 text-xs font-semibold uppercase tracking-wider text-obsidian-text-secondary">Verified</th>
              <th className="py-2 text-xs font-semibold uppercase tracking-wider text-obsidian-text-secondary">Role</th>
              <th className="py-2 text-xs font-semibold uppercase tracking-wider text-obsidian-text-secondary">Subscriptions</th>
              <th className="py-2 text-xs font-semibold uppercase tracking-wider text-obsidian-text-secondary">Created</th>
            </tr>
          </thead>
          <tbody>
            {data?.map((user) => (
              <tr key={user.id} className="border-b border-obsidian-border-dim">
                <td className="py-3 text-obsidian-text-primary">{user.email}</td>
                <td className="py-3 text-obsidian-text-secondary">{user.name || '-'}</td>
                <td className="py-3">
                  <span
                    className={`rounded px-2 py-0.5 font-mono text-[10px] uppercase ${
                      user.email_verified
                        ? 'bg-obsidian-positive-dim text-obsidian-positive'
                        : 'bg-obsidian-negative-dim text-obsidian-negative'
                    }`}
                  >
                    {user.email_verified ? 'Yes' : 'No'}
                  </span>
                </td>
                <td className="py-3">
                  {user.is_superadmin ? (
                    <span className="rounded bg-obsidian-accent/20 px-2 py-0.5 font-mono text-[10px] uppercase text-obsidian-accent">
                      SuperAdmin
                    </span>
                  ) : (
                    <span className="text-obsidian-text-tertiary">User</span>
                  )}
                </td>
                <td className="py-3">
                  {user.subscriptions.length === 0 ? (
                    <span className="text-obsidian-text-tertiary">-</span>
                  ) : (
                    <div className="space-y-1">
                      {user.subscriptions.map((sub, idx) => (
                        <div key={idx} className="text-xs text-obsidian-text-secondary">
                          {sub.name || 'Unnamed'} · {sub.plan_name || 'No plan'} · {sub.duration_months}mo ·{' '}
                          {sub.price} {sub.currency}
                          <span
                            className={`ml-1 rounded px-1 py-0.5 text-[10px] uppercase ${
                              sub.status === 'active'
                                ? 'bg-obsidian-positive-dim text-obsidian-positive'
                                : 'bg-obsidian-negative-dim text-obsidian-negative'
                            }`}
                          >
                            {sub.status}
                          </span>
                          <br />
                          <span className="text-obsidian-text-tertiary">
                            expires {sub.expires_at ? new Date(sub.expires_at).toLocaleDateString() : '-'}
                          </span>
                        </div>
                      ))}
                    </div>
                  )}
                </td>
                <td className="py-3 text-obsidian-text-tertiary">
                  {new Date(user.created_at).toLocaleDateString()}
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </section>
  )
}
