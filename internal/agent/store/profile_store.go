package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/queries"
)

// PGProfileStore 是 UserProfile 的 PostgreSQL 实现。
type PGProfileStore struct{}

// NewPGProfileStore 创建一个新的 PGProfileStore。
func NewPGProfileStore() ProfileStore {
	return &PGProfileStore{}
}

func (st *PGProfileStore) CreateOrUpdate(ctx context.Context, tx Querier, profile *domain.UserProfile) error {
	_, err := tx.Exec(ctx, queries.UserProfileInsert,
		profile.UserID, profile.SoulID, profile.DefaultLLMProfileID, profile.Status,
	)
	return err
}

func (st *PGProfileStore) GetByUserID(ctx context.Context, tx Querier, userID uuid.UUID) (*domain.UserProfile, error) {
	var p domain.UserProfile
	err := tx.QueryRow(ctx, queries.UserProfileSelectByUser, userID).Scan(
		&p.ID, &p.UserID, &p.SoulID, &p.DefaultLLMProfileID, &p.Status, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &p, nil
}
