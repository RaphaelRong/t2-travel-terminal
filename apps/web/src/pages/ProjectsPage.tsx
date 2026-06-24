import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { api } from '../lib/api'
import { useTenantStore } from '../store/tenantStore'

interface Project {
  id: string
  name: string
  description?: string
  created_at: string
}

export function ProjectsPage() {
  const { currentTenant } = useTenantStore()
  const queryClient = useQueryClient()
  const [newProject, setNewProject] = useState({ name: '', description: '' })

  const { data: projects, isLoading } = useQuery({
    queryKey: ['projects', currentTenant?.id],
    queryFn: async () => {
      const res = await api.get<{ projects: Project[] }>('/projects')
      return res.data.projects
    },
    enabled: !!currentTenant,
  })

  const createMutation = useMutation({
    mutationFn: (payload: { name: string; description: string }) =>
      api.post('/projects', payload),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', currentTenant?.id] })
      setNewProject({ name: '', description: '' })
    },
  })

  const deleteMutation = useMutation({
    mutationFn: (id: string) => api.delete(`/projects/${id}`),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['projects', currentTenant?.id] })
    },
  })

  const handleCreate = (e: React.FormEvent) => {
    e.preventDefault()
    if (!newProject.name.trim()) return
    createMutation.mutate({
      name: newProject.name.trim(),
      description: newProject.description.trim(),
    })
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
        <span className="text-obsidian-accent">&gt;</span> Projects
      </h1>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Create Project
        </h2>
        <form onSubmit={handleCreate} className="space-y-3">
          <input
            type="text"
            value={newProject.name}
            onChange={(e) => setNewProject({ ...newProject, name: e.target.value })}
            placeholder="Project name"
            className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
          />
          <textarea
            value={newProject.description}
            onChange={(e) => setNewProject({ ...newProject, description: e.target.value })}
            placeholder="Description"
            className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none transition-colors focus:border-obsidian-accent"
          />
          <button
            type="submit"
            disabled={createMutation.isPending}
            className="border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:opacity-50"
          >
            Create
          </button>
        </form>
      </section>

      <section className="border border-obsidian-border-dim bg-obsidian-surface p-6">
        <h2 className="mb-4 border-b border-obsidian-border-dim pb-2 font-mono text-xs font-semibold uppercase tracking-wider text-obsidian-accent">
          Project List
        </h2>
        {isLoading ? (
          <p className="font-mono text-sm text-obsidian-text-secondary">Loading...</p>
        ) : projects?.length === 0 ? (
          <p className="font-mono text-sm text-obsidian-text-secondary">No projects yet.</p>
        ) : (
          <div className="space-y-3">
            {projects?.map((p) => (
              <div
                key={p.id}
                className="flex items-start justify-between border border-obsidian-border-dim bg-obsidian-base p-4"
              >
                <div>
                  <h3 className="font-mono text-sm font-medium text-obsidian-text-primary">{p.name}</h3>
                  {p.description && <p className="font-mono text-xs text-obsidian-text-secondary">{p.description}</p>}
                  <p className="mt-1 font-mono text-[10px] text-obsidian-text-tertiary">
                    Created {new Date(p.created_at).toLocaleString()}
                  </p>
                </div>
                <button
                  onClick={() => deleteMutation.mutate(p.id)}
                  className="font-mono text-xs text-obsidian-negative hover:underline"
                >
                  Delete
                </button>
              </div>
            ))}
          </div>
        )}
      </section>
    </div>
  )
}
