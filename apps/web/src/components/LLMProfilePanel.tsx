import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { AxiosError } from 'axios'
import { api } from '../lib/api'
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

type DrawerMode = 'create' | 'edit' | null

interface LLMProfilePanelProps {
  onChanged?: () => void
}

function statusClasses(status: LLMProfile['status']) {
  return status === 'active'
    ? 'border-obsidian-positive-dim bg-obsidian-positive-dim/20 text-obsidian-positive'
    : 'border-obsidian-warning/40 bg-obsidian-warning/10 text-obsidian-warning'
}

function errorMessage(err: unknown, fallback: string) {
  const axiosError = err as AxiosError<{ error?: string }>
  return axiosError.response?.data?.error || (err instanceof Error ? err.message : fallback)
}

function LLMProfileDrawer({
  mode,
  form,
  setForm,
  formError,
  fetchError,
  isFetchingModels,
  isPending,
  onClose,
  onSubmit,
  onFetchModels,
}: {
  mode: Exclude<DrawerMode, null>
  form: LLMProfileFormState
  setForm: (form: LLMProfileFormState) => void
  formError: string
  fetchError: string
  isFetchingModels: boolean
  isPending: boolean
  onClose: () => void
  onSubmit: (event: React.FormEvent) => void
  onFetchModels: () => void
}) {
  const updateForm = (patch: Partial<LLMProfileFormState>) => {
    setForm({ ...form, ...patch })
  }

  const handleProviderChange = (provider: LLMProvider) => {
    const hints = providerModelHints[provider]
    updateForm({
      provider,
      models: form.models || hints.join('\n'),
      default_model: form.default_model || hints[0] || '',
    })
  }

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-black/60">
      <button
        type="button"
        aria-label="Close profile panel"
        className="hidden flex-1 cursor-default md:block"
        onClick={onClose}
      />
      <aside className="h-full w-full max-w-2xl overflow-y-auto border-l border-obsidian-border-dim bg-obsidian-surface shadow-2xl">
        <header className="sticky top-0 z-10 border-b border-obsidian-border-dim bg-obsidian-raised/95 px-5 py-4 backdrop-blur">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="font-mono text-xs uppercase tracking-wide text-obsidian-accent">
                {mode === 'create' ? 'New Profile' : 'Edit Profile'}
              </p>
              <h2 className="mt-1 font-mono text-xl font-semibold text-obsidian-text-primary">
                {mode === 'create' ? 'Create LLM provider profile' : form.display_name || 'Edit profile'}
              </h2>
            </div>
            <button
              type="button"
              onClick={onClose}
              className="border border-obsidian-border-dim bg-obsidian-base px-3 py-1.5 font-mono text-xs text-obsidian-text-secondary hover:border-obsidian-border-med hover:text-obsidian-text-primary"
            >
              Close
            </button>
          </div>
        </header>

        <form onSubmit={onSubmit} className="space-y-5 p-5">
          <section className="space-y-4 border border-obsidian-border-dim bg-obsidian-base p-4">
            <div>
              <p className="font-mono text-xs uppercase tracking-wide text-obsidian-accent">Basics</p>
              <p className="mt-1 font-mono text-xs text-obsidian-text-tertiary">
                API keys are stored server-side and are never returned to the browser.
              </p>
            </div>

            <div className="grid gap-4 sm:grid-cols-2">
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
          </section>

          <section className="space-y-4 border border-obsidian-border-dim bg-obsidian-base p-4">
            <p className="font-mono text-xs uppercase tracking-wide text-obsidian-accent">Connection</p>

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
          </section>

          <section className="space-y-4 border border-obsidian-border-dim bg-obsidian-base p-4">
            <p className="font-mono text-xs uppercase tracking-wide text-obsidian-accent">Models</p>

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
                  onClick={onFetchModels}
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
                rows={5}
                className="w-full resize-none border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
              />
            </label>
          </section>

          {formError && (
            <p className="border border-obsidian-negative-dim bg-obsidian-negative-dim/20 p-2 font-mono text-xs text-obsidian-negative">
              {formError}
            </p>
          )}

          <button
            type="submit"
            disabled={isPending}
            className="border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            {isPending ? 'Saving...' : mode === 'create' ? 'Create Profile' : 'Save Changes'}
          </button>
        </form>
      </aside>
    </div>
  )
}

export function LLMProfilePanel({ onChanged }: LLMProfilePanelProps) {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<LLMProfileFormState>(emptyLLMProfileForm)
  const [formError, setFormError] = useState('')
  const [fetchError, setFetchError] = useState('')
  const [isFetchingModels, setIsFetchingModels] = useState(false)
  const [drawerMode, setDrawerMode] = useState<DrawerMode>(null)

  const { data: profiles = [], isLoading } = useQuery({
    queryKey: ['llm-profiles'],
    queryFn: async () => {
      const res = await api.get<{ profiles: LLMProfile[] }>('/llm-profiles')
      return res.data.profiles
    },
  })

  const saveMutation = useMutation({
    mutationFn: (payload: { id?: string; body: ReturnType<typeof formToLLMProfilePayload> }) => {
      if (payload.id) {
        return api.put(`/llm-profiles/${payload.id}`, payload.body)
      }
      return api.post('/llm-profiles', payload.body)
    },
    onSuccess: () => {
      closeDrawer()
      queryClient.invalidateQueries({ queryKey: ['llm-profiles'] })
      onChanged?.()
    },
    onError: (err: unknown) => {
      setFormError(errorMessage(err, 'Failed to save LLM profile'))
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/llm-profiles/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['llm-profiles'] })
      onChanged?.()
    },
  })

  const closeDrawer = () => {
    setDrawerMode(null)
    setForm(emptyLLMProfileForm)
    setFormError('')
    setFetchError('')
  }

  const openCreate = () => {
    setForm(emptyLLMProfileForm)
    setFormError('')
    setFetchError('')
    setDrawerMode('create')
  }

  const openEdit = (profile: LLMProfile) => {
    setForm(profileToForm(profile))
    setFormError('')
    setFetchError('')
    setDrawerMode('edit')
  }

  const handleFetchModels = async () => {
    if (!form.api_key.trim()) {
      setFetchError('API key is required')
      return
    }
    if (!form.base_url.trim() && form.provider !== 'anthropic' && form.provider !== 'google') {
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
      setForm((current) => ({
        ...current,
        models: models.join('\n'),
        default_model: current.default_model || models[0],
      }))
    } catch (err: unknown) {
      setFetchError(errorMessage(err, 'Failed to fetch models'))
    } finally {
      setIsFetchingModels(false)
    }
  }

  const handleSubmit = (event: React.FormEvent) => {
    event.preventDefault()
    if (!form.display_name.trim()) {
      setFormError('Display name is required')
      return
    }
    setFormError('')
    saveMutation.mutate({ id: form.id, body: formToLLMProfilePayload(form) })
  }

  return (
    <section className="space-y-4 border border-obsidian-border-dim bg-obsidian-surface p-5">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h2 className="font-mono text-lg font-semibold text-obsidian-text-primary">LLM Provider Profiles</h2>
          <p className="mt-1 max-w-2xl font-mono text-sm text-obsidian-text-secondary">
            Configure the model providers available in Playground. API keys are stored server-side and are never returned to the browser.
          </p>
        </div>
        <button
          type="button"
          onClick={openCreate}
          className="w-fit border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white"
        >
          New Profile
        </button>
      </div>

      {isLoading ? (
        <p className="font-mono text-sm text-obsidian-text-secondary">Loading profiles...</p>
      ) : profiles.length === 0 ? (
        <p className="border border-obsidian-border-dim bg-obsidian-base p-4 font-mono text-sm text-obsidian-text-secondary">
          No provider profiles yet. Add OpenAI, Anthropic, Google, or a custom compatible endpoint.
        </p>
      ) : (
        <div className="grid gap-3 xl:grid-cols-2">
          {profiles.map((profile) => (
            <article
              key={profile.id}
              className="border border-obsidian-border-dim bg-obsidian-base p-4 transition-colors hover:border-obsidian-border-med"
            >
              <div className="flex items-start justify-between gap-4">
                <div className="min-w-0 flex-1">
                  <div className="flex flex-wrap items-center gap-2">
                    <h3 className="truncate font-mono text-sm font-semibold text-obsidian-text-primary">
                      {profile.display_name}
                    </h3>
                    <span className="border border-obsidian-border-dim px-2 py-0.5 font-mono text-[10px] uppercase tracking-wider text-obsidian-text-tertiary">
                      {providerLabels[profile.provider]}
                    </span>
                    <span className={`border px-2 py-0.5 font-mono text-[10px] uppercase ${statusClasses(profile.status)}`}>
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
                <div className="flex shrink-0 flex-col items-end gap-2">
                  <button
                    type="button"
                    onClick={() => openEdit(profile)}
                    className="border border-obsidian-border-dim px-3 py-1.5 font-mono text-xs text-obsidian-text-secondary hover:border-obsidian-accent hover:text-obsidian-accent"
                  >
                    Edit
                  </button>
                  <button
                    type="button"
                    onClick={() => {
                      if (window.confirm(`Delete profile "${profile.display_name}"?`)) {
                        deleteMutation.mutate(profile.id)
                      }
                    }}
                    className="font-mono text-xs text-obsidian-negative hover:underline"
                  >
                    Delete
                  </button>
                </div>
              </div>
            </article>
          ))}
        </div>
      )}

      {drawerMode && (
        <LLMProfileDrawer
          mode={drawerMode}
          form={form}
          setForm={setForm}
          formError={formError}
          fetchError={fetchError}
          isFetchingModels={isFetchingModels}
          isPending={saveMutation.isPending}
          onClose={closeDrawer}
          onSubmit={handleSubmit}
          onFetchModels={handleFetchModels}
        />
      )}
    </section>
  )
}
