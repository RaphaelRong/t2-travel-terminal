package runtime

import (
	"encoding/json"
	"fmt"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
)

// GuardrailDecision 表示护栏决策。
type GuardrailDecision struct {
	Action  string // "allow", "warn", "halt"
	Message string
}

// Guardrail 实现工具调用护栏逻辑。
type Guardrail struct {
	exactFailureThreshold  int
	sameToolFailureThreshold int
	noProgressThreshold    int
}

// NewGuardrail 创建新的 Guardrail。
func NewGuardrail() *Guardrail {
	return &Guardrail{
		exactFailureThreshold:    3,
		sameToolFailureThreshold: 5,
		noProgressThreshold:      3,
	}
}

// AfterCall 在工具调用后评估是否需要警告或阻断。
func (g *Guardrail) AfterCall(controller *domain.GuardrailController, toolName string, args map[string]any, result any, failed bool) GuardrailDecision {
	if controller == nil {
		return GuardrailDecision{Action: "allow"}
	}

	sig := signature(toolName, args)

	if failed {
		controller.ExactFailureCounts[sig]++
		controller.SameToolFailureCounts[toolName]++
	} else {
		controller.ExactFailureCounts[sig] = 0
		controller.SameToolFailureCounts[toolName] = 0
	}

	if controller.ExactFailureCounts[sig] >= g.exactFailureThreshold {
		return GuardrailDecision{
			Action:  "halt",
			Message: fmt.Sprintf("Tool %s called with the same arguments too many times (%d). Change strategy.", toolName, controller.ExactFailureCounts[sig]),
		}
	}
	if controller.ExactFailureCounts[sig] >= g.exactFailureThreshold-1 {
		return GuardrailDecision{
			Action:  "warn",
			Message: fmt.Sprintf("Warning: repeated failures for %s with same arguments.", toolName),
		}
	}
	if controller.SameToolFailureCounts[toolName] >= g.sameToolFailureThreshold {
		return GuardrailDecision{
			Action:  "halt",
			Message: fmt.Sprintf("Tool %s has failed %d times. Change strategy.", toolName, controller.SameToolFailureCounts[toolName]),
		}
	}

	return GuardrailDecision{Action: "allow"}
}

func signature(toolName string, args map[string]any) domain.ToolSignature {
	b, _ := json.Marshal(args)
	return domain.ToolSignature{
		Name: toolName,
		Args: string(b),
	}
}
