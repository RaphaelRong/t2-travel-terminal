package tools

import "github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"

// CapabilityAdapter converts persisted Project capabilities into runtime tools.
type CapabilityAdapter interface {
	Adapt(capability *domain.ProjectCapability, project *domain.Project, integration *domain.ProjectIntegration) Tool
}

// ProjectCapabilityAdapter is the default adapter for Project-backed API, Skill, and MCP tools.
type ProjectCapabilityAdapter struct {
	localBaseURL string
}

// NewProjectCapabilityAdapter creates the default Project capability adapter.
func NewProjectCapabilityAdapter(localBaseURL string) *ProjectCapabilityAdapter {
	return &ProjectCapabilityAdapter{localBaseURL: localBaseURL}
}

// Adapt wraps one Project capability as an executable Agent tool.
func (a *ProjectCapabilityAdapter) Adapt(capability *domain.ProjectCapability, project *domain.Project, integration *domain.ProjectIntegration) Tool {
	return NewProjectTool(capability, project, integration, a.localBaseURL)
}
