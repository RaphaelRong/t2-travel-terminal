package llm

import "github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"

// Message 是对话消息，适配不同 provider 前的统一表示。
type Message struct {
	Role             string
	Content          string
	ToolCalls        []domain.ToolCall
	ToolCallID       string
	ToolName         string
	ToolResult       map[string]any
	ReasoningContent string
}

// ToolDefinition 是暴露给 LLM 的工具定义。
type ToolDefinition struct {
	Type        string         `json:"type"`
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters"`
}

// Response 是 LLM 返回的统一表示。
type Response struct {
	Content          string
	ToolCalls        []domain.ToolCall
	ReasoningContent string
	FinishReason     string
	Usage            Usage
}

// Usage 表示 token 使用情况。
type Usage struct {
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
}

// FinishReason 常量。
const (
	FinishReasonStop      = "stop"
	FinishReasonToolCalls = "tool_calls"
	FinishReasonLength    = "length"
	FinishReasonError     = "error"
)

// Role 常量。
const (
	RoleSystem    = "system"
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleTool      = "tool"
)

// FromDomainMessage 将 domain.Message 转换为 llm.Message。
func FromDomainMessage(m *domain.Message) Message {
	return Message{
		Role:             string(m.Role),
		Content:          m.Content,
		ToolCalls:        m.ToolCalls,
		ToolCallID:       m.ToolCallID,
		ToolName:         m.ToolName,
		ToolResult:       m.ToolResult,
		ReasoningContent: m.ReasoningContent,
	}
}
