package tools

import (
	"context"
	"fmt"
	"strings"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/store"
)

// WorkflowTool 帮助 Agent 将需要拆分、隔离、汇总的任务表达为受控流程。
type WorkflowTool struct{}

// NewWorkflowTool 创建 workflow_tool。
func NewWorkflowTool() *WorkflowTool {
	return &WorkflowTool{}
}

func (t *WorkflowTool) Name() string   { return "workflow_tool" }
func (t *WorkflowTool) Domain() string { return "workflow" }
func (t *WorkflowTool) Description() string {
	return "Plan and explain a multi-step workflow. Use this for tasks that need decomposition, isolated branches, parallel subtasks, or deterministic result comparison. Do not create database tables for short-lived workflow state."
}
func (t *WorkflowTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"goal": map[string]interface{}{
				"type":        "string",
				"description": "The user's workflow goal.",
			},
			"strategy": map[string]interface{}{
				"type":        "string",
				"description": "How the workflow should be executed and coordinated.",
			},
			"steps": map[string]interface{}{
				"type": "array",
				"items": map[string]interface{}{
					"type": "object",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":        "string",
							"description": "Short step name.",
						},
						"instruction": map[string]interface{}{
							"type":        "string",
							"description": "What this step should do.",
						},
						"isolation": map[string]interface{}{
							"type":        "string",
							"enum":        []string{"shared", "isolated"},
							"description": "Use isolated when this step must not see sibling step outputs.",
						},
						"expected_output": map[string]interface{}{
							"type":        "string",
							"description": "The expected output shape or value type.",
						},
					},
					"required": []string{"name", "instruction"},
				},
			},
			"comparison": map[string]interface{}{
				"type":        "string",
				"description": "Deterministic comparison or merge rule to apply after the steps complete.",
			},
		},
		"required": []string{"goal", "steps"},
	}
}

func (t *WorkflowTool) Execute(ctx context.Context, args map[string]interface{}, runCtx *Context) (interface{}, error) {
	goal := getString(args, "goal", "")
	if goal == "" {
		return nil, fmt.Errorf("goal is required")
	}

	rawSteps, ok := args["steps"].([]interface{})
	if !ok || len(rawSteps) == 0 {
		return nil, fmt.Errorf("at least one workflow step is required")
	}

	steps := make([]map[string]interface{}, 0, len(rawSteps))
	for i, raw := range rawSteps {
		step, ok := raw.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("step %d must be an object", i+1)
		}
		name := strings.TrimSpace(getString(step, "name", fmt.Sprintf("step_%d", i+1)))
		instruction := strings.TrimSpace(getString(step, "instruction", ""))
		if instruction == "" {
			return nil, fmt.Errorf("step %d instruction is required", i+1)
		}
		isolation := getString(step, "isolation", "shared")
		if isolation != "isolated" {
			isolation = "shared"
		}
		steps = append(steps, map[string]interface{}{
			"index":           i + 1,
			"name":            name,
			"instruction":     instruction,
			"isolation":       isolation,
			"expected_output": getString(step, "expected_output", ""),
			"status":          "planned",
		})
	}

	return map[string]interface{}{
		"workflow_id": fmt.Sprintf("wf_%s_%d", runCtx.SessionID.String()[:8], len(steps)),
		"goal":        goal,
		"strategy":    getString(args, "strategy", "Decompose into explicit steps, keep isolated steps independent, then merge or compare results deterministically."),
		"steps":       steps,
		"comparison":  getString(args, "comparison", ""),
		"next_action": "Execute each planned step in order. Do not use SQL or persistent tables unless the user explicitly asks to store workflow state.",
	}, nil
}

// MemoryTool 允许 Agent 将信息写入用户记忆库。
type MemoryTool struct {
	memoryStore store.MemoryStore
}

// NewMemoryTool 创建 memory_tool。
func NewMemoryTool(memoryStore store.MemoryStore) *MemoryTool {
	return &MemoryTool{memoryStore: memoryStore}
}

func (t *MemoryTool) Name() string   { return "memory_tool" }
func (t *MemoryTool) Domain() string { return "memory" }
func (t *MemoryTool) Description() string {
	return "Store a fact, preference, or project convention into the user's memory."
}
func (t *MemoryTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"category": map[string]interface{}{
				"type":        "string",
				"enum":        []string{"preference", "project", "fact", "skill"},
				"description": "Category of the memory.",
			},
			"content": map[string]interface{}{
				"type":        "string",
				"description": "The content to remember.",
			},
		},
		"required": []string{"category", "content"},
	}
}

func (t *MemoryTool) Execute(ctx context.Context, args map[string]interface{}, runCtx *Context) (interface{}, error) {
	category := domain.MemoryCategory(getString(args, "category", "fact"))
	content := getString(args, "content", "")
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	memory := &domain.Memory{
		UserID:   runCtx.UserID,
		Category: category,
		Content:  content,
	}
	if err := t.memoryStore.Create(ctx, runCtx.Querier, memory); err != nil {
		return nil, err
	}
	return map[string]interface{}{"memory_id": memory.ID.String()}, nil
}

// ExecuteSQLTool 允许 Agent 对用户数据集执行受控 SQL。
type ExecuteSQLTool struct {
	userdataStore store.UserDataStore
}

// NewExecuteSQLTool 创建 execute_sql 工具。
func NewExecuteSQLTool(userdataStore store.UserDataStore) *ExecuteSQLTool {
	return &ExecuteSQLTool{userdataStore: userdataStore}
}

func (t *ExecuteSQLTool) Name() string   { return "execute_sql" }
func (t *ExecuteSQLTool) Domain() string { return "userdata" }
func (t *ExecuteSQLTool) Description() string {
	return "Execute SQL only when the user explicitly asks to query or analyze stored user datasets. Do not use this for short-lived planning, games, workflow state, or general reasoning."
}
func (t *ExecuteSQLTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"query": map[string]interface{}{
				"type":        "string",
				"description": "SQL query to execute. Only operates on user's own data.",
			},
		},
		"required": []string{"query"},
	}
}

func (t *ExecuteSQLTool) Execute(ctx context.Context, args map[string]interface{}, runCtx *Context) (interface{}, error) {
	query := getString(args, "query", "")
	if query == "" {
		return nil, fmt.Errorf("query is required")
	}

	upper := strings.ToUpper(strings.TrimSpace(query))
	forbidden := []string{"DROP DATABASE", "DROP USER", "DELETE FROM agent_user_dataset_rows", "DELETE FROM agent_user_datasets"}
	for _, f := range forbidden {
		if strings.Contains(upper, strings.ToUpper(f)) {
			return nil, fmt.Errorf("query contains forbidden pattern: %s", f)
		}
	}

	return t.userdataStore.Query(ctx, runCtx.Querier, runCtx.UserID, query)
}

// DoneTool 允许 Agent 声明任务已完成。
type DoneTool struct{}

// NewDoneTool 创建 done 工具。
func NewDoneTool() *DoneTool {
	return &DoneTool{}
}

func (t *DoneTool) Name() string   { return "done" }
func (t *DoneTool) Domain() string { return "control" }
func (t *DoneTool) Description() string {
	return "Signal that the task is complete and no further tools are needed."
}
func (t *DoneTool) InputSchema() map[string]interface{} {
	return map[string]interface{}{
		"type": "object",
		"properties": map[string]interface{}{
			"summary": map[string]interface{}{
				"type":        "string",
				"description": "A brief summary of what was accomplished.",
			},
		},
	}
}

func (t *DoneTool) Execute(ctx context.Context, args map[string]interface{}, runCtx *Context) (interface{}, error) {
	return map[string]interface{}{"status": "done", "summary": getString(args, "summary", "")}, nil
}

func getString(m map[string]interface{}, key, defaultValue string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return defaultValue
}
