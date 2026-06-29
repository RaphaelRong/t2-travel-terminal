package llm

import (
	"context"
)

// Provider 是 LLM 调用的统一接口。
type Provider interface {
	// Chat 发送对话请求并返回模型响应。
	Chat(ctx context.Context, model string, messages []Message, tools []ToolDefinition) (*Response, error)
	// Name 返回 provider 名称，如 openai / anthropic / custom。
	Name() string
}

// Profile 是从 user_llm_profiles 读取的配置子集。
type Profile struct {
	ID           string
	Provider     string
	BaseURL      string
	APIKey       string
	DefaultModel string
}

// ExtractAPIKey 从 auth_config jsonb 中提取 api_key。
func ExtractAPIKey(authConfig map[string]any) string {
	if authConfig == nil {
		return ""
	}
	if key, ok := authConfig["api_key"].(string); ok {
		return key
	}
	return ""
}
