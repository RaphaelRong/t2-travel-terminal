package god

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
)

func TestBuildSystemPrompt(t *testing.T) {
	scope := &domain.GodScope{
		ForbiddenDomains:     []string{"admin"},
		ForbiddenTools:       []string{"delete_database"},
		RequireApprovalTools: []string{"terminal"},
		MaxIterations:        10,
		Rules:                []string{"Be concise."},
	}

	soul := &domain.Soul{
		IdentityText: "You are a travel assistant.",
		VoiceText:    "Professional but friendly.",
	}

	prompt := BuildSystemPrompt(scope, soul)
	assert.Contains(t, prompt, "God Scope Rules")
	assert.Contains(t, prompt, "admin")
	assert.Contains(t, prompt, "delete_database")
	assert.Contains(t, prompt, "terminal")
	assert.Contains(t, prompt, "10")
	assert.Contains(t, prompt, "travel assistant")
	assert.Contains(t, prompt, "Professional but friendly")
}

func TestBuildSystemPrompt_NoSoul(t *testing.T) {
	scope := &domain.GodScope{
		MaxIterations: 5,
	}

	prompt := BuildSystemPrompt(scope, nil)
	assert.Contains(t, prompt, "5")
	assert.NotContains(t, prompt, "# Soul")
}

func TestMergeLists(t *testing.T) {
	result := mergeLists([]string{"a", "b"}, []string{"b", "c"})
	assert.Equal(t, []string{"a", "b", "c"}, result)
}
