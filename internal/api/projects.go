package api

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/jsonx"
	projectsync "github.com/t2-travel-terminal/t2-travel-terminal/internal/project"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
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

func listProjectCapabilities(c *gin.Context, tx pgx.Tx, projectID uuid.UUID) ([]projectCapabilityResp, error) {
	capabilities, err := projectsync.NewSyncService().ListCapabilities(c.Request.Context(), tx, projectID)
	if err != nil {
		return nil, err
	}

	result := make([]projectCapabilityResp, 0, len(capabilities))
	for _, capability := range capabilities {
		result = append(result, toProjectCapabilityResp(capability))
	}
	return result, nil
}

func listProjectIntegrations(c *gin.Context, tx pgx.Tx, projectID uuid.UUID) ([]projectIntegrationResp, error) {
	integrations, err := projectsync.NewSyncService().ListIntegrations(c.Request.Context(), tx, projectID)
	if err != nil {
		return nil, err
	}

	result := make([]projectIntegrationResp, 0, len(integrations))
	for _, integration := range integrations {
		result = append(result, toProjectIntegrationResp(integration))
	}
	return result, nil
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

func normalizeCapabilityDefaults(req *projectCapabilityReq) {
	if req.Status == "" {
		req.Status = "active"
	}
	req.InputSchema = normalizedJSON(req.InputSchema)
	req.OutputSchema = normalizedJSON(req.OutputSchema)
	req.Metadata = normalizedJSON(req.Metadata)
}

func capabilityInputFromReq(req projectCapabilityReq) projectsync.CapabilityInput {
	return projectsync.CapabilityInput{
		Kind:          req.Kind,
		Name:          req.Name,
		Description:   req.Description,
		Status:        req.Status,
		RequestMethod: req.RequestMethod,
		RequestPath:   req.RequestPath,
		InputSchema:   req.InputSchema,
		OutputSchema:  req.OutputSchema,
		Metadata:      req.Metadata,
	}
}

func toProjectCapabilityResp(capability projectsync.CapabilityRecord) projectCapabilityResp {
	return projectCapabilityResp{
		ID:            capability.ID,
		ProjectID:     capability.ProjectID,
		IntegrationID: capability.IntegrationID,
		Kind:          capability.Kind,
		Name:          capability.Name,
		Description:   capability.Description,
		Status:        capability.Status,
		RequestMethod: capability.RequestMethod,
		RequestPath:   capability.RequestPath,
		InputSchema:   capability.InputSchema,
		OutputSchema:  capability.OutputSchema,
		Metadata:      capability.Metadata,
		CreatedAt:     capability.CreatedAt,
		UpdatedAt:     capability.UpdatedAt,
	}
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

func integrationInputFromReq(req projectIntegrationReq) projectsync.IntegrationInput {
	return projectsync.IntegrationInput{
		Kind:             req.Kind,
		Name:             req.Name,
		Description:      req.Description,
		Status:           req.Status,
		EndpointURL:      req.EndpointURL,
		DocumentationURL: req.DocumentationURL,
		Transport:        req.Transport,
		AuthType:         req.AuthType,
		RequestHeaders:   req.RequestHeaders,
		AuthConfig:       req.AuthConfig,
		Metadata:         req.Metadata,
	}
}

func toProjectIntegrationResp(integration projectsync.IntegrationRecord) projectIntegrationResp {
	return projectIntegrationResp{
		ID:               integration.ID,
		ProjectID:        integration.ProjectID,
		Kind:             integration.Kind,
		Name:             integration.Name,
		Description:      integration.Description,
		Status:           integration.Status,
		EndpointURL:      integration.EndpointURL,
		DocumentationURL: integration.DocumentationURL,
		Transport:        integration.Transport,
		AuthType:         integration.AuthType,
		RequestHeaders:   integration.RequestHeaders,
		AuthConfig:       integration.AuthConfig,
		Metadata:         integration.Metadata,
		LastSyncedAt:     integration.LastSyncedAt,
		SyncStatus:       integration.SyncStatus,
		SyncError:        integration.SyncError,
		CreatedAt:        integration.CreatedAt,
		UpdatedAt:        integration.UpdatedAt,
	}
}

func projectInputFromReq(req projectUpsertReq, tenantID uuid.UUID, sourceScope string, createdBy uuid.UUID) projectsync.ProjectInput {
	return projectsync.ProjectInput{
		TenantID:            tenantID,
		SourceScope:         sourceScope,
		Status:              req.Status,
		SourceType:          req.SourceType,
		Name:                req.Name,
		Description:         req.Description,
		EndpointURL:         req.EndpointURL,
		RequestMethod:       req.RequestMethod,
		RequestPath:         req.RequestPath,
		RequestHeaders:      req.RequestHeaders,
		RequestBodyTemplate: req.RequestBodyTemplate,
		AuthType:            req.AuthType,
		AuthConfig:          req.AuthConfig,
		CapabilitySummary:   req.CapabilitySummary,
		CreatedBy:           createdBy,
	}
}

func toProjectResp(project projectsync.ProjectRecord) projectResp {
	return projectResp{
		ID:                  project.ID,
		TenantID:            project.TenantID,
		SourceScope:         project.SourceScope,
		Kind:                project.Kind,
		Status:              project.Status,
		SourceType:          project.SourceType,
		Name:                project.Name,
		Description:         project.Description,
		EndpointURL:         project.EndpointURL,
		RequestMethod:       project.RequestMethod,
		RequestPath:         project.RequestPath,
		RequestHeaders:      project.RequestHeaders,
		RequestBodyTemplate: project.RequestBodyTemplate,
		AuthType:            project.AuthType,
		AuthConfig:          project.AuthConfig,
		CapabilitySummary:   project.CapabilitySummary,
		CreatedBy:           project.CreatedBy,
		CreatedAt:           project.CreatedAt,
		UpdatedAt:           project.UpdatedAt,
		LastPublishedAt:     project.LastPublishedAt,
	}
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

	project, err := projectsync.NewSyncService().GetProject(c.Request.Context(), tx, projectID)
	if errors.Is(err, projectsync.ErrProjectNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	result := toProjectResp(project)

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

	if err := projectsync.NewSyncService().UpdateProject(c.Request.Context(), tx, projectID, projectInputFromReq(req, uuid.Nil, "", uuid.Nil)); err != nil {
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

	if err := projectsync.NewSyncService().DeleteProject(c.Request.Context(), conn, projectID); err != nil {
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

	id, err := projectsync.NewSyncService().CreateCapability(c.Request.Context(), conn, projectID, capabilityInputFromReq(req))
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

	if err := projectsync.NewSyncService().UpdateCapability(c.Request.Context(), conn, projectID, capabilityID, capabilityInputFromReq(req)); err != nil {
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

	if err := projectsync.NewSyncService().DeleteCapability(c.Request.Context(), conn, projectID, capabilityID); err != nil {
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

	id, err := projectsync.NewSyncService().CreateIntegration(c.Request.Context(), conn, projectID, integrationInputFromReq(req))
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

	if err := projectsync.NewSyncService().UpdateIntegration(c.Request.Context(), conn, projectID, integrationID, integrationInputFromReq(req)); err != nil {
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

	if err := projectsync.NewSyncService().DeleteIntegration(c.Request.Context(), conn, projectID, integrationID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "integration deleted"})
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

	created, err := projectsync.NewSyncService().SyncIntegration(c.Request.Context(), tx, projectID, integrationID)
	if errors.Is(err, projectsync.ErrIntegrationNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "integration not found"})
		return
	}
	if errors.Is(err, projectsync.ErrProjectNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}
	var syncFailure *projectsync.SyncFailure
	if errors.As(err, &syncFailure) {
		_ = tx.Commit(c.Request.Context())
		c.JSON(http.StatusBadGateway, gin.H{"error": syncFailure.Error()})
		return
	}
	if err != nil {
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
