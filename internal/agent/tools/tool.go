package tools

import (
	"context"

	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/store"
)

// Context 是工具执行时的上下文。
type Context struct {
	UserID    uuid.UUID
	TenantID  *uuid.UUID
	SessionID uuid.UUID
	Run       *domain.AgentRun
	Querier   store.Querier
}

// Tool 是 Agent 可调用的工具接口。
type Tool interface {
	Name() string
	Description() string
	InputSchema() map[string]any
	Domain() string
	Execute(ctx context.Context, args map[string]any, runCtx *Context) (any, error)
}
