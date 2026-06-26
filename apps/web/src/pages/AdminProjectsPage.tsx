import { useQuery, useQueryClient } from '@tanstack/react-query'
import { ProjectManagementPanel } from '../components/ProjectManagementPanel'
import { api } from '../lib/api'
import { type Project } from '../lib/projectTypes'

export function AdminProjectsPage() {
  const queryClient = useQueryClient()

  const { data: projects = [], isLoading, error } = useQuery({
    queryKey: ['admin-projects'],
    queryFn: async () => {
      const res = await api.get<{ projects: Project[] }>('/admin/projects')
      return res.data.projects
    },
  })

  const refreshProjects = () => {
    queryClient.invalidateQueries({ queryKey: ['admin-projects'] })
    queryClient.invalidateQueries({ queryKey: ['projects'] })
  }

  const errorMessage =
    (error as { response?: { data?: { error?: string } }; message?: string } | null)?.response?.data?.error ||
    (error as { message?: string } | null)?.message

  return (
    <ProjectManagementPanel
      title="System Projects"
      subtitle="Admin-managed projects. Create the container first, configure sources in the detail panel, then bring it online."
      projects={projects}
      basePath="/admin/projects"
      isLoading={isLoading}
      editable
      creatable
      emptyText="No system projects yet."
      onChanged={refreshProjects}
      error={errorMessage}
    />
  )
}
