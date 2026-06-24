import { useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { removeToken } from '../lib/auth'

interface Profile {
  id: string
  email: string
  name?: string
  email_verified: boolean
}

export function ProfilePage() {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const { data: profile, isLoading } = useQuery({
    queryKey: ['me'],
    queryFn: async () => {
      const res = await api.get<Profile>('/me')
      return res.data
    },
  })

  const [name, setName] = useState(profile?.name || '')
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [confirmEmail, setConfirmEmail] = useState('')
  const [deleteError, setDeleteError] = useState('')

  const updateMutation = useMutation({
    mutationFn: (newName: string) => api.put('/me', { name: newName }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me'] })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: () => api.delete('/me'),
    onSuccess: () => {
      removeToken()
      queryClient.clear()
      navigate('/login')
    },
    onError: (err: { response?: { data?: { error?: string } } }) => {
      setDeleteError(err.response?.data?.error || 'Failed to delete account')
    },
  })

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    updateMutation.mutate(name)
  }

  const handleDelete = (e: React.FormEvent) => {
    e.preventDefault()
    if (confirmEmail !== profile?.email) {
      setDeleteError('Email does not match')
      return
    }
    setDeleteError('')
    deleteMutation.mutate()
  }

  if (isLoading) return <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>

  return (
    <div className="space-y-6">
      <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
        <span className="text-obsidian-accent">&gt;</span> Profile
      </h1>
      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <div className="mb-6 space-y-1 font-mono text-sm">
          <p className="text-obsidian-text-secondary">
            <span className="text-obsidian-text-tertiary">Email:</span>{' '}
            <span className="text-obsidian-text-primary">{profile?.email}</span>
          </p>
          <p className="text-obsidian-text-secondary">
            <span className="text-obsidian-text-tertiary">Verified:</span>{' '}
            <span className={profile?.email_verified ? 'text-obsidian-positive' : 'text-obsidian-negative'}>
              {profile?.email_verified ? 'Yes' : 'No'}
            </span>
          </p>
        </div>

        <form onSubmit={handleSubmit} className="space-y-3">
          <div>
            <label className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              Name
            </label>
            <input
              type="text"
              defaultValue={profile?.name || ''}
              onChange={(e) => setName(e.target.value)}
              className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
            />
          </div>
          <button
            type="submit"
            disabled={updateMutation.isPending}
            className="border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            Update
          </button>
        </form>
      </section>

      <section className="border border-obsidian-negative-dim bg-obsidian-negative-dim/10 p-6">
        <h2 className="mb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-negative">
          Danger Zone
        </h2>
        <p className="mb-4 font-mono text-xs text-obsidian-text-secondary">
          Deleting your account will remove your user record and memberships. Tenants you created will remain but lose their creator reference.
        </p>

        {!showDeleteConfirm ? (
          <button
            onClick={() => setShowDeleteConfirm(true)}
            className="border border-obsidian-negative bg-obsidian-negative/10 px-4 py-2 font-mono text-sm text-obsidian-negative transition-colors hover:bg-obsidian-negative hover:text-white"
          >
            Delete Account
          </button>
        ) : (
          <form onSubmit={handleDelete} className="space-y-3">
            <p className="font-mono text-xs text-obsidian-text-secondary">
              To confirm, type your email: <span className="text-obsidian-text-primary">{profile?.email}</span>
            </p>
            <input
              type="email"
              value={confirmEmail}
              onChange={(e) => setConfirmEmail(e.target.value)}
              placeholder={profile?.email}
              className="w-full border border-obsidian-negative-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-negative"
            />
            {deleteError && (
              <p className="font-mono text-xs text-obsidian-negative">{deleteError}</p>
            )}
            <div className="flex gap-2">
              <button
                type="button"
                onClick={() => {
                  setShowDeleteConfirm(false)
                  setConfirmEmail('')
                  setDeleteError('')
                }}
                className="border border-obsidian-border-dim bg-obsidian-surface px-4 py-2 font-mono text-sm text-obsidian-text-secondary transition-colors hover:border-obsidian-border-med hover:text-obsidian-text-primary"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={deleteMutation.isPending}
                className="border border-obsidian-negative bg-obsidian-negative/10 px-4 py-2 font-mono text-sm text-obsidian-negative transition-colors hover:bg-obsidian-negative hover:text-white disabled:opacity-50"
              >
                {deleteMutation.isPending ? 'Deleting...' : 'Confirm Delete'}
              </button>
            </div>
          </form>
        )}
      </section>
    </div>
  )
}
