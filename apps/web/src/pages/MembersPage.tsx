import { useState } from 'react'
import { useQuery, useMutation } from '@tanstack/react-query'
import { api } from '../lib/api'
import { useTenantStore } from '../store/tenantStore'

interface Member {
  user_id: string
  email: string
  name?: string
  role: string
  joined_at: string
}

export function MembersPage() {
  const { currentTenant } = useTenantStore()
  const [inviteEmail, setInviteEmail] = useState('')
  const [inviteRole, setInviteRole] = useState('member')
  const [inviteResult, setInviteResult] = useState('')

  const { data: members, isLoading } = useQuery({
    queryKey: ['members', currentTenant?.id],
    queryFn: async () => {
      const res = await api.get<{ members: Member[] }>('/tenants/current/members')
      return res.data.members
    },
    enabled: !!currentTenant,
  })

  const inviteMutation = useMutation({
    mutationFn: (payload: { email: string; role: string }) =>
      api.post('/tenants/current/members', payload),
    onSuccess: (res) => {
      setInviteResult(`Invite token: ${res.data.token}`)
      setInviteEmail('')
    },
  })

  const handleInvite = (e: React.FormEvent) => {
    e.preventDefault()
    if (!inviteEmail.trim()) return
    inviteMutation.mutate({ email: inviteEmail.trim(), role: inviteRole })
  }

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
      <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
        <span className="text-obsidian-accent">&gt;</span> Members
      </h1>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Invite Member
        </h2>
        <form onSubmit={handleInvite} className="flex flex-col gap-2 sm:flex-row">
          <input
            type="email"
            value={inviteEmail}
            onChange={(e) => setInviteEmail(e.target.value)}
            placeholder="Email"
            className="flex-1 border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
          />
          <select
            value={inviteRole}
            onChange={(e) => setInviteRole(e.target.value)}
            className="border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none"
          >
            <option value="member">Member</option>
            <option value="admin">Admin</option>
          </select>
          <button
            type="submit"
            disabled={inviteMutation.isPending}
            className="border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            Invite
          </button>
        </form>
        {inviteResult && (
          <p className="mt-3 break-all border border-obsidian-positive-dim bg-obsidian-positive-dim/20 p-2 font-mono text-xs text-obsidian-positive">
            {inviteResult}
          </p>
        )}
        {inviteMutation.isError && (
          <p className="mt-2 font-mono text-xs text-obsidian-negative">
            {(inviteMutation.error as { response?: { data?: { error?: string } } })?.response?.data?.error || 'Invite failed'}
          </p>
        )}
      </section>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Member List
        </h2>
        {isLoading ? (
          <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>
        ) : members?.length === 0 ? (
          <p className="font-mono text-sm text-obsidian-text-secondary">No members yet.</p>
        ) : (
          <table className="w-full text-left font-mono text-sm">
            <thead>
              <tr className="border-b border-obsidian-border-dim">
                <th className="py-2 text-xs font-semibold uppercase tracking-wider text-obsidian-text-secondary">Email</th>
                <th className="py-2 text-xs font-semibold uppercase tracking-wider text-obsidian-text-secondary">Role</th>
                <th className="py-2 text-xs font-semibold uppercase tracking-wider text-obsidian-text-secondary">Joined</th>
              </tr>
            </thead>
            <tbody>
              {members?.map((m) => (
                <tr key={m.user_id} className="border-b border-obsidian-border-dim">
                  <td className="py-2 text-obsidian-text-primary">{m.email}</td>
                  <td className="py-2 capitalize text-obsidian-text-secondary">{m.role}</td>
                  <td className="py-2 text-obsidian-text-tertiary">
                    {new Date(m.joined_at).toLocaleDateString()}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </section>
    </div>
  )
}
