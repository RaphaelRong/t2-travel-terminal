// Package store 提供 Agent 领域对象的数据访问接口。
package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
)

// Querier 是 pgx 查询接口的最小抽象。
// *pgxpool.Conn 和 pgx.Tx 都满足此接口。
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// SoulStore 管理 Soul 记录。
type SoulStore interface {
	Create(ctx context.Context, q Querier, soul *domain.Soul) error
	Update(ctx context.Context, q Querier, soul *domain.Soul) error
	Delete(ctx context.Context, q Querier, id uuid.UUID) error
	GetByID(ctx context.Context, q Querier, id uuid.UUID) (*domain.Soul, error)
	ListByUser(ctx context.Context, q Querier, userID uuid.UUID) ([]*domain.Soul, error)
	ListSystemSouls(ctx context.Context, q Querier) ([]*domain.Soul, error)
}

// MemoryStore 管理记忆记录。
type MemoryStore interface {
	Create(ctx context.Context, q Querier, memory *domain.Memory) error
	Update(ctx context.Context, q Querier, memory *domain.Memory) error
	Delete(ctx context.Context, q Querier, id uuid.UUID) error
	GetByID(ctx context.Context, q Querier, id uuid.UUID) (*domain.Memory, error)
	ListByUser(ctx context.Context, q Querier, userID uuid.UUID, limit int) ([]*domain.Memory, error)
	Search(ctx context.Context, q Querier, userID uuid.UUID, query string, limit int) ([]*domain.Memory, error)
}

// SessionStore 管理会话和消息。
type SessionStore interface {
	CreateSession(ctx context.Context, q Querier, session *domain.Session) error
	UpdateSession(ctx context.Context, q Querier, session *domain.Session) error
	GetSessionByID(ctx context.Context, q Querier, id uuid.UUID) (*domain.Session, error)
	ListSessionsByUser(ctx context.Context, q Querier, userID uuid.UUID, status domain.SessionStatus, limit int) ([]*domain.Session, error)

	CreateMessage(ctx context.Context, q Querier, message *domain.Message) error
	ListMessagesBySession(ctx context.Context, q Querier, sessionID uuid.UUID, limit int) ([]*domain.Message, error)

	AttachProject(ctx context.Context, q Querier, sessionID, projectID uuid.UUID) error
	DetachProject(ctx context.Context, q Querier, sessionID, projectID uuid.UUID) error
	ListProjectIDsBySession(ctx context.Context, q Querier, sessionID uuid.UUID) ([]uuid.UUID, error)
}

// ProfileStore 管理用户 Agent 配置。
type ProfileStore interface {
	CreateOrUpdate(ctx context.Context, q Querier, profile *domain.UserProfile) error
	GetByUserID(ctx context.Context, q Querier, userID uuid.UUID) (*domain.UserProfile, error)
}

// GodConfigStore 管理 God 全局配置。
type GodConfigStore interface {
	Create(ctx context.Context, q Querier, config *domain.GodConfig) error
	Update(ctx context.Context, q Querier, config *domain.GodConfig) error
	GetByName(ctx context.Context, q Querier, name string) (*domain.GodConfig, error)
	GetActive(ctx context.Context, q Querier) (*domain.GodConfig, error)
	List(ctx context.Context, q Querier) ([]*domain.GodConfig, error)
}

// ProjectStore 是 Agent 视角的 Project / Capability 只读访问接口。
type ProjectStore interface {
	ListBySession(ctx context.Context, q Querier, sessionID uuid.UUID) ([]*domain.Project, error)
	ListCapabilitiesBySession(ctx context.Context, q Querier, sessionID uuid.UUID) ([]*domain.ProjectCapability, []*domain.Project, []*domain.ProjectIntegration, error)
	GetCapabilityByID(ctx context.Context, q Querier, capabilityID uuid.UUID) (*domain.ProjectCapability, *domain.Project, *domain.ProjectIntegration, error)
	GetProjectByID(ctx context.Context, q Querier, projectID uuid.UUID) (*domain.Project, error)
	GetIntegrationByID(ctx context.Context, q Querier, integrationID uuid.UUID) (*domain.ProjectIntegration, error)
}

// UserDataStore 管理用户业务数据集。
type UserDataStore interface {
	CreateDataset(ctx context.Context, q Querier, dataset *domain.UserDataset) error
	UpdateDataset(ctx context.Context, q Querier, dataset *domain.UserDataset) error
	DeleteDataset(ctx context.Context, q Querier, id uuid.UUID) error
	GetDatasetByID(ctx context.Context, q Querier, id uuid.UUID) (*domain.UserDataset, error)
	GetDatasetByName(ctx context.Context, q Querier, userID uuid.UUID, name string) (*domain.UserDataset, error)
	ListDatasetsByUser(ctx context.Context, q Querier, userID uuid.UUID, limit int) ([]*domain.UserDataset, error)

	InsertRows(ctx context.Context, q Querier, datasetID uuid.UUID, rows []map[string]any) error
	Query(ctx context.Context, q Querier, userID uuid.UUID, sql string, args ...any) ([]map[string]any, error)
}
