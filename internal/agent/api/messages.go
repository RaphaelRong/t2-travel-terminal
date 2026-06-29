package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/runtime"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

type messageResponse struct {
	ID               string            `json:"id"`
	Role             string            `json:"role"`
	Content          string            `json:"content"`
	ToolCalls        []domain.ToolCall `json:"tool_calls,omitempty"`
	ToolCallID       string            `json:"tool_call_id,omitempty"`
	ToolName         string            `json:"tool_name,omitempty"`
	ToolResult       map[string]any    `json:"tool_result,omitempty"`
	ReasoningContent string            `json:"reasoning_content,omitempty"`
	CreatedAt        string            `json:"created_at"`
}

type sendMessageRequest struct {
	Content      string     `json:"content" binding:"required"`
	LLMProfileID *uuid.UUID `json:"llm_profile_id,omitempty"`
	Model        string     `json:"model,omitempty"`
}

func (h *Handler) ListMessages(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	conn := tenant.ConnFromContext(c)
	messages, err := h.sessionStore.ListMessagesBySession(c.Request.Context(), conn, sessionID, 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"messages": toMessageResponses(messages)})
}

func (h *Handler) SendMessage(c *gin.Context) {
	sessionID, err := uuid.Parse(c.Param("session_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid session_id"})
		return
	}

	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	req.Content = strings.TrimSpace(req.Content)
	if req.Content == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "content is required"})
		return
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	// Phase 2：加载 GodScope 并对用户输入做范围初筛。
	godScope, err := h.godLoader.LoadForUser(c.Request.Context(), conn, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := validateUserMessageAgainstScope(godScope, req.Content); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// 保存用户消息
	userMsg := &domain.Message{
		SessionID: sessionID,
		Role:      domain.MessageRoleUser,
		Content:   req.Content,
	}
	if err := h.sessionStore.CreateMessage(c.Request.Context(), conn, userMsg); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 加载会话以获取 tenant_id
	session, err := h.sessionStore.GetSessionByID(c.Request.Context(), conn, sessionID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}

	// 加载用户 profile 和 soul
	profile, err := h.profileStore.GetByUserID(c.Request.Context(), conn, userID)
	if err != nil {
		profile = &domain.UserProfile{UserID: userID}
	}

	var llmProfileID uuid.UUID
	if req.LLMProfileID != nil {
		llmProfileID = *req.LLMProfileID
	} else if profile.DefaultLLMProfileID != nil {
		llmProfileID = *profile.DefaultLLMProfileID
	}

	var soul *domain.Soul
	souls, _ := h.soulStore.ListByUser(c.Request.Context(), conn, userID)
	if len(souls) > 0 {
		soul = souls[0]
	}

	// 拉取历史消息
	history, err := h.sessionStore.ListMessagesBySession(c.Request.Context(), conn, sessionID, 1000)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	runReq := runtime.RunRequest{
		UserID:        userID,
		TenantID:      session.TenantID,
		SessionID:     sessionID,
		Profile:       profile,
		GodScope:      godScope,
		Soul:          soul,
		Messages:      history,
		LLMProfileID:  llmProfileID,
		OverrideModel: req.Model,
	}

	// SSE 输出
	events := make(chan runtime.StreamEvent, 16)
	ctx := c.Request.Context()

	go func() {
		defer close(events)
		_, _ = h.runner.Run(ctx, conn, runReq, events)
	}()

	// 设置 SSE 响应头
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)

	// 先推送用户消息
	writeSSEEvent(c.Writer, "user_message", toMessageResponse(userMsg))
	c.Writer.Flush()

	// 推送运行过程中的事件
	for event := range events {
		writeSSEEvent(c.Writer, string(event.Type), event.Data)
		c.Writer.Flush()
	}
}

func writeSSEEvent(w io.Writer, event string, data interface{}) {
	b, _ := json.Marshal(data)
	_, _ = fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(b))
}

func toMessageResponse(m *domain.Message) messageResponse {
	resp := messageResponse{
		ID:               m.ID.String(),
		Role:             string(m.Role),
		Content:          m.Content,
		ToolCalls:        m.ToolCalls,
		ToolCallID:       m.ToolCallID,
		ToolName:         m.ToolName,
		ToolResult:       m.ToolResult,
		ReasoningContent: m.ReasoningContent,
		CreatedAt:        m.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
	return resp
}

func toMessageResponses(messages []*domain.Message) []messageResponse {
	result := make([]messageResponse, len(messages))
	for i, m := range messages {
		result[i] = toMessageResponse(m)
	}
	return result
}

// validateUserMessageAgainstScope 对用户输入做简单的 GodScope 初筛。
func validateUserMessageAgainstScope(scope *domain.GodScope, content string) error {
	lower := strings.ToLower(content)
	for _, domain := range scope.ForbiddenDomains {
		if domain != "" && strings.Contains(lower, strings.ToLower(domain)) {
			return fmt.Errorf("message touches forbidden domain: %s", domain)
		}
	}
	return nil
}
