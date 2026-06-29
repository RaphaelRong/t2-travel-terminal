package domain

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestPlan_FindNextTask(t *testing.T) {
	task1ID := uuid.New()
	task2ID := uuid.New()

	plan := Plan{
		ID: uuid.New(),
		Tasks: []Task{
			{ID: task1ID, Status: TaskStatusPending},
			{ID: task2ID, Status: TaskStatusPending, Deps: []uuid.UUID{task1ID}},
		},
	}

	next := plan.FindNextTask()
	assert.NotNil(t, next)
	assert.Equal(t, task1ID, next.ID)

	plan.Tasks[0].Status = TaskStatusCompleted
	next = plan.FindNextTask()
	assert.NotNil(t, next)
	assert.Equal(t, task2ID, next.ID)
}

func TestPlan_FindNextTask_NoPending(t *testing.T) {
	task1 := Task{ID: uuid.New(), Status: TaskStatusCompleted}
	plan := Plan{ID: uuid.New(), Tasks: []Task{task1}}

	assert.Nil(t, plan.FindNextTask())
}
