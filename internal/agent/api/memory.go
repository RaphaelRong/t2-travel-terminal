package api

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

type memoryResponse struct {
	ID              string   `json:"id"`
	Category        string   `json:"category"`
	Content         string   `json:"content"`
	SourceSessionID *string  `json:"source_session_id,omitempty"`
	Confidence      float64  `json:"confidence"`
	CreatedAt       string   `json:"created_at"`
	UpdatedAt       string   `json:"updated_at"`
}

type memoryCreateRequest struct {
	Category   string  `json:"category"`
	Content    string  `json:"content"`
	Confidence float64 `json:"confidence"`
}

type memorySearchRequest struct {
	Query string `json:"query"`
	Limit int    `json:"limit"`
}

func (h *Handler) ListMemory(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	memories, err := h.memoryStore.ListByUser(c.Request.Context(), conn, userID, 100)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"memories": toMemoryResponses(memories)})
}

func (h *Handler) SearchMemory(c *gin.Context) {
	var req memorySearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query is required"})
		return
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	memories, err := h.memoryStore.Search(c.Request.Context(), conn, userID, req.Query, req.Limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"memories": toMemoryResponses(memories)})
}

func (h *Handler) CreateMemory(c *gin.Context) {
	var req memoryCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	category := domain.MemoryCategory(req.Category)
	if category == "" {
		category = domain.MemoryCategoryFact
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	memory := &domain.Memory{
		UserID:     userID,
		Category:   category,
		Content:    req.Content,
		Confidence: req.Confidence,
	}
	if memory.Confidence <= 0 {
		memory.Confidence = 1.0
	}

	if err := h.memoryStore.Create(c.Request.Context(), conn, memory); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"memory": toMemoryResponse(memory)})
}

func (h *Handler) DeleteMemory(c *gin.Context) {
	memoryID, err := uuid.Parse(c.Param("memory_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid memory_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	if err := h.memoryStore.Delete(c.Request.Context(), conn, memoryID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "memory deleted"})
}

func toMemoryResponse(m *domain.Memory) memoryResponse {
	resp := memoryResponse{
		ID:         m.ID.String(),
		Category:   string(m.Category),
		Content:    m.Content,
		Confidence: m.Confidence,
		CreatedAt:  m.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:  m.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}
	if m.SourceSessionID != nil {
		s := m.SourceSessionID.String()
		resp.SourceSessionID = &s
	}
	return resp
}

func toMemoryResponses(memories []*domain.Memory) []memoryResponse {
	result := make([]memoryResponse, len(memories))
	for i, m := range memories {
		result[i] = toMemoryResponse(m)
	}
	return result
}
