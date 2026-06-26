import { useMemo, useState } from 'react'
import { useMutation } from '@tanstack/react-query'
import { api } from '../lib/api'
import { ProjectSourceManager } from './ProjectSourceManager'
import {
  emptyProjectForm,
  formToPayload,
  projectToForm,
  type Project,
  type ProjectCapability,
  type ProjectFormState,
  type ProjectStatus,
} from '../lib/projectTypes'

type DrawerMode = 'create' | 'view' | 'edit'
type DetailTab = 'overview' | 'sources' | 'capabilities' | 'advanced'
type StatusFilter = 'all' | 'online' | 'offline'

interface ProjectManagementPanelProps {
  title: string
  subtitle: string
  projects: Project[]
  basePath: string
  isLoading: boolean
  editable: boolean
  creatable: boolean
  emptyText: string
  onChanged: () => void
  error?: string
}

const statusClasses: Record<ProjectStatus, string> = {
  online: 'border-obsidian-positive-dim bg-obsidian-positive-dim/20 text-obsidian-positive',
  offline: 'border-obsidian-warning/40 bg-obsidian-warning/10 text-obsidian-warning',
  draft: 'border-obsidian-border-med bg-obsidian-highlight text-obsidian-text-secondary',
  archived: 'border-obsidian-negative-dim bg-obsidian-negative-dim/20 text-obsidian-negative',
}

function blankProjectForm(): ProjectFormState {
  return { ...emptyProjectForm, status: 'offline', source_type: 'mixed' }
}

function sourceCounts(project: Project) {
  const counts = { api: 0, mcp: 0, skill: 0 }
  ;(project.integrations || []).forEach((source) => {
    counts[source.kind] += 1
  })
  return counts
}

function latestSync(project: Project) {
  const dates = (project.integrations || [])
    .map((source) => source.last_synced_at)
    .filter(Boolean)
    .map((value) => new Date(value as string))
    .filter((date) => !Number.isNaN(date.getTime()))
    .sort((a, b) => b.getTime() - a.getTime())
  return dates[0]
}

function ProjectForm({
  form,
  setForm,
  onSubmit,
  submitLabel,
  error,
  isPending,
}: {
  form: ProjectFormState
  setForm: (form: ProjectFormState) => void
  onSubmit: (event: React.FormEvent) => void
  submitLabel: string
  error: string
  isPending: boolean
}) {
  const [advancedOpen, setAdvancedOpen] = useState(false)

  return (
    <form onSubmit={onSubmit} className="space-y-5">
      <section className="space-y-3 border border-obsidian-border-dim bg-obsidian-base p-4">
        <div>
          <p className="font-mono text-xs uppercase tracking-wide text-obsidian-accent">Basics</p>
          <p className="mt-1 font-mono text-xs text-obsidian-text-tertiary">Project is created offline. Bring it online from the card after configuration.</p>
        </div>
        <label className="block">
          <span className="mb-1 block font-mono text-xs uppercase text-obsidian-text-secondary">Name</span>
          <input
            value={form.name}
            onChange={(event) => setForm({ ...form, name: event.target.value })}
            placeholder="Finance data cloud"
            className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          />
        </label>
        <label className="block">
          <span className="mb-1 block font-mono text-xs uppercase text-obsidian-text-secondary">Description</span>
          <textarea
            value={form.description}
            onChange={(event) => setForm({ ...form, description: event.target.value })}
            placeholder="What this project provides to agents"
            className="min-h-24 w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          />
        </label>
        <label className="block">
          <span className="mb-1 block font-mono text-xs uppercase text-obsidian-text-secondary">Capability Summary</span>
          <input
            value={form.capability_summary}
            onChange={(event) => setForm({ ...form, capability_summary: event.target.value })}
            placeholder="Short summary for Agent selection"
            className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          />
        </label>
      </section>

      <section className="space-y-3 border border-obsidian-border-dim bg-obsidian-base p-4">
        <div>
          <p className="font-mono text-xs uppercase tracking-wide text-obsidian-accent">Connection Defaults</p>
          <p className="mt-1 font-mono text-xs text-obsidian-text-tertiary">Sources can inherit these settings or override them individually.</p>
        </div>
        <label className="block">
          <span className="mb-1 block font-mono text-xs uppercase text-obsidian-text-secondary">Unified Endpoint URL</span>
          <input
            value={form.endpoint_url}
            onChange={(event) => setForm({ ...form, endpoint_url: event.target.value })}
            placeholder="https://api.example.com"
            className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
          />
        </label>
      </section>

      <section className="border border-obsidian-border-dim bg-obsidian-base">
        <button
          type="button"
          onClick={() => setAdvancedOpen((value) => !value)}
          className="flex w-full items-center justify-between px-4 py-3 text-left font-mono text-xs uppercase tracking-wide text-obsidian-accent"
        >
          Advanced Auth
          <span className="text-obsidian-text-tertiary">{advancedOpen ? 'Hide' : 'Show'}</span>
        </button>
        {advancedOpen && (
          <div className="space-y-3 border-t border-obsidian-border-dim p-4">
            <select
              value={form.auth_type}
              onChange={(event) => setForm({ ...form, auth_type: event.target.value as ProjectFormState['auth_type'] })}
              className="w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
            >
              <option value="none">No Auth</option>
              <option value="api_key">API Key</option>
              <option value="bearer">Bearer</option>
              <option value="oauth2">OAuth2</option>
              <option value="basic">Basic</option>
              <option value="custom">Custom</option>
            </select>
            <div className="grid gap-3 md:grid-cols-2">
              <label className="block">
                <span className="mb-1 block font-mono text-xs uppercase text-obsidian-text-secondary">Request Headers</span>
                <textarea
                  value={form.request_headers}
                  onChange={(event) => setForm({ ...form, request_headers: event.target.value })}
                  className="min-h-28 w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-xs text-obsidian-text-primary outline-none focus:border-obsidian-accent"
                />
              </label>
              <label className="block">
                <span className="mb-1 block font-mono text-xs uppercase text-obsidian-text-secondary">Auth Config</span>
                <textarea
                  value={form.auth_config}
                  onChange={(event) => setForm({ ...form, auth_config: event.target.value })}
                  className="min-h-28 w-full border border-obsidian-border-dim bg-obsidian-surface px-3 py-2 font-mono text-xs text-obsidian-text-primary outline-none focus:border-obsidian-accent"
                />
              </label>
            </div>
          </div>
        )}
      </section>

      {error && <p className="border border-obsidian-negative-dim bg-obsidian-negative-dim/20 p-2 font-mono text-xs text-obsidian-negative">{error}</p>}

      <button
        type="submit"
        disabled={isPending}
        className="border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
      >
        {submitLabel}
      </button>
    </form>
  )
}

function CapabilitiesTable({ project }: { project: Project }) {
  const capabilities = project.capabilities || []
  const sourceByID = new Map((project.integrations || []).map((source) => [source.id, source.name]))

  if (capabilities.length === 0) {
    return <p className="font-mono text-sm text-obsidian-text-secondary">No capabilities loaded yet.</p>
  }

  return (
    <div className="overflow-x-auto border border-obsidian-border-dim">
      <table className="w-full min-w-[680px] border-collapse text-left font-mono text-xs">
        <thead className="bg-obsidian-base text-obsidian-text-tertiary">
          <tr>
            <th className="border-b border-obsidian-border-dim px-3 py-2 font-medium uppercase">Name</th>
            <th className="border-b border-obsidian-border-dim px-3 py-2 font-medium uppercase">Kind</th>
            <th className="border-b border-obsidian-border-dim px-3 py-2 font-medium uppercase">Source</th>
            <th className="border-b border-obsidian-border-dim px-3 py-2 font-medium uppercase">Status</th>
            <th className="border-b border-obsidian-border-dim px-3 py-2 font-medium uppercase">Description</th>
          </tr>
        </thead>
        <tbody>
          {capabilities.map((capability: ProjectCapability) => (
            <tr key={capability.id || capability.name} className="border-b border-obsidian-border-dim last:border-b-0">
              <td className="px-3 py-2 text-obsidian-text-primary">{capability.name}</td>
              <td className="px-3 py-2 uppercase text-obsidian-accent">{capability.kind}</td>
              <td className="px-3 py-2 text-obsidian-text-secondary">
                {capability.integration_id ? sourceByID.get(capability.integration_id) || 'Source' : 'Manual'}
              </td>
              <td className="px-3 py-2 text-obsidian-text-secondary">{capability.status || 'active'}</td>
              <td className="max-w-sm px-3 py-2 text-obsidian-text-tertiary">{capability.description || '-'}</td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  )
}

function ProjectDrawer({
  project,
  mode,
  editable,
  basePath,
  form,
  setForm,
  formError,
  activeTab,
  setActiveTab,
  isPending,
  onClose,
  onEdit,
  onSubmit,
  onChanged,
}: {
  project: Project | null
  mode: DrawerMode
  editable: boolean
  basePath: string
  form: ProjectFormState
  setForm: (form: ProjectFormState) => void
  formError: string
  activeTab: DetailTab
  setActiveTab: (tab: DetailTab) => void
  isPending: boolean
  onClose: () => void
  onEdit: () => void
  onSubmit: (event: React.FormEvent) => void
  onChanged: () => void
}) {
  const isForm = mode === 'create' || mode === 'edit'

  return (
    <div className="fixed inset-0 z-50 flex justify-end bg-black/60">
      <button type="button" aria-label="Close project panel" className="hidden flex-1 cursor-default md:block" onClick={onClose} />
      <aside className="h-full w-full max-w-3xl overflow-y-auto border-l border-obsidian-border-dim bg-obsidian-surface shadow-2xl">
        <header className="sticky top-0 z-10 border-b border-obsidian-border-dim bg-obsidian-raised/95 px-5 py-4 backdrop-blur">
          <div className="flex items-start justify-between gap-4">
            <div>
              <p className="font-mono text-xs uppercase tracking-wide text-obsidian-accent">
                {mode === 'create' ? 'New Project' : mode === 'edit' ? 'Edit Project' : 'Project Detail'}
              </p>
              <h2 className="mt-1 font-mono text-xl font-semibold text-obsidian-text-primary">
                {mode === 'create' ? 'Create project' : project?.name}
              </h2>
              {project && (
                <p className="mt-1 font-mono text-xs uppercase text-obsidian-text-tertiary">
                  {project.source_scope} · {project.status} · {project.auth_type}
                </p>
              )}
            </div>
            <div className="flex gap-2">
              {!isForm && editable && (
                <button
                  type="button"
                  onClick={onEdit}
                  className="border border-obsidian-accent bg-obsidian-accent/10 px-3 py-1.5 font-mono text-xs text-obsidian-accent hover:bg-obsidian-accent hover:text-white"
                >
                  Edit
                </button>
              )}
              <button
                type="button"
                onClick={onClose}
                className="border border-obsidian-border-dim bg-obsidian-base px-3 py-1.5 font-mono text-xs text-obsidian-text-secondary hover:border-obsidian-border-med hover:text-obsidian-text-primary"
              >
                Close
              </button>
            </div>
          </div>
        </header>

        <div className="p-5">
          {isForm ? (
            <ProjectForm
              form={form}
              setForm={setForm}
              onSubmit={onSubmit}
              submitLabel={mode === 'create' ? 'Create Project' : 'Save Changes'}
              error={formError}
              isPending={isPending}
            />
          ) : (
            project && (
              <div className="space-y-5">
                <nav className="flex flex-wrap gap-2 border-b border-obsidian-border-dim pb-3">
                  {(['overview', 'sources', 'capabilities', 'advanced'] as DetailTab[]).map((tab) => (
                    <button
                      key={tab}
                      type="button"
                      onClick={() => setActiveTab(tab)}
                      className={`border px-3 py-1.5 font-mono text-xs uppercase ${
                        activeTab === tab
                          ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-accent'
                          : 'border-obsidian-border-dim bg-obsidian-base text-obsidian-text-secondary hover:border-obsidian-border-med'
                      }`}
                    >
                      {tab}
                    </button>
                  ))}
                </nav>

                {activeTab === 'overview' && (
                  <div className="space-y-4">
                    <section className="border border-obsidian-border-dim bg-obsidian-base p-4">
                      <p className="font-mono text-xs uppercase text-obsidian-text-tertiary">Description</p>
                      <p className="mt-2 font-mono text-sm leading-6 text-obsidian-text-secondary">{project.description || '-'}</p>
                    </section>
                    <section className="border border-obsidian-border-dim bg-obsidian-base p-4">
                      <p className="font-mono text-xs uppercase text-obsidian-text-tertiary">Agent Summary</p>
                      <p className="mt-2 font-mono text-sm leading-6 text-obsidian-text-secondary">{project.capability_summary || '-'}</p>
                    </section>
                    <div className="grid gap-3 md:grid-cols-3">
                      <Stat label="Sources" value={`${project.integrations?.length || 0}`} />
                      <Stat label="Capabilities" value={`${project.capabilities?.length || 0}`} />
                      <Stat label="Endpoint" value={project.endpoint_url ? 'Configured' : 'None'} />
                    </div>
                  </div>
                )}

                {activeTab === 'sources' && (
                  <ProjectSourceManager
                    project={project}
                    basePath={basePath}
                    editable={editable}
                    onChanged={onChanged}
                    showCapabilities={false}
                  />
                )}

                {activeTab === 'capabilities' && <CapabilitiesTable project={project} />}

                {activeTab === 'advanced' && (
                  <div className="space-y-3">
                    <section className="border border-obsidian-border-dim bg-obsidian-base p-4">
                      <p className="font-mono text-xs uppercase text-obsidian-text-tertiary">Endpoint URL</p>
                      <p className="mt-2 break-all font-mono text-sm text-obsidian-text-secondary">{project.endpoint_url || '-'}</p>
                    </section>
                    <section className="border border-obsidian-border-dim bg-obsidian-base p-4">
                      <p className="font-mono text-xs uppercase text-obsidian-text-tertiary">Auth</p>
                      <p className="mt-2 font-mono text-sm text-obsidian-text-secondary">{project.auth_type}</p>
                    </section>
                    <pre className="overflow-auto border border-obsidian-border-dim bg-obsidian-base p-4 font-mono text-xs text-obsidian-text-secondary">
                      {JSON.stringify(
                        {
                          request_headers: project.request_headers || {},
                          auth_config: project.auth_config || {},
                        },
                        null,
                        2,
                      )}
                    </pre>
                  </div>
                )}
              </div>
            )
          )}
        </div>
      </aside>
    </div>
  )
}

function Stat({ label, value }: { label: string; value: string }) {
  return (
    <div className="border border-obsidian-border-dim bg-obsidian-base p-3">
      <p className="font-mono text-lg font-semibold text-obsidian-text-primary">{value}</p>
      <p className="font-mono text-[10px] uppercase tracking-wide text-obsidian-text-tertiary">{label}</p>
    </div>
  )
}

export function ProjectManagementPanel({
  title,
  subtitle,
  projects,
  basePath,
  isLoading,
  editable,
  creatable,
  emptyText,
  onChanged,
  error,
}: ProjectManagementPanelProps) {
  const [query, setQuery] = useState('')
  const [statusFilter, setStatusFilter] = useState<StatusFilter>('all')
  const [drawerMode, setDrawerMode] = useState<DrawerMode | null>(null)
  const [activeProject, setActiveProject] = useState<Project | null>(null)
  const [activeTab, setActiveTab] = useState<DetailTab>('overview')
  const [form, setForm] = useState<ProjectFormState>(blankProjectForm())
  const [formError, setFormError] = useState('')

  const visibleProjects = useMemo(() => {
    const normalized = query.trim().toLowerCase()
    return projects.filter((project) => {
      const matchesStatus = statusFilter === 'all' || project.status === statusFilter
      const matchesQuery =
        !normalized ||
        project.name.toLowerCase().includes(normalized) ||
        (project.description || '').toLowerCase().includes(normalized) ||
        (project.capability_summary || '').toLowerCase().includes(normalized)
      return matchesStatus && matchesQuery
    })
  }, [projects, query, statusFilter])

  const createMutation = useMutation({
    mutationFn: (payload: ReturnType<typeof formToPayload>) => api.post(basePath, payload),
    onSuccess: () => {
      onChanged()
      closeDrawer()
    },
  })

  const updateMutation = useMutation({
    mutationFn: (payload: { id: string; data: ReturnType<typeof formToPayload> }) => api.put(`${basePath}/${payload.id}`, payload.data),
    onSuccess: () => {
      onChanged()
      closeDrawer()
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`${basePath}/${id}`),
    onSuccess: () => {
      onChanged()
      closeDrawer()
    },
  })

  const statusMutation = useMutation({
    mutationFn: (payload: { project: Project; status: ProjectStatus }) =>
      api.put(`${basePath}/${payload.project.id}`, formToPayload({ ...projectToForm(payload.project), status: payload.status })),
    onSuccess: onChanged,
  })

  const closeDrawer = () => {
    setDrawerMode(null)
    setActiveProject(null)
    setActiveTab('overview')
    setForm(blankProjectForm())
    setFormError('')
  }

  const openCreate = () => {
    setActiveProject(null)
    setForm(blankProjectForm())
    setFormError('')
    setDrawerMode('create')
  }

  const openProject = (project: Project, tab: DetailTab = 'overview') => {
    setActiveProject(project)
    setActiveTab(tab)
    setDrawerMode('view')
  }

  const openEdit = (project: Project) => {
    setActiveProject(project)
    setForm(projectToForm(project))
    setFormError('')
    setDrawerMode('edit')
  }

  const submit = (event: React.FormEvent) => {
    event.preventDefault()
    setFormError('')
    try {
      const payload = formToPayload({ ...form, status: drawerMode === 'create' ? 'offline' : form.status, source_type: 'mixed' })
      if (!payload.name) return
      if (drawerMode === 'edit' && activeProject) {
        updateMutation.mutate({ id: activeProject.id, data: payload })
      } else {
        createMutation.mutate(payload)
      }
    } catch (err) {
      setFormError((err as Error).message)
    }
  }

  const currentProject = activeProject ? projects.find((project) => project.id === activeProject.id) || activeProject : null

  return (
    <section className="space-y-4 border border-obsidian-border-dim bg-obsidian-surface p-5">
      <div className="flex flex-col gap-4 lg:flex-row lg:items-start lg:justify-between">
        <div>
          <h2 className="font-mono text-lg font-semibold text-obsidian-text-primary">{title}</h2>
          <p className="mt-1 max-w-2xl font-mono text-sm text-obsidian-text-secondary">{subtitle}</p>
        </div>
        {creatable && (
          <button
            type="button"
            onClick={openCreate}
            className="w-fit border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white"
          >
            New Project
          </button>
        )}
      </div>

      <div className="grid gap-3 md:grid-cols-[1fr_auto]">
        <input
          value={query}
          onChange={(event) => setQuery(event.target.value)}
          placeholder="Search projects"
          className="border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
        />
        <div className="flex gap-2">
          {(['all', 'online', 'offline'] as StatusFilter[]).map((status) => (
            <button
              key={status}
              type="button"
              onClick={() => setStatusFilter(status)}
              className={`border px-3 py-2 font-mono text-xs uppercase ${
                statusFilter === status
                  ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-accent'
                  : 'border-obsidian-border-dim bg-obsidian-base text-obsidian-text-secondary hover:border-obsidian-border-med'
              }`}
            >
              {status}
            </button>
          ))}
        </div>
      </div>

      {error ? (
        <p className="border border-obsidian-negative-dim bg-obsidian-negative-dim/20 p-3 font-mono text-sm text-obsidian-negative">{error}</p>
      ) : isLoading ? (
        <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>
      ) : visibleProjects.length === 0 ? (
        <p className="border border-obsidian-border-dim bg-obsidian-base p-4 font-mono text-sm text-obsidian-text-secondary">{emptyText}</p>
      ) : (
        <div className="grid gap-3 xl:grid-cols-2">
          {visibleProjects.map((project) => {
            const counts = sourceCounts(project)
            const syncedAt = latestSync(project)
            return (
              <article key={project.id} className="border border-obsidian-border-dim bg-obsidian-base p-4 transition-colors hover:border-obsidian-border-med">
                <div className="flex items-start justify-between gap-4">
                  <button type="button" onClick={() => openProject(project)} className="min-w-0 flex-1 text-left">
                    <div className="flex flex-wrap items-center gap-2">
                      <h3 className="truncate font-mono text-sm font-semibold text-obsidian-text-primary">{project.name}</h3>
                      <span className={`border px-2 py-0.5 font-mono text-[10px] uppercase ${statusClasses[project.status] || statusClasses.offline}`}>
                        {project.status}
                      </span>
                      <span className="border border-obsidian-border-dim px-2 py-0.5 font-mono text-[10px] uppercase text-obsidian-text-tertiary">
                        {project.source_scope}
                      </span>
                    </div>
                    <p className="mt-2 line-clamp-2 font-mono text-xs leading-5 text-obsidian-text-secondary">
                      {project.capability_summary || project.description || 'No summary yet.'}
                    </p>
                    <div className="mt-3 grid gap-2 font-mono text-[10px] uppercase text-obsidian-text-tertiary sm:grid-cols-3">
                      <span>API {counts.api} · MCP {counts.mcp} · Skill {counts.skill}</span>
                      <span>{project.capabilities?.length || 0} capabilities</span>
                      <span>{syncedAt ? `Synced ${syncedAt.toLocaleDateString()}` : 'No sync yet'}</span>
                    </div>
                  </button>
                  <div className="flex shrink-0 flex-col items-end gap-2">
                    {editable && (
                      <button
                        type="button"
                        onClick={() => statusMutation.mutate({ project, status: project.status === 'online' ? 'offline' : 'online' })}
                        disabled={statusMutation.isPending}
                        className="border border-obsidian-border-dim px-3 py-1.5 font-mono text-xs text-obsidian-text-secondary hover:border-obsidian-accent hover:text-obsidian-accent disabled:opacity-50"
                      >
                        {project.status === 'online' ? 'Take Offline' : 'Bring Online'}
                      </button>
                    )}
                    <div className="flex gap-2">
                      <button
                        type="button"
                        onClick={() => openProject(project, 'sources')}
                        className="font-mono text-xs text-obsidian-accent hover:underline"
                      >
                        Sources
                      </button>
                      {editable && (
                        <button type="button" onClick={() => openEdit(project)} className="font-mono text-xs text-obsidian-accent hover:underline">
                          Edit
                        </button>
                      )}
                      {editable && (
                        <button
                          type="button"
                          onClick={() => {
                            if (window.confirm(`Delete project "${project.name}"?`)) {
                              deleteMutation.mutate(project.id)
                            }
                          }}
                          className="font-mono text-xs text-obsidian-negative hover:underline"
                        >
                          Delete
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              </article>
            )
          })}
        </div>
      )}

      {drawerMode && (
        <ProjectDrawer
          project={currentProject}
          mode={drawerMode}
          editable={editable}
          basePath={basePath}
          form={form}
          setForm={setForm}
          formError={formError}
          activeTab={activeTab}
          setActiveTab={setActiveTab}
          isPending={createMutation.isPending || updateMutation.isPending}
          onClose={closeDrawer}
          onEdit={() => currentProject && openEdit(currentProject)}
          onSubmit={submit}
          onChanged={onChanged}
        />
      )}
    </section>
  )
}
