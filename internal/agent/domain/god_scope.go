package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

// GodConfig 是系统级全局范围配置，仅 SuperAdmin 可管理。
type GodConfig struct {
	ID                   uuid.UUID
	Name                 string
	IsActive             bool
	AllowedDomains       []string
	ForbiddenDomains     []string
	AllowedTools         []string
	ForbiddenTools       []string
	RequireApprovalTools []string
	MaxIterations        int
	CanDelegate          bool
	CanRunWorkflow       bool
	Rules                string
	CreatedAt            time.Time
	UpdatedAt            time.Time
}

// GodScope 是运行时的有效范围，由 GodConfig + Soul 合并生成。
type GodScope struct {
	AllowedDomains       []string
	ForbiddenDomains     []string
	AllowedTools         []string
	ForbiddenTools       []string
	RequireApprovalTools []string
	MaxIterations        int
	CanDelegate          bool
	CanRunWorkflow       bool
	Rules                []string
}

// CanExecute 校验指定工具是否允许执行。
// 禁止优先于允许。
func (s *GodScope) CanExecute(toolName string, domain string) error {
	if s.isForbiddenTool(toolName) {
		return fmt.Errorf("tool %q is forbidden by God scope", toolName)
	}
	if s.isForbiddenDomain(domain) {
		return fmt.Errorf("domain %q is forbidden by God scope", domain)
	}
	if len(s.AllowedTools) > 0 && !contains(s.AllowedTools, toolName) {
		return fmt.Errorf("tool %q is not in the allowed tool list", toolName)
	}
	if len(s.AllowedDomains) > 0 && !contains(s.AllowedDomains, domain) {
		return fmt.Errorf("domain %q is not in the allowed domain list", domain)
	}
	return nil
}

// RequiresApproval 判断工具是否需要用户审批。
func (s *GodScope) RequiresApproval(toolName string) bool {
	return contains(s.RequireApprovalTools, toolName)
}

func (s *GodScope) isForbiddenTool(toolName string) bool {
	return contains(s.ForbiddenTools, toolName)
}

func (s *GodScope) isForbiddenDomain(domain string) bool {
	return contains(s.ForbiddenDomains, domain)
}

func contains(list []string, item string) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}
