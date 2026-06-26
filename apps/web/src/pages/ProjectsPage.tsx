import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ProjectManagementPanel } from '../components/ProjectManagementPanel'
import { api } from '../lib/api'
import { type Project } from '../lib/projectTypes'
import { useTenantStore } from '../store/tenantStore'

export function ProjectsPage() {
  const { currentTenant } = useTenantStore()
  const queryClient = useQueryClient()

  const { data: projects = [], isLoading, error } = useQuery({
    queryKey: ['projects', currentTenant?.id],
    queryFn: async () => {
      const res = await api.get<{ projects: Project[] }>('/projects')
      return res.data.projects
    },
    enabled: !!currentTenant,
  })

  if (!currentTenant) {
    return (
      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <p className="font-mono text-sm text-obsidian-text-secondary">Please select or create a plan first.</p>
      </section>
    )
  }

  const refreshProjects = () => {
    queryClient.invalidateQueries({ queryKey: ['projects', currentTenant.id] })
  }

  const systemProjects = projects.filter((project) => project.source_scope === 'system')
  const tenantProjects = projects.filter((project) => project.source_scope === 'tenant')
  const errorMessage =
    (error as { response?: { data?: { error?: string } }; message?: string } | null)?.response?.data?.error ||
    (error as { message?: string } | null)?.message

  return (
    <div className="space-y-6">
      <header>
        <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
          <span className="text-obsidian-accent">&gt;</span> Projects
        </h1>
        <p className="mt-1 font-mono text-sm text-obsidian-text-secondary">
          Manage project containers, then configure API, HTTP MCP, and Skill sources inside each project.
        </p>
      </header>

      <ProjectManagementPanel
        title="System Projects"
        subtitle="Shared projects published by admins. You can inspect sources and loaded capabilities here."
        projects={systemProjects}
        basePath="/projects"
        isLoading={isLoading}
        editable={false}
        creatable={false}
        emptyText="No system projects available."
        onChanged={refreshProjects}
        error={errorMessage}
      />

      <ProjectManagementPanel
        title="My Projects"
        subtitle="Your own projects. New projects start offline, then can be brought online after sources are configured."
        projects={tenantProjects}
        basePath="/projects"
        isLoading={isLoading}
        editable
        creatable
        emptyText="No personal projects yet."
        onChanged={refreshProjects}
        error={errorMessage}
      />
    </div>
  )
}
