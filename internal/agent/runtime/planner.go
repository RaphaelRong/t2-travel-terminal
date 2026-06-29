package runtime

import (
	"context"

	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/llm"
)

// Planner 负责任务规划。
type Planner struct{}

// NewPlanner 创建 Planner。
func NewPlanner() *Planner {
	return &Planner{}
}

// PlanRequest 是规划请求。
type PlanRequest struct {
	UserMessage string
	History     []llm.Message
}

// Plan 根据用户输入和上下文生成一个简单计划。
// Phase 3 先返回单任务计划，后续可接入 LLM 做复杂拆解。
func (p *Planner) Plan(ctx context.Context, req PlanRequest) (*domain.Plan, error) {
	plan := &domain.Plan{
		ID:          uuid.New(),
		UserMessage: req.UserMessage,
		Tasks: []domain.Task{
			{
				ID:     uuid.New(),
				Goal:   req.UserMessage,
				Status: domain.TaskStatusInProgress,
			},
		},
	}
	return plan, nil
}
