package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGodScope_CanExecute(t *testing.T) {
	scope := &GodScope{
		AllowedDomains:   []string{"travel", "analytics"},
		ForbiddenDomains: []string{"admin"},
		AllowedTools:     []string{"search_hotels", "read_file"},
		ForbiddenTools:   []string{"delete_database"},
	}

	err := scope.CanExecute("search_hotels", "travel")
	assert.NoError(t, err)

	err = scope.CanExecute("delete_database", "travel")
	assert.Error(t, err)

	err = scope.CanExecute("search_hotels", "admin")
	assert.Error(t, err)

	err = scope.CanExecute("unknown_tool", "travel")
	assert.Error(t, err)
}

func TestGodScope_RequiresApproval(t *testing.T) {
	scope := &GodScope{
		RequireApprovalTools: []string{"terminal"},
	}

	assert.True(t, scope.RequiresApproval("terminal"))
	assert.False(t, scope.RequiresApproval("read_file"))
}
