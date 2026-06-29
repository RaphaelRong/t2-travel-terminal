package domain

import (
	"time"

	"github.com/google/uuid"
)

// Scope 定义 Soul 的作用范围。
type SoulScope string

const (
	SoulScopeSystem SoulScope = "system"
	SoulScopeUser   SoulScope = "user"
)

// Soul 表示 Agent 的人格、沟通风格和价值观约束。
type Soul struct {
	ID               uuid.UUID
	Scope            SoulScope
	UserID           *uuid.UUID // user scope 时必填
	Name             string
	IdentityText     string
	VoiceText        string
	ValuesText       string
	AllowedDomains   []string
	ForbiddenDomains []string
	Metadata         map[string]any
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IsUserScope 判断是否为用户的私有 Soul。
func (s *Soul) IsUserScope() bool {
	return s.Scope == SoulScopeUser
}
