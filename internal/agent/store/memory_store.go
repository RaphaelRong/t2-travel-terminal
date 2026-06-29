package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/queries"
)

// PGMemoryStore 是 Memory 的 PostgreSQL 实现。
type PGMemoryStore struct{}

// NewPGMemoryStore 创建一个新的 PGMemoryStore。
func NewPGMemoryStore() MemoryStore {
	return &PGMemoryStore{}
}

func scanMemory(row pgx.Row) (*domain.Memory, error) {
	var m domain.Memory
	err := row.Scan(
		&m.ID, &m.UserID, &m.Category, &m.Content, &m.SourceSessionID, &m.SourceMessageID,
		&m.Confidence, &m.ExpiresAt, &m.Metadata, &m.CreatedAt, &m.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &m, nil
}

func (st *PGMemoryStore) Create(ctx context.Context, tx Querier, memory *domain.Memory) error {
	return tx.QueryRow(ctx, queries.MemoryInsert,
		memory.UserID, memory.Category, memory.Content, memory.SourceSessionID,
		memory.SourceMessageID, memory.Confidence, memory.ExpiresAt, emptyMetadata(memory.Metadata),
	).Scan(&memory.ID)
}

func (st *PGMemoryStore) Update(ctx context.Context, tx Querier, memory *domain.Memory) error {
	_, err := tx.Exec(ctx, queries.MemoryUpdate,
		memory.Category, memory.Content, memory.Confidence, memory.ExpiresAt, emptyMetadata(memory.Metadata), memory.ID,
	)
	return err
}

func (st *PGMemoryStore) Delete(ctx context.Context, tx Querier, id uuid.UUID) error {
	_, err := tx.Exec(ctx, queries.MemoryDelete, id)
	return err
}

func (st *PGMemoryStore) GetByID(ctx context.Context, tx Querier, id uuid.UUID) (*domain.Memory, error) {
	return scanMemory(tx.QueryRow(ctx, queries.MemorySelectByID, id))
}

func (st *PGMemoryStore) ListByUser(ctx context.Context, tx Querier, userID uuid.UUID, limit int) ([]*domain.Memory, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := tx.Query(ctx, queries.MemoriesSelectByUser, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*domain.Memory
	for rows.Next() {
		m, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	return memories, rows.Err()
}

func (st *PGMemoryStore) Search(ctx context.Context, tx Querier, userID uuid.UUID, query string, limit int) ([]*domain.Memory, error) {
	if limit <= 0 {
		limit = 20
	}
	rows, err := tx.Query(ctx, queries.MemoriesSearch, userID, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var memories []*domain.Memory
	for rows.Next() {
		m, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, m)
	}
	return memories, rows.Err()
}
