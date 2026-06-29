package domain

import (
	"time"

	"github.com/google/uuid"
)

// SessionStatus 定义会话状态。
type SessionStatus string

const (
	SessionStatusActive   SessionStatus = "active"
	SessionStatusArchived SessionStatus = "archived"
	SessionStatusDeleted  SessionStatus = "deleted"
)

// MessageRole 定义消息角色。
type MessageRole string

const (
	MessageRoleSystem    MessageRole = "system"
	MessageRoleUser      MessageRole = "user"
	MessageRoleAssistant MessageRole = "assistant"
	MessageRoleTool      MessageRole = "tool"
)

// Session 表示一次完整的 Agent 对话上下文。
type Session struct {
	ID               uuid.UUID
	UserID           uuid.UUID
	TenantID         *uuid.UUID
	Title            string
	Status           SessionStatus
	ParentSessionID  *uuid.UUID
	ContextSummary   string
	ContextSummaryAt *time.Time
	Metadata         map[string]any
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsActive 判断会话是否处于活跃状态。
func (s *Session) IsActive() bool {
	return s.Status == SessionStatusActive
}

// ToolCall 表示 assistant 发起的一次工具调用。
type ToolCall struct {
	ID       string
	Type     string
	Function ToolCallFunction
}

// ToolCallFunction 包含工具调用的函数名和参数。
type ToolCallFunction struct {
	Name      string
	Arguments string
}

// Message 表示会话中的一条消息。
type Message struct {
	ID               uuid.UUID
	SessionID        uuid.UUID
	Role             MessageRole
	Content          string
	ToolCalls        []ToolCall
	ToolCallID       string
	ToolName         string
	ToolResult       map[string]any
	ReasoningContent string
	TokenCount       int
	Metadata         map[string]any
	CreatedAt        time.Time
}

// IsUserMessage 判断是否为用户消息。
func (m *Message) IsUserMessage() bool {
	return m.Role == MessageRoleUser
}

// IsToolMessage 判断是否为工具结果消息。
func (m *Message) IsToolMessage() bool {
	return m.Role == MessageRoleTool
}
