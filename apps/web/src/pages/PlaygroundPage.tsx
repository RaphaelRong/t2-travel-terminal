import { useEffect, useMemo, useState } from 'react'
import { Link } from 'react-router-dom'
import { useMutation, useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import { getToken, type UserRole } from '../lib/auth'
import { type LLMProfile } from '../lib/llmProfiles'
import { type Project } from '../lib/projectTypes'
import { useAuthStore } from '../store/authStore'
import { useTenantStore } from '../store/tenantStore'

type ChatRole = 'user' | 'assistant' | 'tool'
type JsonRecord = Record<string, unknown>

interface AgentSession {
  id: string
  title: string
  status: string
  created_at: string
  updated_at: string
}

interface SessionProject {
  project_id: string
  name: string
}

interface AgentMessage {
  id: string
  role: 'system' | ChatRole
  content: string
  tool_name?: string
  tool_result?: JsonRecord
  created_at: string
}

interface Subscription {
  id: string
  name: string
  slug?: string
  plan_id?: string
  plan_name: string
  effective_role: UserRole
  role: string
}

interface ChatMessage {
  id: string
  role: ChatRole
  content: string
  createdAt: string
  toolName?: string
}

interface SSEEvent {
  type: string
  data: JsonRecord
}

function makeMessage(role: ChatRole, content: string, extra?: Partial<ChatMessage>): ChatMessage {
  return {
    id: `${role}-${Date.now()}-${Math.random().toString(16).slice(2)}`,
    role,
    content,
    createdAt: new Date().toLocaleTimeString(),
    ...extra,
  }
}

function parseSSEEvents(chunk: string): { events: SSEEvent[]; remainder: string } {
  const events: SSEEvent[] = []
  const parts = chunk.split('\n\n')
  const remainder = parts.pop() || ''

  for (const part of parts) {
    const lines = part.split('\n')
    let event = 'message'
    const dataLines: string[] = []
    for (const line of lines) {
      if (line.startsWith('event:')) {
        event = line.slice(6).trim()
      } else if (line.startsWith('data:')) {
        dataLines.push(line.slice(5).trim())
      }
    }
    if (dataLines.length === 0) continue
    try {
      const data = JSON.parse(dataLines.join('\n')) as Record<string, unknown>
      events.push({ type: event, data })
    } catch {
      events.push({ type: event, data: { raw: dataLines.join('\n') } })
    }
  }

  return { events, remainder }
}

function readString(record: JsonRecord | undefined, ...keys: string[]) {
  for (const key of keys) {
    const value = record?.[key]
    if (typeof value === 'string') return value
  }
  return ''
}

function readRecord(record: JsonRecord | undefined, ...keys: string[]) {
  for (const key of keys) {
    const value = record?.[key]
    if (value && typeof value === 'object' && !Array.isArray(value)) {
      return value as JsonRecord
    }
  }
  return undefined
}

function sessionTitleFromPrompt(content: string, projectName?: string) {
  const compact = content.replace(/\s+/g, ' ').trim()
  const title = compact.length > 28 ? `${compact.slice(0, 28)}...` : compact
  return projectName ? `${projectName}: ${title}` : title || 'New Agent Session'
}

function formatToolMessage(name: string, result: JsonRecord) {
  if (name === 'workflow_tool') {
    const goal = readString(result, 'goal')
    const strategy = readString(result, 'strategy')
    const comparison = readString(result, 'comparison')
    const nextAction = readString(result, 'next_action')
    const steps = Array.isArray(result.steps) ? result.steps : []
    const lines = ['Workflow plan']
    if (goal) lines.push(`Goal: ${goal}`)
    if (strategy) lines.push(`Strategy: ${strategy}`)
    if (steps.length > 0) {
      lines.push('', 'Steps:')
      steps.forEach((raw, index) => {
        const step = raw && typeof raw === 'object' ? (raw as JsonRecord) : undefined
        const stepName = readString(step, 'name') || `Step ${index + 1}`
        const instruction = readString(step, 'instruction')
        const isolation = readString(step, 'isolation')
        const expected = readString(step, 'expected_output')
        lines.push(`${index + 1}. ${stepName}${isolation ? ` [${isolation}]` : ''}`)
        if (instruction) lines.push(`   - ${instruction}`)
        if (expected) lines.push(`   - Output: ${expected}`)
      })
    }
    if (comparison) lines.push('', `Comparison: ${comparison}`)
    if (nextAction) lines.push('', `Next: ${nextAction}`)
    return lines.join('\n')
  }

  return `Tool: ${name}\n\`\`\`json\n${JSON.stringify(result, null, 2)}\n\`\`\``
}

export function PlaygroundPage() {
  const { currentTenant, setCurrentTenant } = useTenantStore()
  const setRole = useAuthStore((s) => s.setRole)
  const [selectedProfileId, setSelectedProfileId] = useState('')
  const [selectedModel, setSelectedModel] = useState('')
  const [selectedProjectId, setSelectedProjectId] = useState('')
  const [draft, setDraft] = useState('')
  const [messages, setMessages] = useState<ChatMessage[]>([
    makeMessage(
      'assistant',
      'Playground is ready. Choose an LLM profile, model, and project, then send a message to start an Agent session.',
    ),
  ])
  const [sessionId, setSessionId] = useState<string | null>(null)
  const [isStreaming, setIsStreaming] = useState(false)

  const { data: profiles = [], isLoading: profilesLoading } = useQuery({
    queryKey: ['llm-profiles'],
    queryFn: async () => {
      const res = await api.get<{ profiles: LLMProfile[] }>('/llm-profiles')
      return res.data.profiles
    },
  })

  const { data: subscriptions = [] } = useQuery({
    queryKey: ['subscriptions'],
    queryFn: async () => {
      const res = await api.get<{ subscriptions: Subscription[] }>('/tenants')
      return res.data.subscriptions
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

  const {
    data: sessions = [],
    isLoading: sessionsLoading,
    refetch: refetchSessions,
  } = useQuery({
    queryKey: ['agent-sessions', currentTenant?.id],
    queryFn: async () => {
      const res = await api.get<{ sessions: AgentSession[] }>('/agent/sessions', {
        params: { status: 'active', limit: 20 },
      })
      return res.data.sessions
    },
  })

  useEffect(() => {
    if (!currentTenant && subscriptions.length > 0) {
      const selected = subscriptions[0]
      setCurrentTenant(selected)
      setRole(selected.effective_role)
    }
  }, [currentTenant, setCurrentTenant, setRole, subscriptions])

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
  const canSend = Boolean(draft.trim() && selectedProfile && currentModel && !isStreaming)

  const resetConversation = (projectId = selectedProjectId) => {
    setSessionId(null)
    setSelectedProjectId(projectId)
    setMessages([
      makeMessage(
        'assistant',
        projectId
          ? 'New Agent session. Send a message when you are ready.'
          : 'New Agent session. Send a message when you are ready.',
      ),
    ])
  }

  const [deletingSessionId, setDeletingSessionId] = useState<string | null>(null)

  const deleteSessionMutation = useMutation({
    mutationFn: async (id: string) => {
      await api.delete(`/agent/sessions/${id}`)
    },
    onSuccess: () => {
      void refetchSessions()
      if (sessionId === deletingSessionId) {
        setSessionId(null)
        resetConversation()
      }
    },
  })

  const handleDeleteSession = (session: AgentSession) => {
    if (window.confirm(`是否删除 ${session.title || 'Untitled session'}？此操作不可恢复。`)) {
      setDeletingSessionId(session.id)
      deleteSessionMutation.mutate(session.id, {
        onSettled: () => setDeletingSessionId(null),
      })
    }
  }

  const selectSession = async (id: string) => {
    if (isStreaming) return
    setSessionId(id)
    const [messageRes, projectRes] = await Promise.all([
      api.get<{ messages: AgentMessage[] }>(`/agent/sessions/${id}/messages`),
      api.get<{ projects: SessionProject[] }>(`/agent/sessions/${id}/projects`),
    ])
    const attachedProject = projectRes.data.projects[0]
    setSelectedProjectId(attachedProject?.project_id || '')
    const loaded = messageRes.data.messages
      .filter((message) => message.role !== 'system')
      .reduce<ChatMessage[]>((items, message) => {
        const role = message.role as ChatRole
        if (role === 'assistant' && !message.content?.trim()) return items
        if (role === 'tool') {
          const name = message.tool_name || 'tool'
          const result = message.tool_result || {}
          items.push({
              id: message.id,
              role,
              content: formatToolMessage(name, result),
              createdAt: new Date(message.created_at).toLocaleTimeString(),
              toolName: name,
          })
          return items
        }
        items.push({
          id: message.id,
          role,
          content: message.content,
          createdAt: new Date(message.created_at).toLocaleTimeString(),
          toolName: message.tool_name,
        })
        return items
      }, [])
    setMessages(
      loaded.length > 0
        ? loaded
        : [makeMessage('assistant', 'Session opened. Continue the conversation when you are ready.')],
    )
  }

  const ensureSessionProject = async (id: string) => {
    if (!selectedProject) return
    await api.post(`/agent/sessions/${id}/projects`, {
      project_id: selectedProject.id,
    })
  }

  const ensureSession = async (firstMessage: string): Promise<string> => {
    if (sessionId) {
      await ensureSessionProject(sessionId)
      return sessionId
    }
    const res = await api.post<{ session: { id: string } }>('/agent/sessions', {
      title: sessionTitleFromPrompt(firstMessage, selectedProject?.name),
      project_ids: selectedProject ? [selectedProject.id] : [],
    })
    const id = res.data.session.id
    setSessionId(id)
    void refetchSessions()
    return id
  }

  const handleSend = async (e: React.FormEvent) => {
    e.preventDefault()
    if (!selectedProfile || !currentModel || !draft.trim()) return

    const content = draft.trim()
    setDraft('')
    setMessages((items) => {
      // Replace placeholder with first real user message if it's still the initial message.
      const hasRealMessages = items.some((m) => m.role === 'user')
      const userMsg = makeMessage('user', content)
      if (!hasRealMessages) {
        return [userMsg]
      }
      return [...items, userMsg]
    })
    setIsStreaming(true)

    try {
      const sid = await ensureSession(content)
      await streamAgentResponse(sid, content)
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Failed to reach Agent orchestration'
      setMessages((items) => [
        ...items,
        makeMessage('assistant', `Error: ${message}`),
      ])
    } finally {
      setIsStreaming(false)
      void refetchSessions()
    }
  }

  const streamAgentResponse = async (sid: string, content: string) => {
    const token = getToken()
    const res = await fetch(`/api/v1/agent/sessions/${sid}/messages`, {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        ...(token ? { Authorization: `Bearer ${token}` } : {}),
        ...(currentTenant ? { 'X-Tenant-ID': currentTenant.id } : {}),
      },
      body: JSON.stringify({
        content,
        llm_profile_id: selectedProfile?.id,
        model: currentModel,
      }),
    })

    if (!res.ok) {
      const body = await res.text()
      throw new Error(`HTTP ${res.status}: ${body}`)
    }

    const reader = res.body?.getReader()
    if (!reader) throw new Error('Response body is not readable')

    const decoder = new TextDecoder()
    let buffer = ''
    let assistantStarted = false

    let streamDone = false
    while (!streamDone) {
      const { done, value } = await reader.read()
      if (done) {
        streamDone = true
        break
      }
      buffer += decoder.decode(value, { stream: true })
      const { events, remainder } = parseSSEEvents(buffer)
      buffer = remainder

      for (const event of events) {
        switch (event.type) {
          case 'assistant_message': {
            const contentText = readString(event.data, 'content', 'Content')
            const reasoningText = readString(event.data, 'reasoning_content', 'ReasoningContent')
            const text = [contentText, reasoningText ? `\n[reasoning]\n${reasoningText}` : '']
              .filter(Boolean)
              .join('')
            if (!text.trim()) break
            if (!assistantStarted) {
              assistantStarted = true
              setMessages((items) => [...items, makeMessage('assistant', text)])
            } else {
              setMessages((items) => {
                const last = items[items.length - 1]
                if (last && last.role === 'assistant') {
                  const updated = [...items]
                  updated[updated.length - 1] = { ...last, content: last.content + '\n' + text }
                  return updated
                }
                return [...items, makeMessage('assistant', text)]
              })
            }
            break
          }
          case 'tool_message': {
            const tool = readRecord(event.data, 'tool_message', 'toolMessage') || event.data
            const name = readString(tool, 'tool_name', 'ToolName') || 'tool'
            const result = readRecord(tool, 'tool_result', 'ToolResult') || {}
            setMessages((items) => [
              ...items,
              makeMessage('tool', formatToolMessage(name, result), { toolName: name }),
            ])
            break
          }
          case 'error': {
            const text = (event.data?.error as string) || 'Unknown agent error'
            setMessages((items) => [...items, makeMessage('assistant', `Error: ${text}`)])
            break
          }
          case 'done':
          case 'user_message':
          default:
            // no-op
            break
        }
      }
    }
  }

  return (
    <div className="flex h-[calc(100vh-120px)] min-h-[620px] flex-col gap-4">
      <header className="flex flex-col gap-3 border border-obsidian-border-dim bg-obsidian-surface p-4 lg:flex-row lg:items-center lg:justify-between">
        <div>
          <h1 className="font-mono text-2xl font-bold tracking-tight text-obsidian-text-primary">
            <span className="text-obsidian-accent">&gt;</span> Playground
          </h1>
          <p className="mt-1 font-mono text-sm text-obsidian-text-secondary">
            Test chat requests against a selected LLM profile, model, and project context.
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-2">
          <select
            value={selectedProject?.id || ''}
            onChange={(e) => resetConversation(e.target.value)}
            disabled={!currentTenant || projectsLoading || isStreaming}
            className="max-w-[180px] border border-obsidian-border-dim bg-obsidian-base px-2 py-1.5 font-mono text-xs text-obsidian-text-primary outline-none focus:border-obsidian-accent disabled:opacity-50"
            title="Project"
          >
            <option value="">No project</option>
            {projects.map((project) => (
              <option key={project.id} value={project.id}>
                {project.source_scope === 'system' ? '[System] ' : ''}
                {project.name}
              </option>
            ))}
          </select>

          <select
            value=""
            onChange={() => {}}
            disabled={!selectedProject || isStreaming}
            className="max-w-[160px] border border-obsidian-border-dim bg-obsidian-base px-2 py-1.5 font-mono text-xs text-obsidian-text-primary outline-none focus:border-obsidian-accent disabled:opacity-50"
            title="Capabilities"
          >
            <option value="">
              {selectedProject
                ? `${(selectedProject.capabilities || []).filter((c) => (c.status || 'active') === 'active').length} capabilities`
                : 'Capabilities'}
            </option>
            {selectedProject?.capabilities?.map((capability) => (
              <option key={capability.id || capability.name} value={capability.id || capability.name} disabled>
                {capability.name} ({capability.status || 'active'})
              </option>
            ))}
          </select>

          <select
            value={selectedProfile?.id || ''}
            onChange={(e) => {
              setSelectedProfileId(e.target.value)
              setSelectedModel('')
            }}
            disabled={activeProfiles.length === 0 || isStreaming}
            className="max-w-[160px] border border-obsidian-border-dim bg-obsidian-base px-2 py-1.5 font-mono text-xs text-obsidian-text-primary outline-none focus:border-obsidian-accent disabled:opacity-50"
            title="LLM profile"
          >
            {activeProfiles.map((profile) => (
              <option key={profile.id} value={profile.id}>
                {profile.display_name}
              </option>
            ))}
          </select>

          <select
            value={currentModel}
            onChange={(e) => setSelectedModel(e.target.value)}
            disabled={!selectedProfile || modelOptions.length === 0 || isStreaming}
            className="max-w-[180px] border border-obsidian-border-dim bg-obsidian-base px-2 py-1.5 font-mono text-xs text-obsidian-text-primary outline-none focus:border-obsidian-accent disabled:opacity-50"
            title="Model"
          >
            {modelOptions.length === 0 ? (
              <option value="">No models</option>
            ) : (
              modelOptions.map((model) => (
                <option key={model} value={model}>
                  {model}
                </option>
              ))
            )}
          </select>
        </div>
      </header>

      <section className="grid min-h-0 flex-1 gap-4 overflow-hidden lg:grid-cols-[360px_1fr]">
        <aside className="min-h-0 space-y-4 overflow-y-auto border border-obsidian-border-dim bg-obsidian-surface p-4">
          <button
            type="button"
            onClick={() => resetConversation()}
            disabled={isStreaming}
            className="w-full border border-obsidian-accent bg-obsidian-accent/10 px-3 py-2 text-left font-mono text-sm text-obsidian-accent transition-colors hover:bg-obsidian-accent hover:text-white disabled:cursor-not-allowed disabled:opacity-50"
          >
            + New session
          </button>

          <div>
            <div className="mb-2 flex items-center justify-between gap-3">
              <label className="font-mono text-xs uppercase tracking-wide text-obsidian-text-secondary">
                Sessions
              </label>
              <span className="font-mono text-[10px] uppercase tracking-wide text-obsidian-text-tertiary">
                {sessionsLoading ? 'Loading' : `${sessions.length}`}
              </span>
            </div>
            <div className="space-y-2">
              {sessions.length === 0 ? (
                <p className="border border-obsidian-border-dim bg-obsidian-base px-3 py-3 font-mono text-xs leading-5 text-obsidian-text-secondary">
                  No active sessions yet.
                </p>
              ) : (
                sessions.map((session) => (
                  <div
                    key={session.id}
                    className={`group flex items-start gap-2 border px-3 py-2 font-mono transition-colors ${
                      sessionId === session.id
                        ? 'border-obsidian-accent bg-obsidian-accent/10 text-obsidian-accent'
                        : 'border-obsidian-border-dim bg-obsidian-base text-obsidian-text-primary hover:border-obsidian-accent'
                    }`}
                  >
                    <button
                      type="button"
                      onClick={() => void selectSession(session.id)}
                      disabled={isStreaming}
                      className="min-w-0 flex-1 text-left transition-colors disabled:cursor-not-allowed disabled:opacity-50"
                    >
                      <span className="block truncate text-sm">{session.title || 'Untitled session'}</span>
                      <span className="mt-1 block text-[10px] uppercase tracking-wide text-obsidian-text-tertiary">
                        {session.id.slice(0, 8)} · {new Date(session.updated_at).toLocaleDateString()}
                      </span>
                    </button>
                    <button
                      type="button"
                      onClick={() => handleDeleteSession(session)}
                      disabled={isStreaming || deleteSessionMutation.isPending}
                      className="shrink-0 border border-obsidian-negative-dim bg-obsidian-negative-dim/10 px-2 py-1 font-mono text-[10px] uppercase text-obsidian-negative opacity-0 transition-opacity hover:bg-obsidian-negative hover:text-white group-hover:opacity-100 disabled:opacity-50"
                      title="Delete session"
                    >
                      Del
                    </button>
                  </div>
                ))
              )}
            </div>
          </div>

          {!profilesLoading && activeProfiles.length === 0 && (
            <p className="border-t border-obsidian-border-dim pt-4 font-mono text-xs leading-5 text-obsidian-text-secondary">
              No active LLM profile yet. Configure one in{' '}
              <Link to="/profile" className="text-obsidian-accent hover:underline">
                Profile
              </Link>
              .
            </p>
          )}
        </aside>

        <section className="flex min-h-0 flex-col overflow-hidden border border-obsidian-border-dim bg-obsidian-surface">
          <div className="border-b border-obsidian-border-dim px-4 py-3">
            <p className="font-mono text-xs uppercase tracking-wider text-obsidian-accent">
              ChatBox
            </p>
          </div>

          <div className="min-h-0 flex-1 space-y-3 overflow-y-auto p-4">
            {messages.map((message) => (
              <div
                key={message.id}
                className={`max-w-[86%] border px-4 py-3 ${
                  message.role === 'user'
                    ? 'ml-auto border-obsidian-accent bg-obsidian-accent/10'
                    : message.role === 'tool'
                    ? 'border-obsidian-border-dim bg-obsidian-surface'
                    : 'border-obsidian-border-dim bg-obsidian-base'
                }`}
              >
                <div className="mb-2 flex items-center justify-between gap-3 font-mono text-[10px] uppercase tracking-wider text-obsidian-text-tertiary">
                  <span>{message.role}{message.toolName ? ` / ${message.toolName}` : ''}</span>
                  <span>{message.createdAt}</span>
                </div>
                <p className="whitespace-pre-wrap font-mono text-sm leading-6 text-obsidian-text-primary">
                  {message.content}
                </p>
              </div>
            ))}
          </div>

          <form onSubmit={handleSend} className="shrink-0 border-t border-obsidian-border-dim p-4">
            <textarea
              value={draft}
              onChange={(e) => setDraft(e.target.value)}
              placeholder="Ask about selected project data..."
              rows={4}
              disabled={isStreaming}
              className="w-full resize-none border border-obsidian-border-dim bg-obsidian-base px-3 py-2 font-mono text-sm leading-6 text-obsidian-text-primary outline-none focus:border-obsidian-accent disabled:opacity-50"
            />
            <div className="mt-3 flex items-center justify-between gap-3">
              <span className="min-w-[160px] truncate font-mono text-xs text-obsidian-text-secondary">
                {isStreaming
                  ? 'Agent is thinking...'
                  : canSend
                  ? sessionId
                    ? `Session ${sessionId.slice(0, 8)}`
                    : selectedProject?.name || 'Ready'
                  : 'Choose a profile and model.'}
              </span>
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
