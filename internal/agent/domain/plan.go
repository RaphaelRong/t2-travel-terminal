package domain

import "github.com/google/uuid"

// TaskStatus 定义任务状态。
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusInProgress TaskStatus = "in_progress"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusBlocked    TaskStatus = "blocked"
)

// Task 是 Plan 中的单个可执行任务。
type Task struct {
	ID       uuid.UUID
	Goal     string
	Context  string
	Status   TaskStatus
	Deps     []uuid.UUID // 依赖的其他任务 ID
	Result   map[string]any
	Metadata map[string]any
}

// Plan 表示一次用户请求被拆解后的执行计划。
type Plan struct {
	ID          uuid.UUID
	SessionID   uuid.UUID
	UserMessage string
	Tasks       []Task
	Metadata    map[string]any
}

// FindNextTask 返回下一个可执行的任务（pending 且依赖都已完成）。
func (p *Plan) FindNextTask() *Task {
	completed := make(map[uuid.UUID]bool)
	for i := range p.Tasks {
		if p.Tasks[i].Status == TaskStatusCompleted {
			completed[p.Tasks[i].ID] = true
		}
	}

	for i := range p.Tasks {
		if p.Tasks[i].Status != TaskStatusPending {
			continue
		}
		ready := true
		for _, depID := range p.Tasks[i].Deps {
			if !completed[depID] {
				ready = false
				break
			}
		}
		if ready {
			return &p.Tasks[i]
		}
	}
	return nil
}
