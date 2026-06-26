export type ProjectScope = 'system' | 'tenant'
export type ProjectStatus = 'draft' | 'online' | 'offline' | 'archived'
export type ProjectSourceType = 'api' | 'mcp' | 'skill' | 'mixed'
export type ProjectAuthType = 'none' | 'api_key' | 'bearer' | 'oauth2' | 'basic' | 'custom'
export type CapabilityKind = 'api' | 'tool' | 'skill'
export type ProjectIntegrationKind = 'api' | 'mcp' | 'skill'
export type ProjectIntegrationAuthType = 'inherit' | ProjectAuthType

export interface ProjectCapability {
  id?: string
  integration_id?: string
  kind: CapabilityKind
  name: string
  description?: string
  status?: 'active' | 'inactive'
  request_method?: string
  request_path?: string
  input_schema?: Record<string, unknown>
  output_schema?: Record<string, unknown>
  metadata?: Record<string, unknown>
}

export interface ProjectIntegration {
  id?: string
  project_id?: string
  kind: ProjectIntegrationKind
  name: string
  description?: string
  status?: 'active' | 'inactive'
  endpoint_url?: string
  documentation_url?: string
  transport?: 'http'
  auth_type?: ProjectIntegrationAuthType
  request_headers?: Record<string, unknown>
  auth_config?: Record<string, unknown>
  metadata?: Record<string, unknown>
  last_synced_at?: string
  sync_status?: 'idle' | 'success' | 'failed'
  sync_error?: string
}

export interface Project {
  id: string
  tenant_id?: string
  source_scope: ProjectScope
  kind: string
  status: ProjectStatus
  source_type: ProjectSourceType
  name: string
  description?: string
  endpoint_url?: string
  request_method: string
  request_path?: string
  request_headers: Record<string, unknown>
  request_body_template: Record<string, unknown>
  auth_type: ProjectAuthType
  auth_config: Record<string, unknown>
  capability_summary?: string
  created_at: string
  updated_at: string
  last_published_at?: string
  integrations: ProjectIntegration[]
  capabilities: ProjectCapability[]
}

export interface ProjectPayload {
  name: string
  description: string
  status: ProjectStatus
  source_type: ProjectSourceType
  endpoint_url: string
  request_method: string
  request_path: string
  request_headers: Record<string, unknown>
  request_body_template: Record<string, unknown>
  auth_type: ProjectAuthType
  auth_config: Record<string, unknown>
  capability_summary: string
}

export interface ProjectFormState {
  name: string
  description: string
  status: ProjectStatus
  source_type: ProjectSourceType
  endpoint_url: string
  request_method: string
  request_path: string
  request_headers: string
  request_body_template: string
  auth_type: ProjectAuthType
  auth_config: string
  capability_summary: string
}

export const emptyProjectForm: ProjectFormState = {
  name: '',
  description: '',
  status: 'offline',
  source_type: 'mixed',
  endpoint_url: '',
  request_method: 'GET',
  request_path: '',
  request_headers: '{}',
  request_body_template: '{}',
  auth_type: 'none',
  auth_config: '{}',
  capability_summary: '',
}

export function projectToForm(project: Project): ProjectFormState {
  return {
    name: project.name,
    description: project.description || '',
    status: project.status,
    source_type: project.source_type,
    endpoint_url: project.endpoint_url || '',
    request_method: project.request_method || 'GET',
    request_path: project.request_path || '',
    request_headers: JSON.stringify(project.request_headers || {}, null, 2),
    request_body_template: JSON.stringify(project.request_body_template || {}, null, 2),
    auth_type: project.auth_type || 'none',
    auth_config: JSON.stringify(project.auth_config || {}, null, 2),
    capability_summary: project.capability_summary || '',
  }
}

function parseJSONObject(value: string): Record<string, unknown> {
  if (!value.trim()) return {}
  const parsed = JSON.parse(value)
  if (!parsed || Array.isArray(parsed) || typeof parsed !== 'object') {
    throw new Error('JSON fields must be objects')
  }
  return parsed as Record<string, unknown>
}

export function formToPayload(form: ProjectFormState): ProjectPayload {
  return {
    name: form.name.trim(),
    description: form.description.trim(),
    status: form.status,
    source_type: form.source_type,
    endpoint_url: form.endpoint_url.trim(),
    request_method: form.request_method.trim() || 'GET',
    request_path: form.request_path.trim(),
    request_headers: parseJSONObject(form.request_headers),
    request_body_template: parseJSONObject(form.request_body_template),
    auth_type: form.auth_type,
    auth_config: parseJSONObject(form.auth_config),
    capability_summary: form.capability_summary.trim(),
  }
}

export interface IntegrationFormState {
  kind: ProjectIntegrationKind
  name: string
  description: string
  status: 'active' | 'inactive'
  endpoint_url: string
  documentation_url: string
  auth_type: ProjectIntegrationAuthType
  request_headers: string
  auth_config: string
  metadata: string
}

export const emptyIntegrationForm: IntegrationFormState = {
  kind: 'mcp',
  name: '',
  description: '',
  status: 'active',
  endpoint_url: '',
  documentation_url: '',
  auth_type: 'inherit',
  request_headers: '{}',
  auth_config: '{}',
  metadata: '{}',
}

export function integrationToForm(integration: ProjectIntegration): IntegrationFormState {
  return {
    kind: integration.kind,
    name: integration.name,
    description: integration.description || '',
    status: integration.status || 'active',
    endpoint_url: integration.endpoint_url || '',
    documentation_url: integration.documentation_url || '',
    auth_type: integration.auth_type || 'inherit',
    request_headers: JSON.stringify(integration.request_headers || {}, null, 2),
    auth_config: JSON.stringify(integration.auth_config || {}, null, 2),
    metadata: JSON.stringify(integration.metadata || {}, null, 2),
  }
}

export function integrationFormToPayload(form: IntegrationFormState): ProjectIntegration {
  return {
    kind: form.kind,
    name: form.name.trim(),
    description: form.description.trim(),
    status: form.status,
    endpoint_url: form.endpoint_url.trim(),
    documentation_url: form.documentation_url.trim(),
    transport: 'http',
    auth_type: form.auth_type,
    request_headers: parseJSONObject(form.request_headers),
    auth_config: parseJSONObject(form.auth_config),
    metadata: parseJSONObject(form.metadata),
  }
}
