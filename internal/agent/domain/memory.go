package domain

import (
	"time"

	"github.com/google/uuid"
)

// MemoryCategory 定义记忆分类。
type MemoryCategory string

const (
	MemoryCategoryPreference MemoryCategory = "preference"
	MemoryCategoryProject    MemoryCategory = "project"
	MemoryCategoryFact       MemoryCategory = "fact"
	MemoryCategorySkill      MemoryCategory = "skill"
)

// Memory 表示从对话中沉淀下来的用户偏好、项目约定或业务事实。
type Memory struct {
	ID              uuid.UUID
	UserID          uuid.UUID
	Category        MemoryCategory
	Content         string
	SourceSessionID *uuid.UUID
	SourceMessageID *uuid.UUID
	Confidence      float64
	ExpiresAt       *time.Time
	Metadata        map[string]any
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// IsExpired 判断记忆是否已过期。
func (m *Memory) IsExpired(now time.Time) bool {
	return m.ExpiresAt != nil && now.After(*m.ExpiresAt)
}
