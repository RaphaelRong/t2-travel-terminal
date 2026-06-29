package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ProjectCapabilityKind 表示项目能力的类型。
type ProjectCapabilityKind string

const (
	ProjectCapabilityKindAPI   ProjectCapabilityKind = "api"
	ProjectCapabilityKindTool  ProjectCapabilityKind = "tool"
	ProjectCapabilityKindSkill ProjectCapabilityKind = "skill"
)

// Project 是 T2 中已有的项目/数据源的精简视图。
type Project struct {
	ID                  uuid.UUID
	TenantID            *uuid.UUID
	SourceScope         string
	Kind                string
	Status              string
	SourceType          string
	Name                string
	Description         string
	EndpointURL         string
	RequestMethod       string
	RequestPath         string
	RequestHeaders      map[string]string
	RequestBodyTemplate map[string]interface{}
	AuthType            string
	AuthConfig          map[string]string
	CapabilitySummary   string
	CreatedBy           *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastPublishedAt     *time.Time
}

// ProjectIntegration 是项目的能力来源（MCP / API / Skill）。
type ProjectIntegration struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	Kind             string
	Name             string
	Description      string
	Status           string
	EndpointURL      string
	DocumentationURL string
	Transport        string
	AuthType         string
	RequestHeaders   map[string]string
	AuthConfig       map[string]string
	Metadata         map[string]interface{}
	LastSyncedAt     *time.Time
	SyncStatus       string
	SyncError        string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// ProjectCapability 是项目对外暴露的一个能力/工具。
type ProjectCapability struct {
	ID            uuid.UUID
	ProjectID     uuid.UUID
	IntegrationID *uuid.UUID
	Kind          ProjectCapabilityKind
	Name          string
	ExternalName  string
	Description   string
	Status        string
	RequestMethod string
	RequestPath   string
	InputSchema   map[string]interface{}
	OutputSchema  map[string]interface{}
	Metadata      map[string]interface{}
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// RawAuthConfig 返回 JSON 形式的 auth_config，便于复用现有 HTTP 工具逻辑。
func (p *Project) RawAuthConfig() json.RawMessage {
	b, _ := json.Marshal(p.AuthConfig)
	return b
}

// RawAuthConfig 返回 JSON 形式的 auth_config。
func (i *ProjectIntegration) RawAuthConfig() json.RawMessage {
	b, _ := json.Marshal(i.AuthConfig)
	return b
}

// RawRequestHeaders 返回 JSON 形式的 request_headers。
func (p *Project) RawRequestHeaders() json.RawMessage {
	b, _ := json.Marshal(p.RequestHeaders)
	return b
}

// RawRequestHeaders 返回 JSON 形式的 request_headers。
func (i *ProjectIntegration) RawRequestHeaders() json.RawMessage {
	b, _ := json.Marshal(i.RequestHeaders)
	return b
}
