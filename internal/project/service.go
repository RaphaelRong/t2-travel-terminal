package project

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/jsonx"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/queries"
)

var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrIntegrationNotFound = errors.New("integration not found")
)

// Querier is the minimal pgx interface used by the project service.
type Querier interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// CapabilityRecord is the stored representation returned by Project APIs.
type CapabilityRecord struct {
	ID            uuid.UUID
	ProjectID     uuid.UUID
	IntegrationID *uuid.UUID
	Kind          string
	Name          string
	Description   *string
	Status        string
	RequestMethod *string
	RequestPath   *string
	InputSchema   json.RawMessage
	OutputSchema  json.RawMessage
	Metadata      json.RawMessage
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

// CapabilityInput contains mutable fields for a Project capability.
type CapabilityInput struct {
	Kind          string
	Name          string
	Description   string
	Status        string
	RequestMethod string
	RequestPath   string
	InputSchema   json.RawMessage
	OutputSchema  json.RawMessage
	Metadata      json.RawMessage
}

// IntegrationRecord is the stored representation returned by Project APIs.
type IntegrationRecord struct {
	ID               uuid.UUID
	ProjectID        uuid.UUID
	Kind             string
	Name             string
	Description      *string
	Status           string
	EndpointURL      *string
	DocumentationURL *string
	Transport        string
	AuthType         string
	RequestHeaders   json.RawMessage
	AuthConfig       json.RawMessage
	Metadata         json.RawMessage
	LastSyncedAt     *time.Time
	SyncStatus       string
	SyncError        *string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// IntegrationInput contains mutable fields for a Project integration.
type IntegrationInput struct {
	Kind             string
	Name             string
	Description      string
	Status           string
	EndpointURL      string
	DocumentationURL string
	Transport        string
	AuthType         string
	RequestHeaders   json.RawMessage
	AuthConfig       json.RawMessage
	Metadata         json.RawMessage
}

// ProjectRecord is the stored representation returned by Project APIs.
type ProjectRecord struct {
	ID                  uuid.UUID
	TenantID            *uuid.UUID
	SourceScope         string
	Kind                string
	Status              string
	SourceType          string
	Name                string
	Description         *string
	EndpointURL         *string
	RequestMethod       string
	RequestPath         *string
	RequestHeaders      json.RawMessage
	RequestBodyTemplate json.RawMessage
	AuthType            string
	AuthConfig          json.RawMessage
	CapabilitySummary   *string
	CreatedBy           *uuid.UUID
	CreatedAt           time.Time
	UpdatedAt           time.Time
	LastPublishedAt     *time.Time
}

// ProjectInput contains mutable fields for a Project.
type ProjectInput struct {
	TenantID            uuid.UUID
	SourceScope         string
	Status              string
	SourceType          string
	Name                string
	Description         string
	EndpointURL         string
	RequestMethod       string
	RequestPath         string
	RequestHeaders      json.RawMessage
	RequestBodyTemplate json.RawMessage
	AuthType            string
	AuthConfig          json.RawMessage
	CapabilitySummary   string
	CreatedBy           uuid.UUID
}

// SyncFailure wraps a manifest/provider sync failure after the failure status is persisted.
type SyncFailure struct {
	Err error
}

func (e *SyncFailure) Error() string {
	return e.Err.Error()
}

func (e *SyncFailure) Unwrap() error {
	return e.Err
}

// SyncService coordinates Project Integration synchronization.
type SyncService struct {
	syncer *Syncer
}

// NewSyncService creates a Project synchronization service.
func NewSyncService() *SyncService {
	return &SyncService{syncer: NewSyncer()}
}

// ListAccessibleProjects returns tenant-visible projects plus online system projects.
func (s *SyncService) ListAccessibleProjects(ctx context.Context, q Querier, tenantID uuid.UUID) ([]ProjectRecord, error) {
	rows, err := q.Query(ctx, queries.ProjectsListAccessible, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProjectRows(rows)
}

// ListSystemProjects returns system-scope projects.
func (s *SyncService) ListSystemProjects(ctx context.Context, q Querier) ([]ProjectRecord, error) {
	rows, err := q.Query(ctx, queries.ProjectsListSystem)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanProjectRows(rows)
}

// GetProject returns one Project by ID.
func (s *SyncService) GetProject(ctx context.Context, q Querier, projectID uuid.UUID) (ProjectRecord, error) {
	project, err := scanProjectRecord(q.QueryRow(ctx, queries.ProjectsSelectByID, projectID))
	if errors.Is(err, pgx.ErrNoRows) {
		return ProjectRecord{}, ErrProjectNotFound
	}
	return project, err
}

// CreateProject creates a Project.
func (s *SyncService) CreateProject(ctx context.Context, q Querier, input ProjectInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := q.QueryRow(ctx,
		queries.ProjectsInsert,
		input.TenantID, input.SourceScope, input.Status, input.SourceType,
		input.Name, input.Description, input.EndpointURL, input.RequestMethod, input.RequestPath,
		jsonx.Object(input.RequestHeaders), jsonx.Object(input.RequestBodyTemplate), input.AuthType, jsonx.Object(input.AuthConfig),
		input.CapabilitySummary, input.CreatedBy,
	).Scan(&id)
	return id, err
}

// UpdateProject updates a Project.
func (s *SyncService) UpdateProject(ctx context.Context, q Querier, projectID uuid.UUID, input ProjectInput) error {
	_, err := q.Exec(ctx,
		queries.ProjectsUpdate,
		input.Status, input.SourceType, input.Name, input.Description, input.EndpointURL,
		input.RequestMethod, input.RequestPath, jsonx.Object(input.RequestHeaders), jsonx.Object(input.RequestBodyTemplate),
		input.AuthType, jsonx.Object(input.AuthConfig), input.CapabilitySummary, projectID,
	)
	return err
}

// DeleteProject deletes a Project.
func (s *SyncService) DeleteProject(ctx context.Context, q Querier, projectID uuid.UUID) error {
	_, err := q.Exec(ctx, queries.ProjectsDelete, projectID)
	return err
}

// ListCapabilities returns capabilities for one Project.
func (s *SyncService) ListCapabilities(ctx context.Context, q Querier, projectID uuid.UUID) ([]CapabilityRecord, error) {
	rows, err := q.Query(ctx, queries.ProjectCapabilitiesList, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []CapabilityRecord{}
	for rows.Next() {
		capability, err := scanCapability(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, capability)
	}
	return result, rows.Err()
}

// CreateCapability creates one manually managed Project capability.
func (s *SyncService) CreateCapability(ctx context.Context, q Querier, projectID uuid.UUID, input CapabilityInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := q.QueryRow(ctx,
		queries.ProjectCapabilitiesInsert,
		projectID, input.Kind, input.Name, input.Description, input.Status,
		input.RequestMethod, input.RequestPath,
		jsonx.Object(input.InputSchema), jsonx.Object(input.OutputSchema), jsonx.Object(input.Metadata),
	).Scan(&id)
	return id, err
}

// UpdateCapability updates one manually managed Project capability.
func (s *SyncService) UpdateCapability(ctx context.Context, q Querier, projectID uuid.UUID, capabilityID uuid.UUID, input CapabilityInput) error {
	_, err := q.Exec(ctx,
		queries.ProjectCapabilitiesUpdate,
		input.Kind, input.Name, input.Description, input.Status, input.RequestMethod, input.RequestPath,
		jsonx.Object(input.InputSchema), jsonx.Object(input.OutputSchema), jsonx.Object(input.Metadata),
		projectID, capabilityID,
	)
	return err
}

// DeleteCapability deletes one Project capability.
func (s *SyncService) DeleteCapability(ctx context.Context, q Querier, projectID uuid.UUID, capabilityID uuid.UUID) error {
	_, err := q.Exec(ctx, queries.ProjectCapabilitiesDelete, projectID, capabilityID)
	return err
}

// ListIntegrations returns integrations for one Project.
func (s *SyncService) ListIntegrations(ctx context.Context, q Querier, projectID uuid.UUID) ([]IntegrationRecord, error) {
	rows, err := q.Query(ctx, queries.ProjectIntegrationsList, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []IntegrationRecord{}
	for rows.Next() {
		integration, err := scanIntegrationRecord(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, integration)
	}
	return result, rows.Err()
}

// CreateIntegration creates one Project integration.
func (s *SyncService) CreateIntegration(ctx context.Context, q Querier, projectID uuid.UUID, input IntegrationInput) (uuid.UUID, error) {
	var id uuid.UUID
	err := q.QueryRow(ctx,
		queries.ProjectIntegrationsInsert,
		projectID, input.Kind, input.Name, input.Description, input.Status,
		input.EndpointURL, input.DocumentationURL, input.Transport, input.AuthType,
		jsonx.Object(input.RequestHeaders), jsonx.Object(input.AuthConfig), jsonx.Object(input.Metadata),
	).Scan(&id)
	return id, err
}

// UpdateIntegration updates one Project integration.
func (s *SyncService) UpdateIntegration(ctx context.Context, q Querier, projectID uuid.UUID, integrationID uuid.UUID, input IntegrationInput) error {
	_, err := q.Exec(ctx,
		queries.ProjectIntegrationsUpdate,
		input.Kind, input.Name, input.Description, input.Status, input.EndpointURL,
		input.DocumentationURL, input.Transport, input.AuthType,
		jsonx.Object(input.RequestHeaders), jsonx.Object(input.AuthConfig), jsonx.Object(input.Metadata),
		projectID, integrationID,
	)
	return err
}

// DeleteIntegration deletes one Project integration.
func (s *SyncService) DeleteIntegration(ctx context.Context, q Querier, projectID uuid.UUID, integrationID uuid.UUID) error {
	_, err := q.Exec(ctx, queries.ProjectIntegrationsDelete, projectID, integrationID)
	return err
}

// SyncIntegration refreshes capabilities for one Project Integration.
func (s *SyncService) SyncIntegration(ctx context.Context, q Querier, projectID uuid.UUID, integrationID uuid.UUID) ([]uuid.UUID, error) {
	integration, err := scanIntegration(q.QueryRow(ctx, queries.ProjectIntegrationsSelectByID, projectID, integrationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrIntegrationNotFound
	}
	if err != nil {
		return nil, err
	}

	project, err := scanSyncProject(q.QueryRow(ctx, queries.ProjectsSelectByID, projectID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProjectNotFound
	}
	if err != nil {
		return nil, err
	}

	capabilities, syncErr := s.syncer.SyncCapabilities(ctx, project, integration)
	if syncErr != nil {
		_, _ = q.Exec(ctx, queries.ProjectIntegrationsUpdateSyncResult, "failed", syncErr.Error(), projectID, integrationID)
		return nil, &SyncFailure{Err: syncErr}
	}

	if _, err := q.Exec(ctx, queries.ProjectCapabilitiesDeleteByIntegration, projectID, integrationID); err != nil {
		return nil, err
	}

	created := []uuid.UUID{}
	for _, capability := range capabilities {
		name := strings.TrimSpace(capability.Name)
		if name == "" {
			continue
		}

		var id uuid.UUID
		err := q.QueryRow(ctx,
			queries.ProjectCapabilitiesInsertForIntegration,
			projectID, integrationID, capability.Kind, name, capability.ExternalName, capability.Description, "active",
			capability.RequestMethod, capability.RequestPath,
			jsonx.Object(capability.InputSchema), jsonx.Object(capability.OutputSchema), jsonx.Object(capability.Metadata),
		).Scan(&id)
		if err != nil {
			return nil, err
		}
		created = append(created, id)
	}

	if _, err := q.Exec(ctx, queries.ProjectIntegrationsUpdateSyncResult, "success", "", projectID, integrationID); err != nil {
		return nil, err
	}
	return created, nil
}

func scanSyncProject(row pgx.Row) (Project, error) {
	var project Project
	var discardID uuid.UUID
	var discardTenantID *uuid.UUID
	var discardCreatedBy *uuid.UUID
	var discardString string
	var discardStringPtr *string
	var discardTime time.Time
	var discardLastPublishedAt *time.Time
	var requestBodyTemplate json.RawMessage

	err := row.Scan(
		&discardID, &discardTenantID, &discardString, &discardString, &discardString, &discardString,
		&discardString, &discardStringPtr, &discardStringPtr, &discardString, &discardStringPtr,
		&project.RequestHeaders, &requestBodyTemplate, &project.AuthType, &project.AuthConfig,
		&discardStringPtr, &discardCreatedBy, &discardTime, &discardTime, &discardLastPublishedAt,
	)
	project.RequestHeaders = jsonx.Object(project.RequestHeaders)
	project.AuthConfig = jsonx.Object(project.AuthConfig)
	return project, err
}

func scanProjectRows(rows pgx.Rows) ([]ProjectRecord, error) {
	result := []ProjectRecord{}
	for rows.Next() {
		project, err := scanProjectRecord(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, project)
	}
	return result, rows.Err()
}

func scanProjectRecord(row pgx.Row) (ProjectRecord, error) {
	var project ProjectRecord
	err := row.Scan(
		&project.ID, &project.TenantID, &project.SourceScope, &project.Kind, &project.Status, &project.SourceType,
		&project.Name, &project.Description, &project.EndpointURL, &project.RequestMethod, &project.RequestPath,
		&project.RequestHeaders, &project.RequestBodyTemplate, &project.AuthType, &project.AuthConfig,
		&project.CapabilitySummary, &project.CreatedBy, &project.CreatedAt, &project.UpdatedAt, &project.LastPublishedAt,
	)
	project.RequestHeaders = jsonx.Object(project.RequestHeaders)
	project.RequestBodyTemplate = jsonx.Object(project.RequestBodyTemplate)
	project.AuthConfig = jsonx.Object(project.AuthConfig)
	return project, err
}

func scanIntegration(row pgx.Row) (Integration, error) {
	var item Integration
	var discardID uuid.UUID
	var discardProjectID uuid.UUID
	var discardString string
	var discardStringPtr *string
	var endpointURL *string
	var documentationURL *string
	var discardTime time.Time
	var discardLastSyncedAt *time.Time

	err := row.Scan(
		&discardID, &discardProjectID, &item.Kind, &discardString, &discardStringPtr, &discardString,
		&endpointURL, &documentationURL, &discardString, &item.AuthType, &item.RequestHeaders,
		&item.AuthConfig, &discardString, &discardLastSyncedAt, &discardString, &discardStringPtr,
		&discardTime, &discardTime,
	)
	item.EndpointURL = stringValue(endpointURL)
	item.DocumentationURL = stringValue(documentationURL)
	item.RequestHeaders = jsonx.Object(item.RequestHeaders)
	item.AuthConfig = jsonx.Object(item.AuthConfig)
	return item, err
}

func scanCapability(row pgx.Row) (CapabilityRecord, error) {
	var capability CapabilityRecord
	err := row.Scan(
		&capability.ID, &capability.ProjectID, &capability.Kind, &capability.Name, &capability.Description, &capability.Status,
		&capability.IntegrationID, &capability.RequestMethod, &capability.RequestPath, &capability.InputSchema, &capability.OutputSchema, &capability.Metadata,
		&capability.CreatedAt, &capability.UpdatedAt,
	)
	capability.InputSchema = jsonx.Object(capability.InputSchema)
	capability.OutputSchema = jsonx.Object(capability.OutputSchema)
	capability.Metadata = jsonx.Object(capability.Metadata)
	return capability, err
}

func scanIntegrationRecord(row pgx.Row) (IntegrationRecord, error) {
	var integration IntegrationRecord
	err := row.Scan(
		&integration.ID, &integration.ProjectID, &integration.Kind, &integration.Name, &integration.Description,
		&integration.Status, &integration.EndpointURL, &integration.DocumentationURL, &integration.Transport,
		&integration.AuthType, &integration.RequestHeaders, &integration.AuthConfig, &integration.Metadata,
		&integration.LastSyncedAt, &integration.SyncStatus, &integration.SyncError,
		&integration.CreatedAt, &integration.UpdatedAt,
	)
	integration.RequestHeaders = jsonx.Object(integration.RequestHeaders)
	integration.AuthConfig = jsonx.Object(integration.AuthConfig)
	integration.Metadata = jsonx.Object(integration.Metadata)
	return integration, err
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}
