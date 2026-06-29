package store

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/queries"
)

// PGProjectStore 是 Agent 视角的 Project 数据访问层。
type PGProjectStore struct{}

// NewPGProjectStore 创建 PGProjectStore。
func NewPGProjectStore() ProjectStore {
	return &PGProjectStore{}
}

func scanProject(row pgx.Row) (*domain.Project, error) {
	var p domain.Project
	var description, endpointURL, requestPath, capabilitySummary *string
	var requestHeadersRaw []byte
	var requestBodyTemplateRaw []byte
	var authConfigRaw []byte

	err := row.Scan(
		&p.ID, &p.TenantID, &p.SourceScope, &p.Kind, &p.Status, &p.SourceType,
		&p.Name, &description, &endpointURL, &p.RequestMethod, &requestPath,
		&requestHeadersRaw, &requestBodyTemplateRaw, &p.AuthType, &authConfigRaw,
		&capabilitySummary, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &p.LastPublishedAt,
	)
	if err != nil {
		return nil, err
	}
	p.Description = stringValue(description)
	p.EndpointURL = stringValue(endpointURL)
	p.RequestPath = stringValue(requestPath)
	p.CapabilitySummary = stringValue(capabilitySummary)
	p.RequestHeaders = parseStringMap(requestHeadersRaw)
	p.RequestBodyTemplate = parseInterfaceMap(requestBodyTemplateRaw)
	p.AuthConfig = parseStringMap(authConfigRaw)
	return &p, nil
}

func scanProjectIntegration(row pgx.Row) (*domain.ProjectIntegration, error) {
	var i domain.ProjectIntegration
	var description, endpointURL, documentationURL, transport, syncStatus, syncError *string
	var requestHeadersRaw []byte
	var authConfigRaw []byte
	var metadataRaw []byte

	err := row.Scan(
		&i.ID, &i.ProjectID, &i.Kind, &i.Name, &description, &i.Status,
		&endpointURL, &documentationURL, &transport, &i.AuthType, &requestHeadersRaw,
		&authConfigRaw, &metadataRaw, &i.LastSyncedAt, &syncStatus, &syncError,
		&i.CreatedAt, &i.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	i.Description = stringValue(description)
	i.EndpointURL = stringValue(endpointURL)
	i.DocumentationURL = stringValue(documentationURL)
	i.Transport = stringValue(transport)
	i.SyncStatus = stringValue(syncStatus)
	i.SyncError = stringValue(syncError)
	i.RequestHeaders = parseStringMap(requestHeadersRaw)
	i.AuthConfig = parseStringMap(authConfigRaw)
	i.Metadata = parseInterfaceMap(metadataRaw)
	return &i, nil
}

func scanProjectCapability(row pgx.Row) (*domain.ProjectCapability, *domain.Project, *domain.ProjectIntegration, error) {
	var c domain.ProjectCapability
	var projectEndpointURL, projectAuthType *string
	var projectHeadersRaw, projectAuthRaw []byte
	var integrationKind, integrationEndpointURL, integrationAuthType *string
	var integrationHeadersRaw, integrationAuthRaw []byte
	var inputSchemaRaw, outputSchemaRaw, metadataRaw []byte
	var externalName, description, requestMethod, requestPath *string

	err := row.Scan(
		&c.ID, &c.ProjectID, &c.IntegrationID, &c.Kind, &c.Name, &externalName, &description,
		&c.Status, &requestMethod, &requestPath, &inputSchemaRaw, &outputSchemaRaw, &metadataRaw,
		&c.CreatedAt, &c.UpdatedAt,
		&integrationKind,
		&projectEndpointURL, &projectHeadersRaw, &projectAuthType, &projectAuthRaw,
		&integrationEndpointURL, &integrationHeadersRaw, &integrationAuthType, &integrationAuthRaw,
	)
	if err != nil {
		return nil, nil, nil, err
	}

	c.ExternalName = stringValue(externalName)
	c.Description = stringValue(description)
	c.RequestMethod = stringValue(requestMethod)
	c.RequestPath = stringValue(requestPath)
	c.InputSchema = parseInterfaceMap(inputSchemaRaw)
	c.OutputSchema = parseInterfaceMap(outputSchemaRaw)
	c.Metadata = parseInterfaceMap(metadataRaw)

	project := &domain.Project{
		ID:             c.ProjectID,
		EndpointURL:    stringValue(projectEndpointURL),
		RequestHeaders: parseStringMap(projectHeadersRaw),
		AuthType:       stringValue(projectAuthType),
		AuthConfig:     parseStringMap(projectAuthRaw),
	}

	var integration *domain.ProjectIntegration
	if c.IntegrationID != nil {
		integration = &domain.ProjectIntegration{
			ID:             *c.IntegrationID,
			ProjectID:      c.ProjectID,
			Kind:           stringValue(integrationKind),
			EndpointURL:    stringValue(integrationEndpointURL),
			RequestHeaders: parseStringMap(integrationHeadersRaw),
			AuthType:       stringValue(integrationAuthType),
			AuthConfig:     parseStringMap(integrationAuthRaw),
		}
	}

	return &c, project, integration, nil
}

func (st *PGProjectStore) ListBySession(ctx context.Context, q Querier, sessionID uuid.UUID) ([]*domain.Project, error) {
	rows, err := q.Query(ctx, queries.ProjectsSelectBySession, sessionID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*domain.Project
	for rows.Next() {
		p, err := scanProject(rows)
		if err != nil {
			return nil, err
		}
		projects = append(projects, p)
	}
	return projects, rows.Err()
}

func (st *PGProjectStore) ListCapabilitiesBySession(ctx context.Context, q Querier, sessionID uuid.UUID) ([]*domain.ProjectCapability, []*domain.Project, []*domain.ProjectIntegration, error) {
	rows, err := q.Query(ctx, queries.ProjectCapabilitiesSelectBySession, sessionID)
	if err != nil {
		return nil, nil, nil, err
	}
	defer rows.Close()

	var caps []*domain.ProjectCapability
	var projects []*domain.Project
	var integrations []*domain.ProjectIntegration

	for rows.Next() {
		c, p, i, err := scanProjectCapability(rows)
		if err != nil {
			return nil, nil, nil, err
		}
		caps = append(caps, c)
		projects = append(projects, p)
		integrations = append(integrations, i)
	}
	return caps, projects, integrations, rows.Err()
}

func (st *PGProjectStore) GetCapabilityByID(ctx context.Context, q Querier, capabilityID uuid.UUID) (*domain.ProjectCapability, *domain.Project, *domain.ProjectIntegration, error) {
	c, p, i, err := scanProjectCapability(q.QueryRow(ctx, queries.ProjectCapabilitySelectByID, capabilityID))
	if err != nil {
		return nil, nil, nil, err
	}
	return c, p, i, nil
}

func (st *PGProjectStore) GetProjectByID(ctx context.Context, q Querier, projectID uuid.UUID) (*domain.Project, error) {
	return scanProject(q.QueryRow(ctx, queries.ProjectSelectByID, projectID))
}

func (st *PGProjectStore) GetIntegrationByID(ctx context.Context, q Querier, integrationID uuid.UUID) (*domain.ProjectIntegration, error) {
	return scanProjectIntegration(q.QueryRow(ctx, queries.ProjectIntegrationSelectByID, integrationID))
}

func parseStringMap(raw []byte) map[string]string {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]string
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

func parseInterfaceMap(raw []byte) map[string]interface{} {
	if len(raw) == 0 {
		return nil
	}
	var m map[string]interface{}
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil
	}
	return m
}

func stringValue(v *string) string {
	if v == nil {
		return ""
	}
	return *v
}
