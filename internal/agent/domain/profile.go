package domain

import (
	"time"

	"github.com/google/uuid"
)

// ProfileStatus 定义用户 Agent 配置状态。
type ProfileStatus string

const (
	ProfileStatusActive ProfileStatus = "active"
	ProfileStatusPaused ProfileStatus = "paused"
)

// UserProfile 表示每个用户的 Agent 运行时配置。
type UserProfile struct {
	ID                  uuid.UUID
	UserID              uuid.UUID
	SoulID              *uuid.UUID
	DefaultLLMProfileID *uuid.UUID
	Status              ProfileStatus
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// IsActive 判断配置是否启用。
func (p *UserProfile) IsActive() bool {
	return p.Status == ProfileStatusActive
}
