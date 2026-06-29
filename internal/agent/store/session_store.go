package store

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/queries"
)

// PGSessionStore 是 Session 的 PostgreSQL 实现。
type PGSessionStore struct{}

// NewPGSessionStore 创建一个新的 PGSessionStore。
func NewPGSessionStore() SessionStore {
	return &PGSessionStore{}
}

func scanSession(row pgx.Row) (*domain.Session, error) {
	var s domain.Session
	err := row.Scan(
		&s.ID, &s.UserID, &s.TenantID, &s.Title, &s.Status, &s.ParentSessionID,
		&s.ContextSummary, &s.ContextSummaryAt, &s.Metadata, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (st *PGSessionStore) CreateSession(ctx context.Context, tx Querier, session *domain.Session) error {
	metadata := session.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	return tx.QueryRow(ctx, queries.SessionInsert,
		session.UserID, session.TenantID, session.Title, session.Status,
		session.ParentSessionID, session.ContextSummary, session.ContextSummaryAt, metadata,
	).Scan(&session.ID)
}

func (st *PGSessionStore) UpdateSession(ctx context.Context, tx Querier, session *domain.Session) error {
	metadata := session.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	_, err := tx.Exec(ctx, queries.SessionUpdate,
		session.Title, session.Status, session.ContextSummary, session.ContextSummaryAt,
		metadata, session.ID,
	)
	return err
}

func (st *PGSessionStore) GetSessionByID(ctx context.Context, tx Querier, id uuid.UUID) (*domain.Session, error) {
	return scanSession(tx.QueryRow(ctx, queries.SessionSelectByID, id))
}

func (st *PGSessionStore) ListSessionsByUser(ctx context.Context, tx Querier, userID uuid.UUID, status domain.SessionStatus, limit int) ([]*domain.Session, error) {
	if limit <= 0 {
		limit = 100
	}
	rows, err := tx.Query(ctx, queries.SessionsSelectByUser, userID, status, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, rows.Err()
}

func scanMessage(row pgx.Row) (*domain.Message, error) {
	var m domain.Message
	var toolCallsRaw []byte
	var toolResultRaw []byte
	err := row.Scan(
		&m.ID, &m.SessionID, &m.Role, &m.Content, &toolCallsRaw, &m.ToolCallID,
		&m.ToolName, &toolResultRaw, &m.ReasoningContent, &m.TokenCount, &m.Metadata, &m.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	if len(toolCallsRaw) > 0 {
		_ = json.Unmarshal(toolCallsRaw, &m.ToolCalls)
	}
	if len(toolResultRaw) > 0 {
		_ = json.Unmarshal(toolResultRaw, &m.ToolResult)
	}
	return &m, nil
}

func (st *PGSessionStore) CreateMessage(ctx context.Context, tx Querier, message *domain.Message) error {
	toolCallsRaw, _ := json.Marshal(message.ToolCalls)
	toolResultRaw, _ := json.Marshal(message.ToolResult)
	metadata := message.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	return tx.QueryRow(ctx, queries.MessageInsert,
		message.SessionID, message.Role, message.Content, toolCallsRaw,
		message.ToolCallID, message.ToolName, toolResultRaw,
		message.ReasoningContent, message.TokenCount, metadata,
	).Scan(&message.ID)
}

func (st *PGSessionStore) ListMessagesBySession(ctx context.Context, tx Querier, sessionID uuid.UUID, limit int) ([]*domain.Message, error) {
	if limit <= 0 {
		limit = 1000
	}
	rows, err := tx.Query(ctx, queries.MessagesSelectBySession, sessionID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		m, err := scanMessage(rows)
		if err != nil {
			return nil, err
		}
		messages = append(messages, m)
	}
	return messages, rows.Err()
}

func (st *PGSessionStore) AttachProject(ctx context.Context, tx Querier, sessionID, projectID uuid.UUID) error {
	_, err := tx.Exec(ctx, queries.SessionProjectInsert, sessionID, projectID)
	return err
}

func (st *PGSessionStore) DetachProject(ctx context.Context, tx Querier, sessionID, projectID uuid.UUID) error {
	_, err := tx.Exec(ctx, queries.SessionProjectDelete, sessionID, projectID)
	return err
}

func (st *PGSessionStore) ListProjectIDsBySession(ctx context.Context, tx Querier, sessionID uuid.UUID) ([]uuid.UUID, error) {
	rows, err := tx.Query(ctx, queries.SessionProjectSelectBySession, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []uuid.UUID
	for rows.Next() {
		var id uuid.UUID
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}
