export type LLMProvider = 'openai' | 'anthropic' | 'google' | 'custom'
export type LLMProfileStatus = 'active' | 'inactive'

export interface LLMProfile {
  id: string
  provider: LLMProvider
  display_name: string
  base_url: string
  default_model: string
  models: string[]
  status: LLMProfileStatus
  configured: boolean
  created_at: string
  updated_at: string
}

export interface LLMProfilePayload {
  provider: LLMProvider
  display_name: string
  base_url: string
  api_key: string
  default_model: string
  models: string[]
  status: LLMProfileStatus
}

export interface LLMProfileFormState {
  id?: string
  provider: LLMProvider
  display_name: string
  base_url: string
  api_key: string
  default_model: string
  models: string
  status: LLMProfileStatus
}

export const providerLabels: Record<LLMProvider, string> = {
  openai: 'OpenAI',
  anthropic: 'Anthropic',
  google: 'Google',
  custom: 'Custom',
}

export const providerModelHints: Record<LLMProvider, string[]> = {
  openai: ['gpt-4.1', 'gpt-4.1-mini', 'gpt-4o'],
  anthropic: ['claude-3-5-sonnet-latest', 'claude-3-5-haiku-latest'],
  google: ['gemini-1.5-pro', 'gemini-1.5-flash'],
  custom: [],
}

export const emptyLLMProfileForm: LLMProfileFormState = {
  provider: 'openai',
  display_name: '',
  base_url: '',
  api_key: '',
  default_model: '',
  models: '',
  status: 'active',
}

export function profileToForm(profile: LLMProfile): LLMProfileFormState {
  return {
    id: profile.id,
    provider: profile.provider,
    display_name: profile.display_name,
    base_url: profile.base_url || '',
    api_key: '',
    default_model: profile.default_model || '',
    models: profile.models.join('\n'),
    status: profile.status,
  }
}

export function formToLLMProfilePayload(form: LLMProfileFormState): LLMProfilePayload {
  const models = form.models
    .split(/[\n,]/)
    .map((item) => item.trim())
    .filter(Boolean)

  return {
    provider: form.provider,
    display_name: form.display_name.trim(),
    base_url: form.base_url.trim(),
    api_key: form.api_key.trim(),
    default_model: form.default_model.trim(),
    models,
    status: form.status,
  }
}
