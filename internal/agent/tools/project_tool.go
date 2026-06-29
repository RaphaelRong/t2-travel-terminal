package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/integration"
)

// ProjectTool 把 Project Capability 包装成 Agent 可调用的 Tool。
type ProjectTool struct {
	capability   *domain.ProjectCapability
	project      *domain.Project
	integration  *domain.ProjectIntegration
	localBaseURL string
	client       *http.Client
}

// NewProjectTool 创建 ProjectTool。
func NewProjectTool(capability *domain.ProjectCapability, project *domain.Project, integration *domain.ProjectIntegration, localBaseURL string) *ProjectTool {
	return &ProjectTool{
		capability:   capability,
		project:      project,
		integration:  integration,
		localBaseURL: strings.TrimRight(localBaseURL, "/"),
		client:       &http.Client{Timeout: 30 * time.Second},
	}
}

func (t *ProjectTool) Name() string        { return sanitizeToolName(t.capability.Name) }
func (t *ProjectTool) Description() string { return t.capability.Description }
func (t *ProjectTool) Domain() string      { return "project:" + string(t.capability.Kind) }
func (t *ProjectTool) InputSchema() map[string]interface{} {
	if t.capability.InputSchema != nil {
		return t.capability.InputSchema
	}
	return map[string]interface{}{"type": "object"}
}

func (t *ProjectTool) Execute(ctx context.Context, args map[string]interface{}, runCtx *Context) (interface{}, error) {
	if t.isMCPTool() {
		return t.executeMCP(ctx, args, runCtx)
	}

	method := strings.ToUpper(strings.TrimSpace(t.capability.RequestMethod))
	if method == "" {
		method = http.MethodGet
	}

	path := strings.TrimSpace(t.capability.RequestPath)
	baseURL, err := t.resolveBaseURL(path)
	if err != nil {
		return nil, err
	}

	fullURL := integration.BuildURL(baseURL, path)

	var payload any
	if method == http.MethodGet {
		fullURL = integration.AppendQueryParams(fullURL, args)
	} else {
		payload = args
	}

	req, err := integration.NewJSONRequest(ctx, method, fullURL, payload)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	// 调用本机 API（如 /api/v1/hub/...）时，需要带上用户/租户身份头
	t.applyRunHeaders(req, baseURL, runCtx)

	// 应用 headers：project 优先，integration 覆盖
	t.applyProjectAuth(req)

	resp, err := t.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute project capability: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("project capability returned HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		return map[string]interface{}{"raw": string(respBody)}, nil
	}
	return result, nil
}

func (t *ProjectTool) resolveBaseURL(path string) (string, error) {
	if t.integration != nil && t.integration.EndpointURL != "" {
		return strings.TrimRight(t.integration.EndpointURL, "/"), nil
	}
	if t.project.EndpointURL != "" {
		return strings.TrimRight(t.project.EndpointURL, "/"), nil
	}
	if strings.HasPrefix(path, "/api/v1/") && t.localBaseURL != "" {
		return t.localBaseURL, nil
	}
	return "", fmt.Errorf("no endpoint_url available for project capability %s", t.capability.Name)
}

func (t *ProjectTool) resolveAuth() (string, map[string]string) {
	if t.integration != nil && t.integration.AuthType != "" && t.integration.AuthType != "inherit" {
		return t.integration.AuthType, t.integration.AuthConfig
	}
	return t.project.AuthType, t.project.AuthConfig
}

func sanitizeToolName(name string) string {
	name = strings.TrimSpace(name)
	name = strings.ReplaceAll(name, " ", "_")
	name = strings.ReplaceAll(name, "-", "_")
	name = strings.ReplaceAll(name, ".", "_")
	name = strings.ReplaceAll(name, "/", "_")
	return name
}
