package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIterationBudget(t *testing.T) {
	budget := &IterationBudget{MaxTotal: 3}

	assert.True(t, budget.Consume())
	assert.True(t, budget.Consume())
	assert.Equal(t, 1, budget.Remaining())

	assert.True(t, budget.Consume())
	assert.False(t, budget.Consume())
	assert.Equal(t, 0, budget.Remaining())
}
