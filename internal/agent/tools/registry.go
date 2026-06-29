package tools

import (
	"fmt"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/llm"
)

// Registry 是单次 AgentRun 的工具注册表。
type Registry struct {
	tools map[string]Tool
	god   *domain.GodScope
}

// NewRegistry 创建新的工具注册表。
func NewRegistry(godScope *domain.GodScope) *Registry {
	return &Registry{
		tools: make(map[string]Tool),
		god:   godScope,
	}
}

// Register 注册一个工具。
func (r *Registry) Register(t Tool) error {
	if r.god != nil {
		if err := r.god.CanExecute(t.Name(), t.Domain()); err != nil {
			return err
		}
	}
	r.tools[t.Name()] = t
	return nil
}

// Get 获取工具。
func (r *Registry) Get(name string) (Tool, bool) {
	t, ok := r.tools[name]
	return t, ok
}

// List 列出所有工具。
func (r *Registry) List() []Tool {
	list := make([]Tool, 0, len(r.tools))
	for _, t := range r.tools {
		list = append(list, t)
	}
	return list
}

// ToLLMDefinitions 转换为 LLM 工具定义。
func (r *Registry) ToLLMDefinitions() []llm.ToolDefinition {
	defs := make([]llm.ToolDefinition, 0, len(r.tools))
	for _, t := range r.tools {
		defs = append(defs, llm.ToolDefinition{
			Type:        "function",
			Name:        t.Name(),
			Description: t.Description(),
			Parameters:  t.InputSchema(),
		})
	}
	return defs
}

// MustRegister 注册工具，失败时 panic（用于内置工具注册）。
func (r *Registry) MustRegister(t Tool) {
	if err := r.Register(t); err != nil {
		panic(fmt.Sprintf("register tool %s: %v", t.Name(), err))
	}
}
