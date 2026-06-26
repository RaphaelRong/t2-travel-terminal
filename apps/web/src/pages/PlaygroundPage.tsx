import { useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import { type LLMProfile } from '../lib/llmProfiles'
import { type Project } from '../lib/projectTypes'
import { useTenantStore } from '../store/tenantStore'

type ChatRole = 'user' | 'assistant'

interface ChatMessage {
  id: string
  role: ChatRole
  content: string
  createdAt: string
}

function makeMessage(role: ChatRole, content: string): ChatMessage {
  return {
    id: `${role}-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    role,
    content,
    createdAt: new Date().toLocaleTimeString(),
  }
}

export function PlaygroundPage() {
  const { currentTenant } = useTenantStore()
  const [selectedProfileId, setSelectedProfileId] = useState('')
  const [selectedModel, setSelectedModel] = useState('')
  const [selectedProjectId, setSelectedProjectId] = useState('')
  const [draft, setDraft] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([
    makeMessage(
      'assistant',
      'Playground is ready. Choose an LLM profile, model, and project, then send a message to prepare an Agent request.',
    ),
  ])

  const { data: profiles = [], isLoading: profilesLoading } = useQuery({
    queryKey: ['llm-profiles'],
    queryFn: async () => {
      const res = await api.get<{ profiles: LLMProfile[] }>('/llm-profiles')
      return res.data.profiles
    },
  })

  const { data: projects = [], isLoading: projectsLoading } = useQuery({
    queryKey: ['projects', currentTenant?.id],
    queryFn: async () => {
      const res = await api.get<{ projects: Project[] }>('/projects')
      return res.data.projects
    },
    enabled: !!currentTenant,
  })

  const activeProfiles = profiles.filter((profile) => profile.status === 'active')
  const selectedProfile = activeProfiles.find((profile) => profile.id === selectedProfileId) || activeProfiles[0]
  const selectedProject = projects.find((project) => project.id === selectedProjectId)
  const modelOptions = useMemo(() => {
    const models = selectedProfile?.models || []
    if (selectedProfile?.default_model && !models.includes(selectedProfile.default_model)) {
      return [selectedProfile.default_model, ...models]
    }
    return models
  }, [selectedProfile])

  const currentModel = selectedModel || selectedProfile?.default_model || modelOptions[0] || ''
  const canSend = Boolean(draft.trim() && selectedProfile && currentModel && selectedProject)

  const handleSend = (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedProfile || !currentModel || !selectedProject || !draft.trim()) return

    const content = draft.trim()
    setDraft('')
    setMessages((items) => [
      ...items,
      makeMessage('user', content),
      makeMessage(
        'assistant',
        [
          `Prepared request for ${selectedProfile.display_name} / ${currentModel}.`,
          `Project: ${selectedProject.name}.`,
          'Agent orchestration is not connected yet; this Playground now captures the selected LLM, model, project, and prompt shape.',
        ].join('\n'),
      ),
    ])
  }

  return (
    <div className="space-y-6">
      <header>
        <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
          <span className="text-obsidian-accent">&gt;</span> Playground
        </h1>
        <p className="mt-1 font-mono text-sm text-obsidian-text-secondary">
          Test chat requests against a selected LLM profile, model, and project context.
        </p>
      </header>

      <section className="grid min-h-[620px] gap-4 lg:grid-cols-[320px_1fr]">
        <aside className="space-y-4 border border-obsidian-border-dim bg-obsidian-surface p-4">
          <div>
            <label className="mb-2 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              LLM Profile
            </label>
            <select
              value={selectedProfile?.id || ''}
              onChange={(e) => {
                setSelectedProfileId(e.target.value)
                setSelectedModel('')
              }}
              className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent"
            >
              {activeProfiles.map((profile) => (
                <option key={profile.id} value={profile.id}>
                  {profile.display_name}
                </option>
              ))}
            </select>
            {!profilesLoading && activeProfiles.length === 0 && (
              <p className="mt-2 font-mono text-xs leading-5 text-obsidian-text-secondary">
                No active LLM profile yet. Configure one in{' '}
                <Link to="/profile" className="text-obsidian-accent hover:underline">
                  Profile
                </Link>
                .
              </p>
            )}
          </div>

          <div>
            <label className="mb-2 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              Model
            </label>
            <select
              value={currentModel}
              onChange={(e) => setSelectedModel(e.target.value)}
              disabled={!selectedProfile || modelOptions.length === 0}
              className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent disabled:opacity-50"
            >
              {modelOptions.length === 0 ? (
                <option value="">No models configured</option>
              ) : (
                modelOptions.map((model) => (
                  <option key={model} value={model}>
                    {model}
                  </option>
                ))
              )}
            </select>
          </div>

          <div>
            <label className="mb-2 block font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
              Project
            </label>
            <select
              value={selectedProject?.id || ''}
              onChange={(e) => setSelectedProjectId(e.target.value)}
              disabled={!currentTenant || projectsLoading || projects.length === 0}
              className="w-full border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm text-obsidian-text-primary outline-none focus:border-obsidian-accent disabled:opacity-50"
            >
              <option value="">Select a project</option>
              {projects.map((project) => (
                <option key={project.id} value={project.id}>
                  {project.source_scope === 'system' ? '[System] ' : ''}{project.name}
                </option>
              ))}
            </select>
            {!currentTenant && (
              <p className="mt-2 font-mono text-xs leading-5 text-obsidian-text-secondary">
                Select or create a plan before using project context.
              </p>
            )}
          </div>

          <div className="border-t border-obsidian-border-dim pt-4 font-mono text-xs leading-5 text-obsidian-text-secondary">
            <p className="text-obsidian-text-primary">Request context</p>
            <p>Provider: {selectedProfile?.display_name || '-'}</p>
            <p>Model: {currentModel || '-'}</p>
            <p>Project: {selectedProject?.name || '-'}</p>
          </div>
        </aside>

        <section className="flex min-h-[620px] flex-col border border-obsidian-border-dim bg-obsidian-surface">
          <div className="border-b border-obsidian-border-dim px-4 py-3">
            <p className="font-mono text-xs uppercase tracking-wider text-obsidian-accent">
              ChatBox
            </p>
          </div>

          <div className="flex-1 space-y-3 overflow-y-auto p-4">
            {messages.map((message) => (
              <div
                key={message.id}
                className={`max-w-[86%] border px-4 py-3 ${
                  message.role === 'user'
                    ? 'ml-auto border-obsidian-accent bg-obsidian-accent/10'
                    : 'border-obsidian-border-dim bg-obsidian-base'
                }`}
              >
                <div className="mb-2 flex items-center justify-between gap-3 font-mono text-[10px] uppercase tracking-wider text-obsidian-text-tertiary">
                  <span>{message.role}</span>
                  <span>{message.createdAt}</span>
                </div>
                <p className="whitespace-pre-wrap font-mono text-sm leading-6 text-obsidian-text-primary">
                  {message.content}
                </p>
              </div>
            ))}
          </div>

          <form onSubmit={handleSend} className="border-t border-obsidian-border-dim p-4">
            <textarea
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder="Ask about selected project data..."
              rows={4}
              className="w-full resize-none border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm leading-6 text-obsidian-text-primary outline-none focus:border-obsidian-accent"
            />
            <div className="mt-3 flex items-center justify-between gap-3">
              <p className="font-mono text-xs text-obsidian-text-secondary">
                {canSend ? 'Ready to prepare request.' : 'Choose profile, model, and project before sending.'}
              </p>
              <button
                type="submit"
                disabled={!canSend}
                className="border border-obsidian-accent bg-obsidian-accent/10 px-4 py-2 font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
              >
                Send
              </button>
            </div>
          </form>
        </section>
      </section>
    </div>
  )
}
