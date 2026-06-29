package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/queries"
)

// PGSoulStore 是 Soul 的 PostgreSQL 实现。
type PGSoulStore struct{}

// NewPGSoulStore 创建一个新的 PGSoulStore。
func NewPGSoulStore() SoulStore {
	return &PGSoulStore{}
}

func scanSoul(row pgx.Row) (*domain.Soul, error) {
	var s domain.Soul
	err := row.Scan(
		&s.ID, &s.Scope, &s.UserID, &s.Name, &s.IdentityText, &s.VoiceText,
		&s.ValuesText, &s.AllowedDomains, &s.ForbiddenDomains, &s.Metadata, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (st *PGSoulStore) Create(ctx context.Context, tx Querier, soul *domain.Soul) error {
	_, err := tx.Exec(ctx, queries.SoulInsert,
		soul.Scope, soul.UserID, soul.Name, soul.IdentityText, soul.VoiceText,
		soul.ValuesText, emptyStrings(soul.AllowedDomains), emptyStrings(soul.ForbiddenDomains), emptyMetadata(soul.Metadata),
	)
	return err
}

func (st *PGSoulStore) Update(ctx context.Context, tx Querier, soul *domain.Soul) error {
	_, err := tx.Exec(ctx, queries.SoulUpdate,
		soul.Name, soul.IdentityText, soul.VoiceText, soul.ValuesText,
		emptyStrings(soul.AllowedDomains), emptyStrings(soul.ForbiddenDomains), emptyMetadata(soul.Metadata), soul.ID,
	)
	return err
}

func (st *PGSoulStore) Delete(ctx context.Context, tx Querier, id uuid.UUID) error {
	_, err := tx.Exec(ctx, queries.SoulDelete, id)
	return err
}

func (st *PGSoulStore) GetByID(ctx context.Context, tx Querier, id uuid.UUID) (*domain.Soul, error) {
	return scanSoul(tx.QueryRow(ctx, queries.SoulSelectByID, id))
}

func (st *PGSoulStore) ListByUser(ctx context.Context, tx Querier, userID uuid.UUID) ([]*domain.Soul, error) {
	rows, err := tx.Query(ctx, queries.SoulsSelectByUser, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var souls []*domain.Soul
	for rows.Next() {
		s, err := scanSoul(rows)
		if err != nil {
			return nil, err
		}
		souls = append(souls, s)
	}
	return souls, rows.Err()
}

func (st *PGSoulStore) ListSystemSouls(ctx context.Context, tx Querier) ([]*domain.Soul, error) {
	rows, err := tx.Query(ctx, queries.SoulsSelectSystem)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var souls []*domain.Soul
	for rows.Next() {
		s, err := scanSoul(rows)
		if err != nil {
			return nil, err
		}
		souls = append(souls, s)
	}
	return souls, rows.Err()
}
