package runtime

import (
	"context"

	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/store"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/tools"
)

// ProjectLoader 负责把会话关联的 Project Capability 加载为 Tool。
type ProjectLoader struct {
	projectStore store.ProjectStore
	adapter      tools.CapabilityAdapter
}

// NewProjectLoader 创建 ProjectLoader。
func NewProjectLoader(projectStore store.ProjectStore, localBaseURL string) *ProjectLoader {
	return &ProjectLoader{
		projectStore: projectStore,
		adapter:      tools.NewProjectCapabilityAdapter(localBaseURL),
	}
}

// LoadTools 加载指定会话关联的所有 Project Capability 工具。
func (l *ProjectLoader) LoadTools(ctx context.Context, q store.Querier, sessionID uuid.UUID) ([]tools.Tool, error) {
	caps, projects, integrations, err := l.projectStore.ListCapabilitiesBySession(ctx, q, sessionID)
	if err != nil {
		return nil, err
	}

	result := make([]tools.Tool, 0, len(caps))
	for i, cap := range caps {
		project := projects[i]
		integration := integrations[i]
		result = append(result, l.adapter.Adapt(cap, project, integration))
	}
	return result, nil
}
