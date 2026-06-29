// Package god 负责加载和合并 God 全局范围配置。
package god

import (
	"context"

	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/store"
)

// Loader 负责从数据库加载并合并 GodScope。
type Loader struct {
	godStore    store.GodConfigStore
	soulStore   store.SoulStore
}

// NewLoader 创建一个新的 GodScope Loader。
func NewLoader(godStore store.GodConfigStore, soulStore store.SoulStore) *Loader {
	return &Loader{
		godStore:  godStore,
		soulStore: soulStore,
	}
}

// LoadForUser 为指定用户加载有效的 GodScope。
// 合并优先级（低 → 高）：God 全局配置 → 系统默认 Soul → 用户 Soul。
// 禁止优先于允许。
func (l *Loader) LoadForUser(ctx context.Context, q store.Querier, userID uuid.UUID) (*domain.GodScope, error) {
	scope := &domain.GodScope{}

	// 1. God 全局配置
	config, err := l.godStore.GetActive(ctx, q)
	if err == nil && config != nil {
		mergeGodConfig(scope, config)
	}
	// 如果找不到 active config，继续用默认值

	// 2. 系统默认 Soul
	systemSouls, err := l.soulStore.ListSystemSouls(ctx, q)
	if err == nil {
		for _, soul := range systemSouls {
			mergeSoul(scope, soul)
		}
	}

	// 3. 用户 Soul
	userSouls, err := l.soulStore.ListByUser(ctx, q, userID)
	if err == nil {
		for _, soul := range userSouls {
			mergeSoul(scope, soul)
		}
	}

	return scope, nil
}

func mergeGodConfig(scope *domain.GodScope, config *domain.GodConfig) {
	scope.AllowedDomains = mergeLists(scope.AllowedDomains, config.AllowedDomains)
	scope.ForbiddenDomains = mergeLists(scope.ForbiddenDomains, config.ForbiddenDomains)
	scope.AllowedTools = mergeLists(scope.AllowedTools, config.AllowedTools)
	scope.ForbiddenTools = mergeLists(scope.ForbiddenTools, config.ForbiddenTools)
	scope.RequireApprovalTools = mergeLists(scope.RequireApprovalTools, config.RequireApprovalTools)

	if config.MaxIterations > 0 {
		scope.MaxIterations = config.MaxIterations
	}
	scope.CanDelegate = config.CanDelegate
	scope.CanRunWorkflow = config.CanRunWorkflow

	if config.Rules != "" {
		scope.Rules = append(scope.Rules, config.Rules)
	}
}

func mergeSoul(scope *domain.GodScope, soul *domain.Soul) {
	scope.AllowedDomains = mergeLists(scope.AllowedDomains, soul.AllowedDomains)
	scope.ForbiddenDomains = mergeLists(scope.ForbiddenDomains, soul.ForbiddenDomains)

	if soul.IdentityText != "" {
		scope.Rules = append(scope.Rules, soul.IdentityText)
	}
	if soul.VoiceText != "" {
		scope.Rules = append(scope.Rules, soul.VoiceText)
	}
	if soul.ValuesText != "" {
		scope.Rules = append(scope.Rules, soul.ValuesText)
	}
}

func mergeLists(base, incoming []string) []string {
	seen := make(map[string]bool, len(base))
	for _, v := range base {
		seen[v] = true
	}
	for _, v := range incoming {
		if !seen[v] {
			seen[v] = true
			base = append(base, v)
		}
	}
	return base
}

// SystemDefault 返回一个默认的 GodScope，用于数据库未配置时。
func SystemDefault() *domain.GodScope {
	return &domain.GodScope{
		MaxIterations: 30,
		CanDelegate:   false,
		CanRunWorkflow: false,
		Rules: []string{
			"You are a helpful AI assistant for T2 Travel Terminal.",
			"Do not expose secrets, API keys, or private credentials.",
			"Do not execute destructive operations without explicit user confirmation.",
		},
	}
}
