package god

import (
	"strings"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
)

// BuildSystemPrompt 将 GodScope 的规则转换为注入 LLM system prompt 的文本。
func BuildSystemPrompt(scope *domain.GodScope, soul *domain.Soul) string {
	var parts []string

	parts = append(parts, "# God Scope Rules")
	parts = append(parts, "The following rules define what you can and cannot do. Follow them strictly.")
	parts = append(parts, "")
	parts = append(parts, "## Workflow Execution")
	parts = append(parts, "- For requests that require decomposition, parallel branches, isolated subtasks, staged execution, or deterministic result comparison, use workflow_tool to plan the workflow before acting.")
	parts = append(parts, "- Keep isolated workflow branches from seeing each other's intermediate outputs until the merge/comparison step.")
	parts = append(parts, "- Do not create database tables, datasets, or SQL state for short-lived workflow planning unless the user explicitly asks to persist data.")

	if len(scope.Rules) > 0 {
		parts = append(parts, "")
		parts = append(parts, "## Directives")
		for _, rule := range scope.Rules {
			parts = append(parts, "- "+rule)
		}
	}

	if len(scope.ForbiddenDomains) > 0 {
		parts = append(parts, "")
		parts = append(parts, "## Forbidden Domains")
		for _, d := range scope.ForbiddenDomains {
			parts = append(parts, "- "+d)
		}
	}

	if len(scope.ForbiddenTools) > 0 {
		parts = append(parts, "")
		parts = append(parts, "## Forbidden Tools")
		for _, t := range scope.ForbiddenTools {
			parts = append(parts, "- "+t)
		}
	}

	if len(scope.RequireApprovalTools) > 0 {
		parts = append(parts, "")
		parts = append(parts, "## Tools Requiring User Approval")
		for _, t := range scope.RequireApprovalTools {
			parts = append(parts, "- "+t)
		}
	}

	if scope.MaxIterations > 0 {
		parts = append(parts, "")
		parts = append(parts, "## Iteration Budget")
		parts = append(parts, "You have at most "+itoa(scope.MaxIterations)+" tool-calling iterations for this request.")
	}

	if soul != nil {
		parts = append(parts, "")
		parts = append(parts, "# Soul")
		if soul.IdentityText != "" {
			parts = append(parts, soul.IdentityText)
		}
		if soul.VoiceText != "" {
			parts = append(parts, "")
			parts = append(parts, "## Voice")
			parts = append(parts, soul.VoiceText)
		}
		if soul.ValuesText != "" {
			parts = append(parts, "")
			parts = append(parts, "## Values")
			parts = append(parts, soul.ValuesText)
		}
	}

	return strings.Join(parts, "\n")
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var result []byte
	negative := n < 0
	if negative {
		n = -n
	}
	for n > 0 {
		result = append([]byte{byte('0' + n%10)}, result...)
		n /= 10
	}
	if negative {
		result = append([]byte{'-'}, result...)
	}
	return string(result)
}
