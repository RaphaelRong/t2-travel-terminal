package store

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/queries"
)

// PGUserDataStore 是用户业务数据的 PostgreSQL 实现。
type PGUserDataStore struct{}

// NewPGUserDataStore 创建一个新的 PGUserDataStore。
func NewPGUserDataStore() UserDataStore {
	return &PGUserDataStore{}
}

func scanUserDataset(row pgx.Row) (*domain.UserDataset, error) {
	var d domain.UserDataset
	err := row.Scan(
		&d.ID, &d.UserID, &d.Name, &d.Description, &d.Schema,
		&d.RowCount, &d.Source, &d.Metadata, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &d, nil
}

func (st *PGUserDataStore) CreateDataset(ctx context.Context, tx Querier, dataset *domain.UserDataset) error {
	return tx.QueryRow(ctx, queries.UserDatasetInsert,
		dataset.UserID, dataset.Name, dataset.Description, emptyMetadata(dataset.Schema),
		dataset.RowCount, dataset.Source, emptyMetadata(dataset.Metadata),
	).Scan(&dataset.ID)
}

func (st *PGUserDataStore) UpdateDataset(ctx context.Context, tx Querier, dataset *domain.UserDataset) error {
	_, err := tx.Exec(ctx, queries.UserDatasetUpdate,
		dataset.Name, dataset.Description, emptyMetadata(dataset.Schema), dataset.RowCount,
		dataset.Source, emptyMetadata(dataset.Metadata), dataset.ID,
	)
	return err
}

func (st *PGUserDataStore) DeleteDataset(ctx context.Context, tx Querier, id uuid.UUID) error {
	_, err := tx.Exec(ctx, queries.UserDatasetDelete, id)
	return err
}

func (st *PGUserDataStore) GetDatasetByID(ctx context.Context, tx Querier, id uuid.UUID) (*domain.UserDataset, error) {
	return scanUserDataset(tx.QueryRow(ctx, queries.UserDatasetSelectByID, id))
}

func (st *PGUserDataStore) GetDatasetByName(ctx context.Context, tx Querier, userID uuid.UUID, name string) (*domain.UserDataset, error) {
	return scanUserDataset(tx.QueryRow(ctx, queries.UserDatasetSelectByName, userID, name))
}

func (st *PGUserDataStore) ListDatasetsByUser(ctx context.Context, tx Querier, userID uuid.UUID, limit int) ([]*domain.UserDataset, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := tx.Query(ctx, queries.UserDatasetsSelectByUser, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var datasets []*domain.UserDataset
	for rows.Next() {
		d, err := scanUserDataset(rows)
		if err != nil {
			return nil, err
		}
		datasets = append(datasets, d)
	}
	return datasets, rows.Err()
}

func (st *PGUserDataStore) InsertRows(ctx context.Context, tx Querier, datasetID uuid.UUID, rows []map[string]any) error {
	var userID uuid.UUID
	if err := tx.QueryRow(ctx, "SELECT user_id FROM agent_user_datasets WHERE id = $1", datasetID).Scan(&userID); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, queries.UserDatasetRowsDeleteByDataset, datasetID); err != nil {
		return err
	}

	for i, row := range rows {
		_, err := tx.Exec(ctx, queries.UserDatasetRowsInsert, datasetID, userID, i, row)
		if err != nil {
			return err
		}
	}

	_, err := tx.Exec(ctx,
		"UPDATE agent_user_datasets SET row_count = $1, updated_at = now() WHERE id = $2",
		len(rows), datasetID,
	)
	return err
}

func (st *PGUserDataStore) Query(ctx context.Context, tx Querier, userID uuid.UUID, sql string, args ...any) ([]map[string]any, error) {
	// 通过设置会话变量让 RLS 策略识别当前用户
	if _, err := tx.Exec(ctx, "SELECT set_config('app.current_user_id', $1, true)", userID.String()); err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colDescriptions := rows.FieldDescriptions()
	var result []map[string]any
	for rows.Next() {
		values, err := rows.Values()
		if err != nil {
			return nil, err
		}
		row := make(map[string]any, len(values))
		for i, col := range colDescriptions {
			row[col.Name] = values[i]
		}
		result = append(result, row)
	}
	return result, rows.Err()
}
