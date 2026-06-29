package tools

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/integration"
)

type mcpCallResponse struct {
	Result map[string]any `json:"result"`
	Error  *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func (t *ProjectTool) isMCPTool() bool {
	if t.integration != nil && t.integration.Kind == "mcp" {
		return true
	}
	if source, ok := t.capability.Metadata["source"].(string); ok && source == "mcp-tools-list" {
		return true
	}
	return false
}

func (t *ProjectTool) executeMCP(ctx context.Context, args map[string]any, runCtx *Context) (any, error) {
	baseURL, err := t.resolveBaseURL("")
	if err != nil {
		return nil, err
	}

	toolName := strings.TrimSpace(t.capability.ExternalName)
	if toolName == "" {
		toolName = t.capability.Name
	}
	payload := map[string]any{
		"jsonrpc": "2.0",
		"id":      "tool-call",
		"method":  "tools/call",
		"params": map[string]any{
			"name":      toolName,
			"arguments": args,
		},
	}

	req, err := integration.NewJSONRequest(ctx, http.MethodPost, baseURL, payload)
	if err != nil {
		return nil, fmt.Errorf("build mcp request: %w", err)
	}
	req.Header.Set("Accept", "application/json, text/event-stream")
	t.applyRunHeaders(req, baseURL, runCtx)
	t.applyProjectAuth(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute mcp tool: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read mcp response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("mcp tool returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var decoded mcpCallResponse
	if err := json.Unmarshal(extractJSONRPCPayload(respBody), &decoded); err != nil {
		return nil, fmt.Errorf("decode mcp response: %w", err)
	}
	if decoded.Error != nil {
		return nil, fmt.Errorf("mcp tool error %d: %s", decoded.Error.Code, decoded.Error.Message)
	}
	if isError, _ := decoded.Result["isError"].(bool); isError {
		return nil, fmt.Errorf("mcp tool returned error content: %v", decoded.Result["content"])
	}
	return decoded.Result, nil
}

func (t *ProjectTool) applyRunHeaders(req *http.Request, baseURL string, runCtx *Context) {
	if baseURL != t.localBaseURL || runCtx == nil {
		return
	}
	if runCtx.UserID != uuid.Nil {
		req.Header.Set("X-User-ID", runCtx.UserID.String())
	}
	if runCtx.TenantID != nil {
		req.Header.Set("X-Tenant-ID", runCtx.TenantID.String())
	}
}

func (t *ProjectTool) applyProjectAuth(req *http.Request) {
	integration.ApplyHeaders(req.Header, t.project.RequestHeaders)
	if t.integration != nil {
		integration.ApplyHeaders(req.Header, t.integration.RequestHeaders)
	}
	authType, authConfig := t.resolveAuth()
	integration.ApplyAuth(req, integration.AuthConfig{Type: authType, Config: authConfig})
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
