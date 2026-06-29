package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/integration"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/jsonx"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/queries"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
	"gopkg.in/yaml.v3"
)

type projectCapabilityReq struct {
	Kind          string          `json:"kind" binding:"required"`
	Name          string          `json:"name" binding:"required"`
	Description   string          `json:"description"`
	Status        string          `json:"status"`
	RequestMethod string          `json:"request_method"`
	RequestPath   string          `json:"request_path"`
	InputSchema   json.RawMessage `json:"input_schema"`
	OutputSchema  json.RawMessage `json:"output_schema"`
	Metadata      json.RawMessage `json:"metadata"`
}

type projectCapabilityResp struct {
	ID            uuid.UUID       `json:"id"`
	ProjectID     uuid.UUID       `json:"project_id"`
	IntegrationID *uuid.UUID      `json:"integration_id,omitempty"`
	Kind          string          `json:"kind"`
	Name          string          `json:"name"`
	Description   *string         `json:"description,omitempty"`
	Status        string          `json:"status"`
	RequestMethod *string         `json:"request_method,omitempty"`
	RequestPath   *string         `json:"request_path,omitempty"`
	InputSchema   json.RawMessage `json:"input_schema"`
	OutputSchema  json.RawMessage `json:"output_schema"`
	Metadata      json.RawMessage `json:"metadata"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
}

type projectIntegrationReq struct {
	Kind             string          `json:"kind" binding:"required"`
	Name             string          `json:"name" binding:"required"`
	Description      string          `json:"description"`
	Status           string          `json:"status"`
	EndpointURL      string          `json:"endpoint_url"`
	DocumentationURL string          `json:"documentation_url"`
	Transport        string          `json:"transport"`
	AuthType         string          `json:"auth_type"`
	RequestHeaders   json.RawMessage `json:"request_headers"`
	AuthConfig       json.RawMessage `json:"auth_config"`
	Metadata         json.RawMessage `json:"metadata"`
}

type projectIntegrationResp struct {
	ID               uuid.UUID       `json:"id"`
	ProjectID        uuid.UUID       `json:"project_id"`
	Kind             string          `json:"kind"`
	Name             string          `json:"name"`
	Description      *string         `json:"description,omitempty"`
	Status           string          `json:"status"`
	EndpointURL      *string         `json:"endpoint_url,omitempty"`
	DocumentationURL *string         `json:"documentation_url,omitempty"`
	Transport        string          `json:"transport"`
	AuthType         string          `json:"auth_type"`
	RequestHeaders   json.RawMessage `json:"request_headers"`
	AuthConfig       json.RawMessage `json:"auth_config"`
	Metadata         json.RawMessage `json:"metadata"`
	LastSyncedAt     *time.Time      `json:"last_synced_at,omitempty"`
	SyncStatus       string          `json:"sync_status"`
	SyncError        *string         `json:"sync_error,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
}

type projectResp struct {
	ID                  uuid.UUID                `json:"id"`
	TenantID            *uuid.UUID               `json:"tenant_id,omitempty"`
	SourceScope         string                   `json:"source_scope"`
	Kind                string                   `json:"kind"`
	Status              string                   `json:"status"`
	SourceType          string                   `json:"source_type"`
	Name                string                   `json:"name"`
	Description         *string                  `json:"description,omitempty"`
	EndpointURL         *string                  `json:"endpoint_url,omitempty"`
	RequestMethod       string                   `json:"request_method"`
	RequestPath         *string                  `json:"request_path,omitempty"`
	RequestHeaders      json.RawMessage          `json:"request_headers"`
	RequestBodyTemplate json.RawMessage          `json:"request_body_template"`
	AuthType            string                   `json:"auth_type"`
	AuthConfig          json.RawMessage          `json:"auth_config"`
	CapabilitySummary   *string                  `json:"capability_summary,omitempty"`
	CreatedBy           *uuid.UUID               `json:"created_by,omitempty"`
	CreatedAt           time.Time                `json:"created_at"`
	UpdatedAt           time.Time                `json:"updated_at"`
	LastPublishedAt     *time.Time               `json:"last_published_at,omitempty"`
	Integrations        []projectIntegrationResp `json:"integrations"`
	Capabilities        []projectCapabilityResp  `json:"capabilities"`
}

type projectUpsertReq struct {
	Name                string          `json:"name" binding:"required"`
	Description         string          `json:"description"`
	Status              string          `json:"status"`
	SourceType          string          `json:"source_type"`
	EndpointURL         string          `json:"endpoint_url"`
	RequestMethod       string          `json:"request_method"`
	RequestPath         string          `json:"request_path"`
	RequestHeaders      json.RawMessage `json:"request_headers"`
	RequestBodyTemplate json.RawMessage `json:"request_body_template"`
	AuthType            string          `json:"auth_type"`
	AuthConfig          json.RawMessage `json:"auth_config"`
	CapabilitySummary   string          `json:"capability_summary"`
}

func normalizedJSON(raw json.RawMessage) json.RawMessage {
	return jsonx.Object(raw)
}

func normalizeProjectDefaults(req *projectUpsertReq) {
	if req.Status == "" {
		req.Status = "draft"
	}
	if req.SourceType == "" {
		req.SourceType = "api"
	}
	if req.RequestMethod == "" {
		req.RequestMethod = "GET"
	}
	if req.AuthType == "" {
		req.AuthType = "none"
	}
	req.RequestHeaders = normalizedJSON(req.RequestHeaders)
	req.RequestBodyTemplate = normalizedJSON(req.RequestBodyTemplate)
	req.AuthConfig = normalizedJSON(req.AuthConfig)
}

func scanProject(row pgx.Row) (projectResp, error) {
	var p projectResp
	if err := row.Scan(
		&p.ID, &p.TenantID, &p.SourceScope, &p.Kind, &p.Status, &p.SourceType,
		&p.Name, &p.Description, &p.EndpointURL, &p.RequestMethod, &p.RequestPath,
		&p.RequestHeaders, &p.RequestBodyTemplate, &p.AuthType, &p.AuthConfig,
		&p.CapabilitySummary, &p.CreatedBy, &p.CreatedAt, &p.UpdatedAt, &p.LastPublishedAt,
	); err != nil {
		return p, err
	}
	if len(p.RequestHeaders) == 0 {
		p.RequestHeaders = json.RawMessage(`{}`)
	}
	if len(p.RequestBodyTemplate) == 0 {
		p.RequestBodyTemplate = json.RawMessage(`{}`)
	}
	if len(p.AuthConfig) == 0 {
		p.AuthConfig = json.RawMessage(`{}`)
	}
	return p, nil
}

func listProjectCapabilities(c *gin.Context, tx pgx.Tx, projectID uuid.UUID) ([]projectCapabilityResp, error) {
	rows, err := tx.Query(c.Request.Context(), queries.ProjectCapabilitiesList, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []projectCapabilityResp{}
	for rows.Next() {
		var cap projectCapabilityResp
		if err := rows.Scan(
			&cap.ID, &cap.ProjectID, &cap.Kind, &cap.Name, &cap.Description, &cap.Status,
			&cap.IntegrationID, &cap.RequestMethod, &cap.RequestPath, &cap.InputSchema, &cap.OutputSchema, &cap.Metadata,
			&cap.CreatedAt, &cap.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if len(cap.InputSchema) == 0 {
			cap.InputSchema = json.RawMessage(`{}`)
		}
		if len(cap.OutputSchema) == 0 {
			cap.OutputSchema = json.RawMessage(`{}`)
		}
		if len(cap.Metadata) == 0 {
			cap.Metadata = json.RawMessage(`{}`)
		}
		result = append(result, cap)
	}
	return result, rows.Err()
}

func listProjectIntegrations(c *gin.Context, tx pgx.Tx, projectID uuid.UUID) ([]projectIntegrationResp, error) {
	rows, err := tx.Query(c.Request.Context(), queries.ProjectIntegrationsList, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := []projectIntegrationResp{}
	for rows.Next() {
		integration, err := scanProjectIntegration(rows)
		if err != nil {
			return nil, err
		}
		result = append(result, integration)
	}
	return result, rows.Err()
}

func hydrateProjectDetails(c *gin.Context, tx pgx.Tx, projects []projectResp) error {
	for i := range projects {
		integrations, err := listProjectIntegrations(c, tx, projects[i].ID)
		if err != nil {
			return err
		}
		capabilities, err := listProjectCapabilities(c, tx, projects[i].ID)
		if err != nil {
			return err
		}
		projects[i].Integrations = integrations
		projects[i].Capabilities = capabilities
	}
	return nil
}

func scanProjectIntegration(row pgx.Row) (projectIntegrationResp, error) {
	var integration projectIntegrationResp
	err := row.Scan(
		&integration.ID, &integration.ProjectID, &integration.Kind, &integration.Name, &integration.Description,
		&integration.Status, &integration.EndpointURL, &integration.DocumentationURL, &integration.Transport,
		&integration.AuthType, &integration.RequestHeaders, &integration.AuthConfig, &integration.Metadata,
		&integration.LastSyncedAt, &integration.SyncStatus, &integration.SyncError,
		&integration.CreatedAt, &integration.UpdatedAt,
	)
	if len(integration.RequestHeaders) == 0 {
		integration.RequestHeaders = json.RawMessage(`{}`)
	}
	if len(integration.AuthConfig) == 0 {
		integration.AuthConfig = json.RawMessage(`{}`)
	}
	if len(integration.Metadata) == 0 {
		integration.Metadata = json.RawMessage(`{}`)
	}
	return integration, err
}

func normalizeCapabilityDefaults(req *projectCapabilityReq) {
	if req.Status == "" {
		req.Status = "active"
	}
	req.InputSchema = normalizedJSON(req.InputSchema)
	req.OutputSchema = normalizedJSON(req.OutputSchema)
	req.Metadata = normalizedJSON(req.Metadata)
}

func normalizeIntegrationDefaults(req *projectIntegrationReq) {
	if req.Status == "" {
		req.Status = "active"
	}
	if req.Transport == "" {
		req.Transport = "http"
	}
	if req.AuthType == "" {
		req.AuthType = "inherit"
	}
	req.RequestHeaders = normalizedJSON(req.RequestHeaders)
	req.AuthConfig = normalizedJSON(req.AuthConfig)
	req.Metadata = normalizedJSON(req.Metadata)
}

func getProjectHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	result, err := scanProject(tx.QueryRow(c.Request.Context(), queries.ProjectsSelectByID, projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	result.Integrations, err = listProjectIntegrations(c, tx, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result.Capabilities, err = listProjectCapabilities(c, tx, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func updateProjectHandler(c *gin.Context) {
	var req projectUpsertReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	normalizeProjectDefaults(&req)

	conn := tenant.ConnFromContext(c)

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	_, err = tx.Exec(c.Request.Context(),
		queries.ProjectsUpdate,
		req.Status, req.SourceType, req.Name, req.Description, req.EndpointURL,
		req.RequestMethod, req.RequestPath, req.RequestHeaders, req.RequestBodyTemplate,
		req.AuthType, req.AuthConfig, req.CapabilitySummary, projectID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project updated"})
}

func deleteProjectHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	_, err = conn.Exec(c.Request.Context(),
		queries.ProjectsDelete,
		projectID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "project deleted"})
}

func listProjectCapabilitiesHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	capabilities, err := listProjectCapabilities(c, tx, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"capabilities": capabilities})
}

func createProjectCapabilityHandler(c *gin.Context) {
	var req projectCapabilityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	normalizeCapabilityDefaults(&req)

	conn := tenant.ConnFromContext(c)
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var id uuid.UUID
	err = conn.QueryRow(c.Request.Context(),
		queries.ProjectCapabilitiesInsert,
		projectID, req.Kind, req.Name, req.Description, req.Status,
		req.RequestMethod, req.RequestPath,
		req.InputSchema, req.OutputSchema, req.Metadata,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func updateProjectCapabilityHandler(c *gin.Context) {
	var req projectCapabilityReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	normalizeCapabilityDefaults(&req)

	conn := tenant.ConnFromContext(c)
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	capabilityID, err := uuid.Parse(c.Param("capability_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid capability_id"})
		return
	}

	_, err = conn.Exec(c.Request.Context(),
		queries.ProjectCapabilitiesUpdate,
		req.Kind, req.Name, req.Description, req.Status, req.RequestMethod, req.RequestPath,
		req.InputSchema, req.OutputSchema, req.Metadata,
		projectID, capabilityID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "capability updated"})
}

func deleteProjectCapabilityHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	capabilityID, err := uuid.Parse(c.Param("capability_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid capability_id"})
		return
	}

	_, err = conn.Exec(c.Request.Context(), queries.ProjectCapabilitiesDelete, projectID, capabilityID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "capability deleted"})
}

func listProjectIntegrationsHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	integrations, err := listProjectIntegrations(c, tx, projectID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"integrations": integrations})
}

func createProjectIntegrationHandler(c *gin.Context) {
	var req projectIntegrationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	normalizeIntegrationDefaults(&req)

	conn := tenant.ConnFromContext(c)
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	var id uuid.UUID
	err = conn.QueryRow(c.Request.Context(),
		queries.ProjectIntegrationsInsert,
		projectID, req.Kind, req.Name, req.Description, req.Status,
		req.EndpointURL, req.DocumentationURL, req.Transport, req.AuthType,
		req.RequestHeaders, req.AuthConfig, req.Metadata,
	).Scan(&id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"id": id})
}

func updateProjectIntegrationHandler(c *gin.Context) {
	var req projectIntegrationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	normalizeIntegrationDefaults(&req)

	conn := tenant.ConnFromContext(c)
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	integrationID, err := uuid.Parse(c.Param("integration_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration_id"})
		return
	}

	_, err = conn.Exec(c.Request.Context(),
		queries.ProjectIntegrationsUpdate,
		req.Kind, req.Name, req.Description, req.Status, req.EndpointURL,
		req.DocumentationURL, req.Transport, req.AuthType,
		req.RequestHeaders, req.AuthConfig, req.Metadata,
		projectID, integrationID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "integration updated"})
}

func deleteProjectIntegrationHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	integrationID, err := uuid.Parse(c.Param("integration_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration_id"})
		return
	}

	_, err = conn.Exec(c.Request.Context(), queries.ProjectIntegrationsDelete, projectID, integrationID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "integration deleted"})
}

type mcpToolsListResponse struct {
	Result struct {
		Tools []struct {
			Name        string          `json:"name"`
			Description string          `json:"description"`
			InputSchema json.RawMessage `json:"inputSchema"`
		} `json:"tools"`
	} `json:"result"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

type syncedCapability struct {
	Kind          string
	Name          string
	ExternalName  string
	Description   string
	RequestMethod string
	RequestPath   string
	InputSchema   json.RawMessage
	OutputSchema  json.RawMessage
	Metadata      json.RawMessage
}

func syncProjectIntegrationHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}
	integrationID, err := uuid.Parse(c.Param("integration_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid integration_id"})
		return
	}

	tx, err := conn.Begin(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer func() { _ = tx.Rollback(c.Request.Context()) }()

	integration, err := scanProjectIntegration(tx.QueryRow(c.Request.Context(), queries.ProjectIntegrationsSelectByID, projectID, integrationID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}
	project, err := scanProject(tx.QueryRow(c.Request.Context(), queries.ProjectsSelectByID, projectID))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	capabilities, syncErr := syncIntegrationCapabilities(c, project, integration)
	if syncErr != nil {
		_, _ = tx.Exec(c.Request.Context(), queries.ProjectIntegrationsUpdateSyncResult, "failed", syncErr.Error(), projectID, integrationID)
		_ = tx.Commit(c.Request.Context())
		c.JSON(http.StatusBadGateway, gin.H{"error": syncErr.Error()})
		return
	}

	if _, err := tx.Exec(c.Request.Context(), queries.ProjectCapabilitiesDeleteByIntegration, projectID, integrationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	created := []uuid.UUID{}
	for _, capability := range capabilities {
		name := strings.TrimSpace(capability.Name)
		if name == "" {
			continue
		}

		var id uuid.UUID
		err := tx.QueryRow(c.Request.Context(),
			queries.ProjectCapabilitiesInsertForIntegration,
			projectID, integrationID, capability.Kind, name, capability.ExternalName, capability.Description, "active",
			capability.RequestMethod, capability.RequestPath,
			normalizedJSON(capability.InputSchema), normalizedJSON(capability.OutputSchema), normalizedJSON(capability.Metadata),
		).Scan(&id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		created = append(created, id)
	}

	if _, err := tx.Exec(c.Request.Context(), queries.ProjectIntegrationsUpdateSyncResult, "success", "", projectID, integrationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := tx.Commit(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"created": created,
		"count":   len(created),
		"message": "integration capabilities synchronized",
	})
}

func syncIntegrationCapabilities(c *gin.Context, project projectResp, integration projectIntegrationResp) ([]syncedCapability, error) {
	switch integration.Kind {
	case "mcp":
		return syncMCPIntegration(c, project, integration)
	case "api":
		return syncAPIIntegration(c, project, integration)
	case "skill":
		return syncSkillIntegration(c, project, integration)
	default:
		return nil, fmt.Errorf("unsupported integration kind: %s", integration.Kind)
	}
}

func syncMCPIntegration(c *gin.Context, project projectResp, integration projectIntegrationResp) ([]syncedCapability, error) {
	if integration.EndpointURL == nil || strings.TrimSpace(*integration.EndpointURL) == "" {
		return nil, fmt.Errorf("mcp endpoint_url is required")
	}

	tools, err := fetchMCPTools(c, project, integration)
	if err != nil {
		return nil, err
	}

	result := make([]syncedCapability, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			continue
		}
		result = append(result, syncedCapability{
			Kind:          "tool",
			Name:          name,
			ExternalName:  name,
			Description:   tool.Description,
			RequestMethod: "POST",
			InputSchema:   normalizedJSON(tool.InputSchema),
			OutputSchema:  json.RawMessage(`{}`),
			Metadata:      mustJSON(map[string]any{"source": "mcp-tools-list"}),
		})
	}
	return result, nil
}

func fetchMCPTools(c *gin.Context, project projectResp, integration projectIntegrationResp) ([]struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"inputSchema"`
}, error) {
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      "tools-list",
		"method":  "tools/list",
		"params":  map[string]any{},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, strings.TrimSpace(*integration.EndpointURL), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")

	applyHeaders(req, project.RequestHeaders)
	if integration.AuthType == "inherit" {
		applyAuth(req, project.AuthType, project.AuthConfig)
	}
	applyHeaders(req, integration.RequestHeaders)
	if integration.AuthType != "inherit" {
		applyAuth(req, integration.AuthType, integration.AuthConfig)
	}

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mcp tools/list returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var decoded mcpToolsListResponse
	if err := json.Unmarshal(extractJSONRPCPayload(respBody), &decoded); err != nil {
		return nil, err
	}
	if decoded.Error != nil {
		return nil, fmt.Errorf("mcp tools/list error %d: %s", decoded.Error.Code, decoded.Error.Message)
	}
	return decoded.Result.Tools, nil
}

func syncAPIIntegration(c *gin.Context, project projectResp, integration projectIntegrationResp) ([]syncedCapability, error) {
	docURL := integrationURL(integration.DocumentationURL, integration.EndpointURL)
	if docURL == "" {
		return nil, fmt.Errorf("api documentation_url is required")
	}

	doc, err := fetchStructuredDocument(c, project, integration, docURL)
	if err != nil {
		return nil, err
	}

	if openapi := getString(doc, "openapi"); strings.HasPrefix(openapi, "3.") {
		return capabilitiesFromOpenAPI3(doc)
	}
	if swagger := getString(doc, "swagger"); swagger == "2.0" {
		return capabilitiesFromSwagger2(doc)
	}
	return nil, fmt.Errorf("unsupported API spec: expected OpenAPI 3.x or Swagger 2.0")
}

func syncSkillIntegration(c *gin.Context, project projectResp, integration projectIntegrationResp) ([]syncedCapability, error) {
	docURL := integrationURL(integration.DocumentationURL, integration.EndpointURL)
	if docURL == "" {
		return nil, fmt.Errorf("skill documentation_url is required")
	}

	doc, err := fetchSkillDocument(c, project, integration, docURL)
	if err != nil {
		return nil, err
	}

	tools := getSlice(doc, "tools")
	if len(tools) == 0 {
		return nil, fmt.Errorf("skill manifest must include tools")
	}

	result := make([]syncedCapability, 0, len(tools))
	for _, item := range tools {
		tool, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := getString(tool, "name")
		if name == "" {
			continue
		}
		execution := getMap(tool, "execution")
		method := strings.ToUpper(getString(execution, "method"))
		if method == "" {
			method = "POST"
		}
		path := firstNonEmpty(getString(execution, "path"), getString(execution, "url"))
		result = append(result, syncedCapability{
			Kind:          "skill",
			Name:          name,
			ExternalName:  name,
			Description:   firstNonEmpty(getString(tool, "description"), getString(tool, "summary")),
			RequestMethod: method,
			RequestPath:   path,
			InputSchema:   rawObjectField(tool, "input_schema", "inputSchema", "schema"),
			OutputSchema:  rawObjectField(tool, "output_schema", "outputSchema"),
			Metadata: mustJSON(map[string]any{
				"source":         "t2-skill-manifest",
				"schema_version": getString(doc, "schema_version"),
				"skill_name":     getString(doc, "name"),
				"execution":      execution,
			}),
		})
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("skill manifest did not contain valid tools")
	}
	return result, nil
}

func fetchSkillDocument(c *gin.Context, project projectResp, integration projectIntegrationResp, docURL string) (map[string]any, error) {
	switch docURL {
	case "/api/v1/hub/skills/ticketmaster/manifest", "/hub/skills/ticketmaster/manifest", "builtin:ticketmaster":
		return ticketmasterSkillManifestDocument(), nil
	default:
		return fetchStructuredDocument(c, project, integration, docURL)
	}
}

func ticketmasterSkillManifestDocument() map[string]any {
	return map[string]any{
		"schema_version": "t2.skill.v1",
		"name":           "ticketmaster_events",
		"description":    "Fetch city-level event data from Ticketmaster Discovery API and normalize it into T2 event objects.",
		"tools": []any{
			map[string]any{
				"name":        "ticketmaster_search_events",
				"summary":     "Search Ticketmaster events for a city and date range",
				"description": "Returns normalized city event records sourced from Ticketmaster Discovery API.",
				"input_schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"city":       map[string]any{"type": "string", "description": "City name, for example New York or London."},
						"country":    map[string]any{"type": "string", "description": "Optional ISO 3166-1 alpha-2 country code, for example US or GB."},
						"start_date": map[string]any{"type": "string", "format": "date"},
						"end_date":   map[string]any{"type": "string", "format": "date"},
						"days":       map[string]any{"type": "integer", "description": "Defaults to 90 when dates are omitted."},
						"keyword":    map[string]any{"type": "string"},
						"category":   map[string]any{"type": "string"},
						"limit":      map[string]any{"type": "integer", "description": "Defaults to 100, max 200."},
					},
					"required": []any{"city"},
				},
				"output_schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"metadata": map[string]any{"type": "object"},
						"results":  map[string]any{"type": "array", "items": map[string]any{"type": "object"}},
					},
				},
				"execution": map[string]any{
					"type":   "http",
					"method": "POST",
					"path":   "/api/v1/hub/skills/ticketmaster/search-events",
				},
			},
		},
	}
}

func fetchStructuredDocument(c *gin.Context, project projectResp, integration projectIntegrationResp, docURL string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, docURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, application/yaml, text/yaml, */*")
	applyHeaders(req, project.RequestHeaders)
	if integration.AuthType == "inherit" {
		applyAuth(req, project.AuthType, project.AuthConfig)
	}
	applyHeaders(req, integration.RequestHeaders)
	if integration.AuthType != "inherit" {
		applyAuth(req, integration.AuthType, integration.AuthConfig)
	}

	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("document fetch returned HTTP %d: %s", resp.StatusCode, string(body))
	}

	doc, err := decodeStructuredDocument(body)
	if err != nil {
		return nil, err
	}
	return doc, nil
}

func decodeStructuredDocument(body []byte) (map[string]any, error) {
	var doc map[string]any
	if err := json.Unmarshal(body, &doc); err == nil {
		return doc, nil
	}

	var decoded any
	if err := yaml.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("document is not valid JSON or YAML")
	}
	normalized := normalizeYAMLValue(decoded)
	doc, ok := normalized.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("document root must be an object")
	}
	return doc, nil
}

func capabilitiesFromOpenAPI3(doc map[string]any) ([]syncedCapability, error) {
	paths := getMap(doc, "paths")
	if len(paths) == 0 {
		return nil, fmt.Errorf("openapi document has no paths")
	}

	result := []syncedCapability{}
	for path, rawPathItem := range paths {
		pathItem, ok := rawPathItem.(map[string]any)
		if !ok {
			continue
		}
		pathParameters := getSlice(pathItem, "parameters")
		for _, method := range []string{"get", "post", "put", "patch", "delete", "options", "head"} {
			operation := getMap(pathItem, method)
			if len(operation) == 0 {
				continue
			}
			name := operationName(method, path, operation)
			inputSchema := buildOpenAPIInputSchema(doc, append(pathParameters, getSlice(operation, "parameters")...), getMap(operation, "requestBody"))
			outputSchema := firstResponseSchema(doc, getMap(operation, "responses"))
			result = append(result, syncedCapability{
				Kind:          "api",
				Name:          name,
				ExternalName:  firstNonEmpty(getString(operation, "operationId"), name),
				Description:   firstNonEmpty(getString(operation, "summary"), getString(operation, "description")),
				RequestMethod: strings.ToUpper(method),
				RequestPath:   path,
				InputSchema:   inputSchema,
				OutputSchema:  outputSchema,
				Metadata: mustJSON(map[string]any{
					"source":           "openapi",
					"original_version": getString(doc, "openapi"),
					"operation_id":     getString(operation, "operationId"),
					"tags":             getSlice(operation, "tags"),
				}),
			})
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("openapi document did not contain operations")
	}
	return result, nil
}

func capabilitiesFromSwagger2(doc map[string]any) ([]syncedCapability, error) {
	paths := getMap(doc, "paths")
	if len(paths) == 0 {
		return nil, fmt.Errorf("swagger document has no paths")
	}

	basePath := getString(doc, "basePath")
	result := []syncedCapability{}
	for path, rawPathItem := range paths {
		pathItem, ok := rawPathItem.(map[string]any)
		if !ok {
			continue
		}
		pathParameters := getSlice(pathItem, "parameters")
		for _, method := range []string{"get", "post", "put", "patch", "delete", "options", "head"} {
			operation := getMap(pathItem, method)
			if len(operation) == 0 {
				continue
			}
			name := operationName(method, path, operation)
			inputSchema := buildSwaggerInputSchema(doc, append(pathParameters, getSlice(operation, "parameters")...))
			outputSchema := firstResponseSchema(doc, getMap(operation, "responses"))
			result = append(result, syncedCapability{
				Kind:          "api",
				Name:          name,
				ExternalName:  firstNonEmpty(getString(operation, "operationId"), name),
				Description:   firstNonEmpty(getString(operation, "summary"), getString(operation, "description")),
				RequestMethod: strings.ToUpper(method),
				RequestPath:   basePath + path,
				InputSchema:   inputSchema,
				OutputSchema:  outputSchema,
				Metadata: mustJSON(map[string]any{
					"source":           "swagger",
					"original_version": "2.0",
					"operation_id":     getString(operation, "operationId"),
					"tags":             getSlice(operation, "tags"),
					"host":             getString(doc, "host"),
					"schemes":          getSlice(doc, "schemes"),
				}),
			})
		}
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("swagger document did not contain operations")
	}
	return result, nil
}

func buildOpenAPIInputSchema(doc map[string]any, parameters []any, requestBody map[string]any) json.RawMessage {
	properties := map[string]any{}
	required := []string{}

	addParametersToSchema(doc, properties, &required, parameters)

	if len(requestBody) > 0 {
		schema := schemaFromContent(doc, getMap(requestBody, "content"))
		if len(schema) > 0 {
			properties["body"] = schema
			if requiredBool(requestBody) {
				required = append(required, "body")
			}
		}
	}

	return objectSchema(properties, required)
}

func buildSwaggerInputSchema(doc map[string]any, parameters []any) json.RawMessage {
	properties := map[string]any{}
	required := []string{}

	for _, item := range parameters {
		parameter, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := getString(parameter, "name")
		if name == "" {
			continue
		}
		location := getString(parameter, "in")
		if location == "body" {
			schema := resolveSchemaRef(doc, getMap(parameter, "schema"))
			if len(schema) > 0 {
				properties["body"] = schema
				if requiredBool(parameter) {
					required = append(required, "body")
				}
			}
			continue
		}
		schema := swaggerParameterSchema(parameter)
		key := parameterKey(location, name)
		properties[key] = schema
		if requiredBool(parameter) {
			required = append(required, key)
		}
	}

	return objectSchema(properties, required)
}

func addParametersToSchema(doc map[string]any, properties map[string]any, required *[]string, parameters []any) {
	for _, item := range parameters {
		parameter, ok := item.(map[string]any)
		if !ok {
			continue
		}
		name := getString(parameter, "name")
		if name == "" {
			continue
		}
		location := getString(parameter, "in")
		schema := resolveSchemaRef(doc, getMap(parameter, "schema"))
		if len(schema) == 0 {
			schema = map[string]any{"type": "string"}
		}
		key := parameterKey(location, name)
		properties[key] = schema
		if requiredBool(parameter) {
			*required = append(*required, key)
		}
	}
}

func firstResponseSchema(doc map[string]any, responses map[string]any) json.RawMessage {
	for _, status := range []string{"200", "201", "202", "default"} {
		response := getMap(responses, status)
		if len(response) == 0 {
			continue
		}
		if schema := schemaFromContent(doc, getMap(response, "content")); len(schema) > 0 {
			return mustJSON(schema)
		}
		if schema := resolveSchemaRef(doc, getMap(response, "schema")); len(schema) > 0 {
			return mustJSON(schema)
		}
	}
	return json.RawMessage(`{}`)
}

func schemaFromContent(doc map[string]any, content map[string]any) map[string]any {
	for _, contentType := range []string{"application/json", "application/problem+json"} {
		media := getMap(content, contentType)
		if schema := resolveSchemaRef(doc, getMap(media, "schema")); len(schema) > 0 {
			return schema
		}
	}
	for _, rawMedia := range content {
		media, ok := rawMedia.(map[string]any)
		if !ok {
			continue
		}
		if schema := resolveSchemaRef(doc, getMap(media, "schema")); len(schema) > 0 {
			return schema
		}
	}
	return nil
}

func swaggerParameterSchema(parameter map[string]any) map[string]any {
	schema := map[string]any{}
	for _, key := range []string{"type", "format", "enum", "items", "default", "minimum", "maximum"} {
		if value, ok := parameter[key]; ok {
			schema[key] = value
		}
	}
	if len(schema) == 0 {
		schema["type"] = "string"
	}
	return schema
}

func resolveSchemaRef(doc map[string]any, schema map[string]any) map[string]any {
	if len(schema) == 0 {
		return nil
	}
	ref := getString(schema, "$ref")
	if ref == "" || !strings.HasPrefix(ref, "#/") {
		return schema
	}
	current := any(doc)
	for _, part := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		currentMap, ok := current.(map[string]any)
		if !ok {
			return schema
		}
		current = currentMap[strings.ReplaceAll(part, "~1", "/")]
	}
	resolved, ok := current.(map[string]any)
	if !ok {
		return schema
	}
	return resolved
}

func objectSchema(properties map[string]any, required []string) json.RawMessage {
	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return mustJSON(schema)
}

func extractJSONRPCPayload(body []byte) []byte {
	trimmed := bytes.TrimSpace(body)
	if bytes.HasPrefix(trimmed, []byte("{")) {
		return trimmed
	}

	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		payload := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if payload == "" || payload == "[DONE]" {
			continue
		}
		if unquoted, err := strconv.Unquote(payload); err == nil {
			payload = unquoted
		}
		if strings.HasPrefix(payload, "{") {
			return []byte(payload)
		}
	}

	return trimmed
}

func integrationURL(values ...*string) string {
	for _, value := range values {
		if value != nil && strings.TrimSpace(*value) != "" {
			return strings.TrimSpace(*value)
		}
	}
	return ""
}

func operationName(method string, path string, operation map[string]any) string {
	if operationID := getString(operation, "operationId"); operationID != "" {
		return sanitizeCapabilityName(operationID)
	}
	replacer := strings.NewReplacer("/", "_", "{", "", "}", "", "-", "_", ".", "_")
	return sanitizeCapabilityName(method + replacer.Replace(path))
}

func sanitizeCapabilityName(value string) string {
	value = strings.TrimSpace(value)
	value = strings.ReplaceAll(value, " ", "_")
	value = strings.ReplaceAll(value, "-", "_")
	value = strings.Trim(value, "_")
	if value == "" {
		return "api_operation"
	}
	return value
}

func parameterKey(location string, name string) string {
	if location == "" || location == "query" {
		return name
	}
	return location + "." + name
}

func rawObjectField(item map[string]any, keys ...string) json.RawMessage {
	for _, key := range keys {
		if value, ok := item[key]; ok && value != nil {
			return mustJSON(value)
		}
	}
	return json.RawMessage(`{}`)
}

func mustJSON(value any) json.RawMessage {
	return jsonx.MustMarshalObject(value)
}

func normalizeYAMLValue(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		result := map[string]any{}
		for key, item := range typed {
			result[key] = normalizeYAMLValue(item)
		}
		return result
	case map[any]any:
		result := map[string]any{}
		for key, item := range typed {
			result[fmt.Sprint(key)] = normalizeYAMLValue(item)
		}
		return result
	case []any:
		for i := range typed {
			typed[i] = normalizeYAMLValue(typed[i])
		}
		return typed
	default:
		return value
	}
}

func getMap(item map[string]any, key string) map[string]any {
	value, ok := item[key]
	if !ok || value == nil {
		return map[string]any{}
	}
	typed, ok := value.(map[string]any)
	if ok {
		return typed
	}
	return map[string]any{}
}

func getSlice(item map[string]any, key string) []any {
	value, ok := item[key]
	if !ok || value == nil {
		return nil
	}
	typed, ok := value.([]any)
	if ok {
		return typed
	}
	return nil
}

func getString(item map[string]any, key string) string {
	value, ok := item[key]
	if !ok || value == nil {
		return ""
	}
	switch typed := value.(type) {
	case string:
		return typed
	default:
		return fmt.Sprint(typed)
	}
}

func requiredBool(item map[string]any) bool {
	value, ok := item["required"].(bool)
	return ok && value
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func applyHeaders(req *http.Request, raw json.RawMessage) {
	integration.ApplyRawHeaders(req.Header, raw)
}

func applyAuth(req *http.Request, authType string, rawConfig json.RawMessage) {
	integration.ApplyAuth(req, integration.RawAuthConfig(authType, rawConfig))
}
