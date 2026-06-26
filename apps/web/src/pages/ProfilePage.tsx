import { useEffect, useState } from 'react'
import { useNavigate } from 'react-router-dom'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { removeToken } from '../lib/auth'
import {
  emptyLLMProfileForm,
  formToLLMProfilePayload,
  profileToForm,
  providerLabels,
  providerModelHints,
  type LLMProfile,
  type LLMProfileFormState,
  type LLMProvider,
} from '../lib/llmProfiles'

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

  const [name, setName] = useState('')
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false)
  const [confirmEmail, setConfirmEmail] = useState('')
  const [deleteError, setDeleteError] = useState('')

  useEffect(() => {
    setName(profile?.name || '')
  }, [profile?.name])

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
              value={name}
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

      <LLMProfilesSection />

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

function LLMProfilesSection() {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<LLMProfileFormState>(emptyLLMProfileForm)
  const [formError, setFormError] = useState('')
  const [isFetchingModels, setIsFetchingModels] = useState(false)
  const [fetchError, setFetchError] = useState('')

  const { data: profiles = [], isLoading } = useQuery({
    queryKey: ['llm-profiles'],
    queryFn: async () => {
      const res = await api.get<{ profiles: LLMProfile[] }>('/llm-profiles')
      return res.data.profiles
    },
  })

  const saveMutation = useMutation({
    mutationFn: (payload: LLMProfileFormState) => {
      const body = formToLLMProfilePayload(payload)
      if (payload.id) {
        return api.put(`/llm-profiles/${payload.id}`, body)
      }
      return api.post('/llm-profiles', body)
    },
    onSuccess: () => {
      setForm(emptyLLMProfileForm)
      setFormError('')
      queryClient.invalidateQueries({ queryKey: ['llm-profiles'] })
    },
    onError: (err: { response?: { data?: { error?: string } }; message?: string }) => {
      setFormError(err.response?.data?.error || err.message || 'Failed to save LLM profile')
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/llm-profiles/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm-profiles'] })
      setForm((current) => (current.id ? emptyLLMProfileForm : current))
    },
  })

  const handleFetchModels = async () => {
    if (!form.api_key.trim()) {
      setFetchError('API key is required')
      return
    }
    if (
      !form.base_url.trim() &&
      form.provider !== 'anthropic' &&
      form.provider !== 'google'
    ) {
      setFetchError('Endpoint base URL is required for this provider')
      return
    }
    setIsFetchingModels(true)
    setFetchError('')
    try {
      const res = await api.post<{ models: string[] }>('/llm-profiles/fetch-models', {
        provider: form.provider,
        base_url: form.base_url,
        api_key: form.api_key,
      })
      const models = res.data.models
      if (models.length === 0) {
        setFetchError('No models returned from provider')
        return
      }
      updateForm({
        models: models.join('\n'),
        default_model: form.default_model || models[0],
      })
    } catch (err: any) {
      setFetchError(err.response?.data?.error || err.message || 'Failed to fetch models')
    } finally {
      setIsFetchingModels(false)
    }
  }

  const updateForm = (patch: Partial<LLMProfileFormState>) => {
    setForm((current) => ({ ...current, ...patch }))
  }

  const handleProviderChange = (provider: LLMProvider) => {
    const hints = providerModelHints[provider]
    updateForm({
      provider,
      models: form.models || hints.join('\n'),
      default_model: form.default_model || hints[0] || '',
    })
  }

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault()
    if (!form.display_name.trim()) {
      setFormError('Display name is required')
      return
    }
    saveMutation.mutate(form)
  }

  return (
    <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
      <div className="mb-5 flex flex-col justify-between gap-3 md:flex-row md:items-start">
        <div>
          <h2 className="font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
            LLM Provider Profiles
          </h2>
          <p className="mt-2 max-w-3xl font-mono text-sm leading-6 text-obsidian-text-secondary">
            Configure the model providers available in Playground. API keys are stored server-side and are never returned to the browser.
          </p>
        </div>
        <button
          type="button"
          onClick={() => {
            setForm(emptyLLMProfileForm)
            setFormError('')
          }}
          className="border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-xs text-obsidian-text-secondary hover:border-obsidian-accent hover:text-obsidian-accent"
        >
          New Profile
        </button>
      </div>

      <div className="grid gap-5 lg:grid-cols-[minmax(0,1fr)_380px]">
        <div className="space-y-3">
          {isLoading && <p className="font-mono text-sm text-obsidian-text-secondary">Loading profiles...</p>}
          {!isLoading && profiles.length === 0 && (
            <div className="border border-dashed border-obsidian-border-dim bg-obsidian-base p-5">
              <p className="font-mono text-sm text-obsidian-text-secondary">
                No provider profiles yet. Add OpenAI, Anthropic, Google, or a custom compatible endpoint.
              </p>
            </div>
          )}
          {profiles.map((profile) => (
            <article
              key={profile.id}
              className="border border-obsidian-border-dim bg-obsidian-base p-4"
            >
              <div className="flex flex-col gap-3 md:flex-row md:items-start md:justify-between">
                <div>
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="font-mono text-sm font-semibold text-obsidian-text-primary">
                      {profile.display_name}
                    </h3>
                    <span className="border border-obsidian-border-dim px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider text-obsidian-text-tertiary">
                      {providerLabels[profile.provider]}
                    </span>
                    <span className={profile.status === 'active' ? 'font-mono text-[10px] uppercase tracking-wider text-obsidian-positive' : 'font-mono text-[10px] uppercase tracking-wider text-obsidian-text-tertiary'}>
                      {profile.status}
                    </span>
                  </div>
                  <p className="mt-2 font-mono text-xs text-obsidian-text-secondary">
                    Default model: <span className="text-obsidian-text-primary">{profile.default_model || '-'}</span>
                  </p>
                  <p className="mt-1 font-mono text-xs text-obsidian-text-secondary">
                    Models: {profile.models.length ? profile.models.join(', ') : '-'}
                  </p>
                  <p className="mt-1 font-mono text-xs text-obsidian-text-secondary">
                    Endpoint: {profile.base_url || 'Provider default'} · Key: {profile.configured ? 'configured' : 'missing'}
                  </p>
                </div>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => {
                      setForm(profileToForm(profile))
                      setFormError('')
                    }}
                    className="border border-obsidian-border-dim px-3 py-1.5 font-mono text-xs text-obsidian-text-secondary hover:border-obsidian-accent hover:text-obsidian-accent"
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    onClick={() => deleteMutation.mutate(profile.id)}
                    className="border border-obsidian-negative-dim px-3 py-1.5 font-mono text-xs text-obsidian-negative hover:border-obsidian-negative"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </article>
          ))}
        </div>

        <form onSubmit={handleSubmit} className="space-y-4 border border-obsidian-border-dim bg-obsidian-base p-4">
          <h3 className="font-mono text-sm font-semibold text-obsidian-text-primary">
            {form.id ? 'Edit Provider Profile' : 'Add Provider Profile'}
          </h3>

          <div className="grid gap-3 sm:grid-cols-2 lg:grid-cols-1">
            <label className="block">
              <span className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
                Provider
              </span>
              <select
                value={form.provider}
                onChange={(e) => handleProviderChange(e.target.value as LLMProvider)}
                className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
              >
                {Object.entries(providerLabels).map(([value, label]) => (
                  <option key={value} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </label>

            <label className="block">
              <span className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
                Status
              </span>
              <select
                value={form.status}
                onChange={(e) => updateForm({ status: e.target.value as LLMProfileFormState['status'] })}
                className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
              </select>
            </label>
          </div>

          <label className="block">
            <span className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              Display Name
            </span>
            <input
              value={form.display_name}
              onChange={(e) => updateForm({ display_name: e.target.value })}
              placeholder="Production OpenAI"
              className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
            />
          </label>

          <label className="block">
            <span className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              Endpoint Base URL
            </span>
            <input
              value={form.base_url}
              onChange={(e) => updateForm({ base_url: e.target.value })}
              placeholder="https://api.openai.com/v1"
              className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
            />
          </label>

          <label className="block">
            <span className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              API Key
            </span>
            <input
              type="password"
              value={form.api_key}
              onChange={(e) => updateForm({ api_key: e.target.value })}
              placeholder={form.id ? 'Leave blank to keep existing key' : 'sk-...'}
              className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
            />
          </label>

          <label className="block">
            <span className="mb-1 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              Default Model
            </span>
            <input
              value={form.default_model}
              onChange={(e) => updateForm({ default_model: e.target.value })}
              placeholder="gpt-4.1-mini"
              className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
            />
          </label>

          <label className="block">
            <div className="mb-1 flex items-center justify-between">
              <span className="font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
                Available Models
              </span>
              <button
                type="button"
                onClick={handleFetchModels}
                disabled={isFetchingModels}
                className="border border-obsidian-border-dim bg-obsidian-base px-2 py-1 font-mono text-[10px] uppercase tracking-wider text-obsidian-text-secondary transition-colors hover:border-obsidian-accent hover:text-obsidian-accent disabled:opacity-50"
              >
                {isFetchingModels ? 'Fetching...' : 'Fetch Models'}
              </button>
            </div>
            {fetchError && (
              <p className="mb-1 font-mono text-xs text-obsidian-negative">{fetchError}</p>
            )}
            <textarea
              value={form.models}
              onChange={(e) => updateForm({ models: e.target.value })}
              placeholder="One model per line, or comma separated"
              rows={4}
              className="w-full resize-none border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
            />
          </label>

          {formError && <p className="font-mono text-xs text-obsidian-negative">{formError}</p>}

          <button
            type="submit"
            disabled={saveMutation.isPending}
            className="w-full border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            {saveMutation.isPending ? 'Saving...' : form.id ? 'Save Changes' : 'Create Profile'}
          </button>
        </form>
      </div>
    </section>
  )
}
