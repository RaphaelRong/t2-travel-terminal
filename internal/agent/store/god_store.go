package store

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/queries"
)

// PGGodConfigStore 是 GodConfig 的 PostgreSQL 实现。
type PGGodConfigStore struct{}

// NewPGGodConfigStore 创建一个新的 PGGodConfigStore。
func NewPGGodConfigStore() GodConfigStore {
	return &PGGodConfigStore{}
}

func scanGodConfig(row pgx.Row) (*domain.GodConfig, error) {
	var c domain.GodConfig
	err := row.Scan(
		&c.ID, &c.Name, &c.IsActive, &c.AllowedDomains, &c.ForbiddenDomains,
		&c.AllowedTools, &c.ForbiddenTools, &c.RequireApprovalTools, &c.MaxIterations,
		&c.CanDelegate, &c.CanRunWorkflow, &c.Rules, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (st *PGGodConfigStore) Create(ctx context.Context, tx Querier, config *domain.GodConfig) error {
	return tx.QueryRow(ctx, queries.GodConfigInsert,
		config.Name, config.IsActive, emptyStrings(config.AllowedDomains), emptyStrings(config.ForbiddenDomains),
		emptyStrings(config.AllowedTools), emptyStrings(config.ForbiddenTools), emptyStrings(config.RequireApprovalTools),
		config.MaxIterations, config.CanDelegate, config.CanRunWorkflow, config.Rules,
	).Scan(&config.ID)
}

func (st *PGGodConfigStore) Update(ctx context.Context, tx Querier, config *domain.GodConfig) error {
	_, err := tx.Exec(ctx, queries.GodConfigUpdate,
		config.IsActive, emptyStrings(config.AllowedDomains), emptyStrings(config.ForbiddenDomains),
		emptyStrings(config.AllowedTools), emptyStrings(config.ForbiddenTools), emptyStrings(config.RequireApprovalTools),
		config.MaxIterations, config.CanDelegate, config.CanRunWorkflow, config.Rules,
		config.ID,
	)
	return err
}

func (st *PGGodConfigStore) GetByName(ctx context.Context, tx Querier, name string) (*domain.GodConfig, error) {
	return scanGodConfig(tx.QueryRow(ctx, queries.GodConfigSelectByName, name))
}

func (st *PGGodConfigStore) GetActive(ctx context.Context, tx Querier) (*domain.GodConfig, error) {
	return scanGodConfig(tx.QueryRow(ctx, queries.GodConfigSelectActive))
}

func (st *PGGodConfigStore) List(ctx context.Context, tx Querier) ([]*domain.GodConfig, error) {
	rows, err := tx.Query(ctx, queries.GodConfigList)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []*domain.GodConfig
	for rows.Next() {
		c, err := scanGodConfig(rows)
		if err != nil {
			return nil, err
		}
		configs = append(configs, c)
	}
	return configs, rows.Err()
}
