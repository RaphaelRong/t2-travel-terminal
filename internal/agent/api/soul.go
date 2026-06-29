package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

type soulResponse struct {
	ID               string   `json:"id"`
	Name             string   `json:"name"`
	IdentityText     string   `json:"identity_text"`
	VoiceText        string   `json:"voice_text,omitempty"`
	ValuesText       string   `json:"values_text,omitempty"`
	AllowedDomains   []string `json:"allowed_domains,omitempty"`
	ForbiddenDomains []string `json:"forbidden_domains,omitempty"`
	CreatedAt        string   `json:"created_at"`
	UpdatedAt        string   `json:"updated_at"`
}

type soulUpdateRequest struct {
	Name             string   `json:"name"`
	IdentityText     string   `json:"identity_text"`
	VoiceText        string   `json:"voice_text"`
	ValuesText       string   `json:"values_text"`
	AllowedDomains   []string `json:"allowed_domains"`
	ForbiddenDomains []string `json:"forbidden_domains"`
}

func (h *Handler) GetSoul(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	souls, err := h.soulStore.ListByUser(c.Request.Context(), conn, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if len(souls) == 0 {
		c.JSON(http.StatusOK, gin.H{"soul": nil})
		return
	}

	c.JSON(http.StatusOK, gin.H{"soul": toSoulResponse(souls[0])})
}

func (h *Handler) UpdateSoul(c *gin.Context) {
	var req soulUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.IdentityText = strings.TrimSpace(req.IdentityText)
	if req.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.IdentityText == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "identity_text is required"})
		return
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	souls, err := h.soulStore.ListByUser(c.Request.Context(), conn, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var soul *domain.Soul
	if len(souls) > 0 {
		soul = souls[0]
		soul.Name = req.Name
		soul.IdentityText = req.IdentityText
		soul.VoiceText = req.VoiceText
		soul.ValuesText = req.ValuesText
		soul.AllowedDomains = req.AllowedDomains
		soul.ForbiddenDomains = req.ForbiddenDomains
		if err := h.soulStore.Update(c.Request.Context(), conn, soul); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	} else {
		soul = &domain.Soul{
			Scope:            domain.SoulScopeUser,
			UserID:           &userID,
			Name:             req.Name,
			IdentityText:     req.IdentityText,
			VoiceText:        req.VoiceText,
			ValuesText:       req.ValuesText,
			AllowedDomains:   req.AllowedDomains,
			ForbiddenDomains: req.ForbiddenDomains,
		}
		if err := h.soulStore.Create(c.Request.Context(), conn, soul); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"soul": toSoulResponse(soul)})
}

func toSoulResponse(s *domain.Soul) soulResponse {
	return soulResponse{
		ID:               s.ID.String(),
		Name:             s.Name,
		IdentityText:     s.IdentityText,
		VoiceText:        s.VoiceText,
		ValuesText:       s.ValuesText,
		AllowedDomains:   s.AllowedDomains,
		ForbiddenDomains: s.ForbiddenDomains,
		CreatedAt:        s.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:        s.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
}
