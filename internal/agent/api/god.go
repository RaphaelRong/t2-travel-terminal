package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

type godConfigResponse struct {
	ID                   string   `json:"id"`
	Name                 string   `json:"name"`
	IsActive             bool     `json:"is_active"`
	AllowedDomains       []string `json:"allowed_domains,omitempty"`
	ForbiddenDomains     []string `json:"forbidden_domains,omitempty"`
	AllowedTools         []string `json:"allowed_tools,omitempty"`
	ForbiddenTools       []string `json:"forbidden_tools,omitempty"`
	RequireApprovalTools []string `json:"require_approval_tools,omitempty"`
	MaxIterations        int      `json:"max_iterations"`
	CanDelegate          bool     `json:"can_delegate"`
	CanRunWorkflow       bool     `json:"can_run_workflow"`
	Rules                string   `json:"rules"`
	CreatedAt            string   `json:"created_at"`
	UpdatedAt            string   `json:"updated_at"`
}

type godConfigUpdateRequest struct {
	Name                 string   `json:"name"`
	IsActive             bool     `json:"is_active"`
	AllowedDomains       []string `json:"allowed_domains"`
	ForbiddenDomains     []string `json:"forbidden_domains"`
	AllowedTools         []string `json:"allowed_tools"`
	ForbiddenTools       []string `json:"forbidden_tools"`
	RequireApprovalTools []string `json:"require_approval_tools"`
	MaxIterations        int      `json:"max_iterations"`
	CanDelegate          bool     `json:"can_delegate"`
	CanRunWorkflow       bool     `json:"can_run_workflow"`
	Rules                string   `json:"rules"`
}

func (h *Handler) GetGodConfig(c *gin.Context) {
	conn := tenant.ConnFromContext(c)

	config, err := h.godStore.GetActive(c.Request.Context(), conn)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"config": toGodConfigResponse(&domain.GodConfig{
			Name:          "default",
			MaxIterations: 30,
		})})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": toGodConfigResponse(config)})
}

func (h *Handler) UpdateGodConfig(c *gin.Context) {
	var req godConfigUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if req.Name == "" {
		req.Name = "default"
	}
	if req.MaxIterations <= 0 {
		req.MaxIterations = 30
	}

	conn := tenant.ConnFromContext(c)

	existing, err := h.godStore.GetByName(c.Request.Context(), conn, req.Name)
	if err != nil {
		// 不存在则创建
		config := &domain.GodConfig{
			Name:                 req.Name,
			IsActive:             req.IsActive,
			AllowedDomains:       req.AllowedDomains,
			ForbiddenDomains:     req.ForbiddenDomains,
			AllowedTools:         req.AllowedTools,
			ForbiddenTools:       req.ForbiddenTools,
			RequireApprovalTools: req.RequireApprovalTools,
			MaxIterations:        req.MaxIterations,
			CanDelegate:          req.CanDelegate,
			CanRunWorkflow:       req.CanRunWorkflow,
			Rules:                req.Rules,
		}
		if err := h.godStore.Create(c.Request.Context(), conn, config); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"config": toGodConfigResponse(config)})
		return
	}

	existing.IsActive = req.IsActive
	existing.AllowedDomains = req.AllowedDomains
	existing.ForbiddenDomains = req.ForbiddenDomains
	existing.AllowedTools = req.AllowedTools
	existing.ForbiddenTools = req.ForbiddenTools
	existing.RequireApprovalTools = req.RequireApprovalTools
	existing.MaxIterations = req.MaxIterations
	existing.CanDelegate = req.CanDelegate
	existing.CanRunWorkflow = req.CanRunWorkflow
	existing.Rules = req.Rules

	if err := h.godStore.Update(c.Request.Context(), conn, existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"config": toGodConfigResponse(existing)})
}

func toGodConfigResponse(c *domain.GodConfig) godConfigResponse {
	return godConfigResponse{
		ID:                   c.ID.String(),
		Name:                 c.Name,
		IsActive:             c.IsActive,
		AllowedDomains:       c.AllowedDomains,
		ForbiddenDomains:     c.ForbiddenDomains,
		AllowedTools:         c.AllowedTools,
		ForbiddenTools:       c.ForbiddenTools,
		RequireApprovalTools: c.RequireApprovalTools,
		MaxIterations:        c.MaxIterations,
		CanDelegate:          c.CanDelegate,
		CanRunWorkflow:       c.CanRunWorkflow,
		Rules:                c.Rules,
		CreatedAt:            c.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:            c.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// AdminGodConfigIDParam 解析 URL 中的 config id（预留）。
func AdminGodConfigIDParam(c *gin.Context) (uuid.UUID, error) {
	return uuid.Parse(c.Param("config_id"))
}
