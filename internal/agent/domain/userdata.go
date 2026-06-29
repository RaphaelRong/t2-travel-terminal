package domain

import (
	"time"

	"github.com/google/uuid"
)

// UserDataset 表示用户上传或生成的业务数据集。
type UserDataset struct {
	ID          uuid.UUID
	UserID      uuid.UUID
	Name        string
	Description string
	Schema      map[string]any
	RowCount    int
	Source      string
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// UserDatasetRow 表示数据集中的一行数据。
type UserDatasetRow struct {
	ID        uuid.UUID
	DatasetID uuid.UUID
	UserID    uuid.UUID
	RowIndex  int
	Data      map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}
