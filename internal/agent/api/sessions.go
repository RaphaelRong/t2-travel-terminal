package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/god"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

type sessionResponse struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	Status          string `json:"status"`
	ParentSessionID string `json:"parent_session_id,omitempty"`
	ContextSummary  string `json:"context_summary,omitempty"`
	CreatedAt       string `json:"created_at"`
	UpdatedAt       string `json:"updated_at"`
}

type sessionCreateRequest struct {
	Title           string      `json:"title"`
	ProjectIDs      []uuid.UUID `json:"project_ids"`
	SoulID          *uuid.UUID  `json:"soul_id"`
	LLMProfileID    *uuid.UUID  `json:"llm_profile_id"`
	ParentSessionID *uuid.UUID  `json:"parent_session_id"`
}

type sessionListQuery struct {
	Status string `form:"status"`
	Limit  int    `form:"limit"`
}

type sessionProjectResponse struct {
	ID        string `json:"id"`
	ProjectID string `json:"project_id"`
	Name      string `json:"name"`
	Kind      string `json:"kind"`
	Status    string `json:"status"`
	AddedAt   string `json:"added_at"`
}

type capabilityResponse struct {
	ID            string                 `json:"id"`
	ProjectID     string                 `json:"project_id"`
	IntegrationID string                 `json:"integration_id,omitempty"`
	Kind          string                 `json:"kind"`
	Name          string                 `json:"name"`
	ExternalName  string                 `json:"external_name,omitempty"`
	Description   string                 `json:"description,omitempty"`
	Status        string                 `json:"status"`
	RequestMethod string                 `json:"request_method,omitempty"`
	RequestPath   string                 `json:"request_path,omitempty"`
	InputSchema   map[string]interface{} `json:"input_schema,omitempty"`
	OutputSchema  map[string]interface{} `json:"output_schema,omitempty"`
}

func (h *Handler) ListSessions(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	var q sessionListQuery
	if err := c.ShouldBindQuery(&q); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if q.Status == "" {
		q.Status = string(domain.SessionStatusActive)
	}
	if q.Limit <= 0 {
		q.Limit = 50
	}

	sessions, err := h.sessionStore.ListSessionsByUser(
		c.Request.Context(), conn, userID, domain.SessionStatus(q.Status), q.Limit,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"sessions": toSessionResponses(sessions)})
}

func (h *Handler) CreateSession(c *gin.Context) {
	var req sessionCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		req.Title = "New Session"
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	session := &domain.Session{
		UserID:          userID,
		Title:           req.Title,
		Status:          domain.SessionStatusActive,
		ParentSessionID: req.ParentSessionID,
	}

	if tenantID, ok := getCurrentTenantID(c); ok {
		session.TenantID = tenantID
	}

	if err := h.sessionStore.CreateSession(c.Request.Context(), conn, session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	for _, projectID := range req.ProjectIDs {
		if err := h.sessionStore.AttachProject(c.Request.Context(), conn, session.ID, projectID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	// Phase 2：创建 system message，注入 GodScope + Soul 规则。
	goScope, _ := h.godLoader.LoadForUser(c.Request.Context(), conn, userID)
	if goScope == nil {
		goScope = god.SystemDefault()
	}

	var userSoul *domain.Soul
	souls, _ := h.soulStore.ListByUser(c.Request.Context(), conn, userID)
	if len(souls) > 0 {
		userSoul = souls[0]
	}

	systemPrompt := god.BuildSystemPrompt(goScope, userSoul)
	systemMsg := &domain.Message{
		SessionID: session.ID,
		Role:      domain.MessageRoleSystem,
		Content:   systemPrompt,
	}
	_ = h.sessionStore.CreateMessage(c.Request.Context(), conn, systemMsg)

	c.JSON(http.StatusCreated, gin.H{"session": toSessionResponse(session)})
}

func (h *Handler) GetSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	session, err := h.sessionStore.GetSessionByID(c.Request.Context(), conn, sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"session": toSessionResponse(session)})
}

func (h *Handler) ArchiveSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	session, err := h.sessionStore.GetSessionByID(c.Request.Context(), conn, sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	session.Status = domain.SessionStatusArchived
	if err := h.sessionStore.UpdateSession(c.Request.Context(), conn, session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session archived"})
}

func (h *Handler) DeleteSession(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	session, err := h.sessionStore.GetSessionByID(c.Request.Context(), conn, sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	session.Status = domain.SessionStatusDeleted
	if err := h.sessionStore.UpdateSession(c.Request.Context(), conn, session); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "session deleted"})
}

func (h *Handler) AttachProject(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	var req struct {
		ProjectID uuid.UUID `json:"project_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	conn := tenant.ConnFromContext(c)
	if _, err := h.sessionStore.GetSessionByID(c.Request.Context(), conn, sessionID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	if _, err := h.projectStore.GetProjectByID(c.Request.Context(), conn, req.ProjectID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
		return
	}

	if err := h.sessionStore.AttachProject(c.Request.Context(), conn, sessionID, req.ProjectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "project attached"})
}

func (h *Handler) DetachProject(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}
	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid project_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	if err := h.sessionStore.DetachProject(c.Request.Context(), conn, sessionID, projectID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "project detached"})
}

func (h *Handler) ListSessionProjects(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	projects, err := h.projectStore.ListBySession(c.Request.Context(), conn, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"projects": toProjectResponses(projects)})
}

func (h *Handler) ListSessionCapabilities(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	caps, _, _, err := h.projectStore.ListCapabilitiesBySession(c.Request.Context(), conn, sessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"capabilities": toCapabilityResponses(caps)})
}

func toProjectResponses(projects []*domain.Project) []sessionProjectResponse {
	result := make([]sessionProjectResponse, len(projects))
	for i, p := range projects {
		result[i] = sessionProjectResponse{
			ID:        p.ID.String(),
			ProjectID: p.ID.String(),
			Name:      p.Name,
			Kind:      p.Kind,
			Status:    p.Status,
			AddedAt:   p.CreatedAt.Format("2006-01-02T15:04:05Z"),
		}
	}
	return result
}

func toCapabilityResponses(caps []*domain.ProjectCapability) []capabilityResponse {
	result := make([]capabilityResponse, len(caps))
	for i, c := range caps {
		resp := capabilityResponse{
			ID:            c.ID.String(),
			ProjectID:     c.ProjectID.String(),
			Kind:          string(c.Kind),
			Name:          c.Name,
			ExternalName:  c.ExternalName,
			Description:   c.Description,
			Status:        c.Status,
			RequestMethod: c.RequestMethod,
			RequestPath:   c.RequestPath,
			InputSchema:   c.InputSchema,
			OutputSchema:  c.OutputSchema,
		}
		if c.IntegrationID != nil {
			resp.IntegrationID = c.IntegrationID.String()
		}
		result[i] = resp
	}
	return result
}

func toSessionResponse(s *domain.Session) sessionResponse {
	resp := sessionResponse{
		ID:        s.ID.String(),
		Title:     s.Title,
		Status:    string(s.Status),
		CreatedAt: s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if s.ParentSessionID != nil {
		resp.ParentSessionID = s.ParentSessionID.String()
	}
	if s.ContextSummary != "" {
		resp.ContextSummary = s.ContextSummary
	}
	return resp
}

func toSessionResponses(sessions []*domain.Session) []sessionResponse {
	result := make([]sessionResponse, len(sessions))
	for i, s := range sessions {
		result[i] = toSessionResponse(s)
	}
	return result
}

func getCurrentTenantID(c *gin.Context) (*uuid.UUID, bool) {
	t, ok := tenant.TenantFromContext(c)
	if !ok {
		return nil, false
	}
	return &t.ID, true
}
