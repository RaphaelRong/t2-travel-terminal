import { useState } from 'react'
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import {
  emptyIntegrationForm,
  integrationFormToPayload,
  integrationToForm,
  type IntegrationFormState,
  type Project,
  type ProjectIntegration,
} from '../lib/projectTypes'

interface ProjectSourceManagerProps {
  project: Project
  basePath: string
  editable: boolean
  onChanged: () => void
  showCapabilities?: boolean
}

const kindLabel: Record<ProjectIntegration['kind'], string> = {
  api: 'API Doc',
  mcp: 'HTTP MCP',
  skill: 'Skill Doc',
}

interface HubProvider {
  id: string
  name: string
  type: string
  auth_type: string
  manifest_url: string
  description?: string
  capabilities: string[]
}

interface ProviderCredential {
  provider_id: string
  auth_type: string
  status: string
  configured: boolean
  updated_at: string
}

function SourceForm({
  form,
  setForm,
  submitLabel,
  onSubmit,
  error,
  providers,
}: {
  form: IntegrationFormState
  setForm: (form: IntegrationFormState) => void
  submitLabel: string
  onSubmit: () => void
  error: string
  providers: HubProvider[]
}) {
  const selectProvider = (providerID: string) => {
    const provider = providers.find((item) => item.id === providerID)
    if (!provider) return
    setForm({
      ...form,
      kind: 'skill',
      name: provider.name,
      description: provider.description || provider.name,
      documentation_url: provider.id === 'ticketmaster' ? 'builtin:ticketmaster' : provider.manifest_url,
      endpoint_url: '',
      auth_type: 'inherit',
      metadata: JSON.stringify({ provider_id: provider.id, builtin: true }, null, 2),
    })
  }

  return (
    <div className="grid gap-2 border border-obsidian-border-dim bg-obsidian-surface/70 p-3 md:grid-cols-6">
      <select
        value={form.kind}
        onChange={(event) => setForm({ ...form, kind: event.target.value as IntegrationFormState['kind'] })}
        className="border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary"
      >
        <option value="mcp">HTTP MCP</option>
        <option value="api">API Document</option>
        <option value="skill">Skill Document</option>
      </select>
      <input
        value={form.name}
        onChange={(event) => setForm({ ...form, name: event.target.value })}
        placeholder="Source name"
        className="border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary md:col-span-2"
      />
      <select
        value={form.status}
        onChange={(event) => setForm({ ...form, status: event.target.value as IntegrationFormState['status'] })}
        className="border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary"
      >
        <option value="active">Active</option>
        <option value="inactive">Inactive</option>
      </select>
      <select
        value={form.auth_type}
        onChange={(event) => setForm({ ...form, auth_type: event.target.value as IntegrationFormState['auth_type'] })}
        className="border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary"
      >
        <option value="inherit">Inherit Project Auth</option>
        <option value="none">No Auth</option>
        <option value="api_key">API Key</option>
        <option value="bearer">Bearer</option>
        <option value="oauth2">OAuth2</option>
        <option value="basic">Basic</option>
        <option value="custom">Custom</option>
      </select>
      <button
        type="button"
        onClick={onSubmit}
        className="border border-obsidian-accent bg-obsidian-accent/10 px-2 py-2 font-mono text-xs text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white"
      >
        {submitLabel}
      </button>
      {form.kind === 'skill' && (
        <select
          value=""
          onChange={(event) => selectProvider(event.target.value)}
          className="border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary md:col-span-6"
        >
          <option value="">Select built-in skill...</option>
          {providers.map((provider) => (
            <option key={provider.id} value={provider.id}>
              {provider.name}
            </option>
          ))}
        </select>
      )}
      <input
        value={form.endpoint_url}
        onChange={(event) => setForm({ ...form, endpoint_url: event.target.value })}
        placeholder={form.kind === 'mcp' ? 'HTTP MCP endpoint URL' : 'Runtime endpoint URL'}
        className="border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary md:col-span-3"
      />
      <input
        value={form.documentation_url}
        onChange={(event) => setForm({ ...form, documentation_url: event.target.value })}
        placeholder={form.kind === 'api' ? 'OpenAPI / Swagger spec URL' : form.kind === 'skill' ? 'T2 skill manifest URL' : 'Optional MCP docs URL'}
        className="border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary md:col-span-3"
      />
      <input
        value={form.description}
        onChange={(event) => setForm({ ...form, description: event.target.value })}
        placeholder="Short description"
        className="border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary md:col-span-6"
      />
      <textarea
        value={form.request_headers}
        onChange={(event) => setForm({ ...form, request_headers: event.target.value })}
        placeholder='{"X-API-Key":"..."}'
        className="min-h-20 border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary md:col-span-2"
      />
      <textarea
        value={form.auth_config}
        onChange={(event) => setForm({ ...form, auth_config: event.target.value })}
        placeholder='{"token":"..."}'
        className="min-h-20 border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary md:col-span-2"
      />
      <textarea
        value={form.metadata}
        onChange={(event) => setForm({ ...form, metadata: event.target.value })}
        placeholder='{"format":"openapi"}'
        className="min-h-20 border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary md:col-span-2"
      />
      {error && <p className="font-mono text-xs text-obsidian-negative md:col-span-6">{error}</p>}
    </div>
  )
}

function ProviderAccessManager({
  providers,
  credentials,
  onSaved,
}: {
  providers: HubProvider[]
  credentials: ProviderCredential[]
  onSaved: () => void
}) {
  const [keys, setKeys] = useState<Record<string, string>>({})

  const saveMutation = useMutation({
    mutationFn: (payload: { providerID: string; apiKey: string }) =>
      api.put(`/hub/provider-credentials/${payload.providerID}`, {
        auth_type: 'api_key',
        status: 'active',
        auth_config: { api_key: payload.apiKey },
      }),
    onSuccess: onSaved,
  })

  const credentialByProvider = new Map(credentials.map((item) => [item.provider_id, item]))

  if (providers.length === 0) return null

  return (
    <div className="space-y-2 border border-obsidian-border-dim bg-obsidian-base p-3">
      <p className="font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">Provider Access</p>
      <div className="space-y-2">
        {providers.map((provider) => {
          const credential = credentialByProvider.get(provider.id)
          return (
            <div key={provider.id} className="grid gap-2 border border-obsidian-border-dim bg-obsidian-surface/70 p-2 md:grid-cols-[1fr_220px_auto]">
              <div>
                <p className="font-mono text-xs font-semibold text-obsidian-text-primary">{provider.name}</p>
                <p className="mt-1 font-mono text-[11px] text-obsidian-text-tertiary">
                  {credential?.configured ? `Configured · ${credential.auth_type}` : 'API key not configured'}
                </p>
              </div>
              <input
                value={keys[provider.id] || ''}
                onChange={(event) => setKeys({ ...keys, [provider.id]: event.target.value })}
                placeholder="API key"
                type="password"
                className="border border-obsidian-border-dim bg-obsidian-base px-2 py-2 font-mono text-xs text-obsidian-text-primary"
              />
              <button
                type="button"
                onClick={() => {
                  const apiKey = (keys[provider.id] || '').trim()
                  if (apiKey) saveMutation.mutate({ providerID: provider.id, apiKey })
                }}
                disabled={saveMutation.isPending}
                className="border border-obsidian-accent bg-obsidian-accent/10 px-3 py-2 font-mono text-xs text-obsidian-accent hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
              >
                Save Key
              </button>
            </div>
          )
        })}
      </div>
    </div>
  )
}

export function ProjectSourceManager({ project, basePath, editable, onChanged, showCapabilities = true }: ProjectSourceManagerProps) {
  const queryClient = useQueryClient()
  const [form, setForm] = useState<IntegrationFormState>(emptyIntegrationForm)
  const [editing, setEditing] = useState<ProjectIntegration | null>(null)
  const [formError, setFormError] = useState('')
  const [syncMessage, setSyncMessage] = useState('')

  const createMutation = useMutation({
    mutationFn: (payload: ProjectIntegration) => api.post(`${basePath}/${project.id}/integrations`, payload),
    onSuccess: () => {
      setForm(emptyIntegrationForm)
      setFormError('')
      onChanged()
    },
  })

  const updateMutation = useMutation({
    mutationFn: (payload: { id: string; data: ProjectIntegration }) =>
      api.put(`${basePath}/${project.id}/integrations/${payload.id}`, payload.data),
    onSuccess: () => {
      setEditing(null)
      setForm(emptyIntegrationForm)
      setFormError('')
      onChanged()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`${basePath}/${project.id}/integrations/${id}`),
    onSuccess: onChanged,
  })

  const syncMutation = useMutation({
    mutationFn: (id: string) => api.post(`${basePath}/${project.id}/integrations/${id}/sync`),
    onSuccess: (response) => {
      setSyncMessage(`Loaded ${response.data.count || 0} capabilities`)
      onChanged()
    },
    onError: (error) => {
      const message =
        (error as { response?: { data?: { error?: string } }; message?: string })?.response?.data?.error ||
        (error as { message?: string }).message ||
        'Sync failed'
      setFormError(message)
    },
  })

  const submit = async () => {
    setFormError('')
    setSyncMessage('')
    try {
      const payload = integrationFormToPayload(form)
      if (!payload.name) return
      if (payload.kind === 'mcp' && !payload.endpoint_url) {
        setFormError('HTTP MCP endpoint URL is required')
        return
      }
      if (payload.kind !== 'mcp' && !payload.documentation_url && !payload.endpoint_url) {
        setFormError('A document URL or endpoint URL is required')
        return
      }
      if (editing?.id) {
        await updateMutation.mutateAsync({ id: editing.id, data: payload })
        await syncMutation.mutateAsync(editing.id)
      } else {
        const response = await createMutation.mutateAsync(payload)
        const id = response.data.id as string | undefined
        if (id) {
          await syncMutation.mutateAsync(id)
        }
      }
    } catch (error) {
      const message =
        (error as { response?: { data?: { error?: string } }; message?: string })?.response?.data?.error ||
        (error as { message?: string }).message
      setFormError(message || 'Source save failed')
    }
  }

  const integrations = project.integrations || []
  const capabilities = project.capabilities || []

  const { data: providers = [] } = useQuery({
    queryKey: ['hub-providers'],
    queryFn: async () => {
      const res = await api.get<{ providers: HubProvider[] }>('/hub/providers')
      return res.data.providers
    },
  })

  const { data: credentials = [] } = useQuery({
    queryKey: ['hub-provider-credentials'],
    queryFn: async () => {
      const res = await api.get<{ credentials: ProviderCredential[] }>('/hub/provider-credentials')
      return res.data.credentials
    },
    retry: false,
  })

  const refreshCredentials = () => {
    queryClient.invalidateQueries({ queryKey: ['hub-provider-credentials'] })
  }

  return (
    <div className="mt-4 space-y-3 border-t border-obsidian-border-dim pt-3">
      <div className="flex items-center justify-between gap-3">
        <p className="font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">Sources</p>
        {editing && (
          <button
            type="button"
            onClick={() => {
              setEditing(null)
              setForm(emptyIntegrationForm)
              setFormError('')
            }}
            className="font-mono text-xs text-obsidian-text-tertiary hover:text-obsidian-text-primary"
          >
            Cancel source edit
          </button>
        )}
      </div>

      {integrations.length === 0 ? (
        <p className="font-mono text-xs text-obsidian-text-tertiary">No API, MCP, or Skill source configured.</p>
      ) : (
        <div className="space-y-2">
          {integrations.map((integration) => (
            <div key={integration.id || integration.name} className="border border-obsidian-border-dim bg-obsidian-base p-3">
              <div className="flex items-start justify-between gap-3">
                <div className="min-w-0">
                  <p className="font-mono text-xs font-semibold text-obsidian-text-primary">
                    {kindLabel[integration.kind]} · {integration.name}
                  </p>
                  <p className="mt-1 break-all font-mono text-[11px] text-obsidian-text-secondary">
                    {integration.endpoint_url || integration.documentation_url || 'No URL'}
                  </p>
                  <p className="mt-1 font-mono text-[10px] uppercase text-obsidian-text-tertiary">
                    {integration.status || 'active'} · {integration.auth_type || 'inherit'} · sync {integration.sync_status || 'idle'}
                  </p>
                  {integration.sync_error && (
                    <p className="mt-1 font-mono text-[11px] text-obsidian-negative">{integration.sync_error}</p>
                  )}
                  {syncMessage && integration.id === editing?.id && (
                    <p className="mt-1 font-mono text-[11px] text-obsidian-positive">{syncMessage}</p>
                  )}
                </div>
                {editable && (
                  <div className="flex shrink-0 flex-wrap justify-end gap-2">
                    {integration.id && (
                      <button
                        type="button"
                        onClick={() => syncMutation.mutate(integration.id!)}
                        disabled={syncMutation.isPending}
                        className="font-mono text-xs text-obsidian-positive hover:underline disabled:opacity-50"
                      >
                        {syncMutation.isPending ? 'Syncing...' : 'Sync'}
                      </button>
                    )}
                    <button
                      type="button"
                      onClick={() => {
                        setEditing(integration)
                        setForm(integrationToForm(integration))
                      }}
                      className="font-mono text-xs text-obsidian-accent hover:underline"
                    >
                      Edit
                    </button>
                    {integration.id && (
                      <button
                        type="button"
                        onClick={() => deleteMutation.mutate(integration.id!)}
                        className="font-mono text-xs text-obsidian-negative hover:underline"
                      >
                        Delete
                      </button>
                    )}
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      )}

      {editable && (
        <>
          <ProviderAccessManager providers={providers} credentials={credentials} onSaved={refreshCredentials} />
          <SourceForm
            form={form}
            setForm={setForm}
            submitLabel={editing ? 'Update Source' : 'Add Source'}
            onSubmit={submit}
            error={formError}
            providers={providers}
          />
        </>
      )}
      {syncMessage && !editing && <p className="font-mono text-xs text-obsidian-positive">{syncMessage}</p>}

      {showCapabilities && (
        <div className="space-y-2">
          <p className="font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">Loaded Capabilities</p>
          {capabilities.length === 0 ? (
            <p className="font-mono text-xs text-obsidian-text-tertiary">No capabilities loaded yet.</p>
          ) : (
            <div className="flex flex-wrap gap-2">
              {capabilities.map((item) => (
                <span
                  key={item.id || item.name}
                  className="rounded border border-obsidian-border-dim px-2 py-1 font-mono text-[10px] text-obsidian-text-secondary"
                >
                  {item.kind}:{item.name}
                </span>
              ))}
            </div>
          )}
        </div>
      )}
    </div>
  )
}
