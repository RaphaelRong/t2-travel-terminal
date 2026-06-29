package domain

import "github.com/google/uuid"

// RunState 定义 AgentRun 的状态。
type RunState string

const (
	RunStatePlanning    RunState = "planning"
	RunStateExecuting   RunState = "executing"
	RunStateWaiting     RunState = "waiting"
	RunStateCompleted   RunState = "completed"
	RunStateError       RunState = "error"
	RunStateInterrupted RunState = "interrupted"
)

// AgentRun 表示一次用户请求的处理实例。
type AgentRun struct {
	SessionID  uuid.UUID
	UserID     uuid.UUID
	TenantID   *uuid.UUID
	Profile    *UserProfile
	Soul       *Soul
	GodScope   *GodScope
	Messages   []Message
	Budget     *IterationBudget
	Guardrails *GuardrailController
	State      RunState
}

// IterationBudget 控制一次 Run 中的最大 LLM 调用次数。
type IterationBudget struct {
	MaxTotal int
	Used     int
}

// Consume 尝试消费一次迭代。
func (b *IterationBudget) Consume() bool {
	if b.Used >= b.MaxTotal {
		return false
	}
	b.Used++
	return true
}

// Remaining 返回剩余迭代次数。
func (b *IterationBudget) Remaining() int {
	remaining := b.MaxTotal - b.Used
	if remaining < 0 {
		return 0
	}
	return remaining
}

// GuardrailController 检测重复/无进展的工具调用。
type GuardrailController struct {
	ExactFailureCounts    map[ToolSignature]int
	SameToolFailureCounts map[string]int
	NoProgressCounts      map[ToolSignature]int
}

// ToolSignature 用于识别重复的工具调用。
type ToolSignature struct {
	Name string
	Args string
}

// NewGuardrailController 创建一个新的护栏控制器。
func NewGuardrailController() *GuardrailController {
	return &GuardrailController{
		ExactFailureCounts:    make(map[ToolSignature]int),
		SameToolFailureCounts: make(map[string]int),
		NoProgressCounts:      make(map[ToolSignature]int),
	}
}
