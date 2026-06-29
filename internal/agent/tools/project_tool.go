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
	schema := t.capability.InputSchema
	if schema == nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}
	return sanitizeInputSchema(schema)
}

// sanitizeInputSchema 把 Project capability 的 input_schema 清理为 LLM 兼容的 JSON Schema。
// 主要处理国产 LLM 兼容接口（如 Zhipu）不支持的字段：$ref、xml、example、file、anyOf 等。
func sanitizeInputSchema(schema map[string]interface{}) map[string]interface{} {
	result := sanitizeSchemaValue(schema)
	if result == nil {
		return map[string]interface{}{
			"type":       "object",
			"properties": map[string]interface{}{},
		}
	}
	if _, ok := result["type"]; !ok {
		result["type"] = "object"
	}
	if _, ok := result["properties"]; !ok {
		result["properties"] = map[string]interface{}{}
	}
	return result
}

func sanitizeSchemaValue(value interface{}) map[string]interface{} {
	schema, ok := value.(map[string]interface{})
	if !ok {
		return nil
	}

	// 如果包含 $ref，说明是未解析的引用，直接降级为 object，避免 LLM 接口报错。
	if ref, ok := schema["$ref"].(string); ok && ref != "" {
		return map[string]interface{}{
			"type":        "object",
			"description": "referenced object",
		}
	}

	result := make(map[string]interface{}, len(schema))
	for k, v := range schema {
		switch k {
		case "xml", "example", "examples":
			// 这些字段对 LLM 工具调用无用，且部分接口不识别，直接丢弃。
			continue
		case "anyOf", "oneOf", "allOf":
			simplified := simplifyCombinators(k, v)
			if simplified != nil {
				for sk, sv := range simplified {
					result[sk] = sv
				}
			}
			continue
		}

		switch typed := v.(type) {
		case map[string]interface{}:
			result[k] = sanitizeSchemaValue(typed)
		case []interface{}:
			result[k] = sanitizeSchemaArray(typed)
		default:
			result[k] = v
		}
	}

	// 处理不支持的 type：file 降级为 string。
	if t, ok := result["type"].(string); ok && t == "file" {
		result["type"] = "string"
		if _, ok := result["description"]; !ok {
			result["description"] = "file content or path"
		}
	}

	return result
}

func sanitizeSchemaArray(items []interface{}) []interface{} {
	result := make([]interface{}, 0, len(items))
	for _, item := range items {
		switch typed := item.(type) {
		case map[string]interface{}:
			if sanitized := sanitizeSchemaValue(typed); sanitized != nil {
				result = append(result, sanitized)
			}
		case []interface{}:
			result = append(result, sanitizeSchemaArray(typed))
		default:
			result = append(result, item)
		}
	}
	return result
}

// simplifyCombinators 简化 anyOf/oneOf/allOf。国产 LLM 兼容接口普遍不支持这些关键字。
func simplifyCombinators(key string, value interface{}) map[string]interface{} {
	options, ok := value.([]interface{})
	if !ok || len(options) == 0 {
		return nil
	}

	// 优先选择 string 类型（最常见且最通用）。
	for _, opt := range options {
		optMap, ok := opt.(map[string]interface{})
		if !ok {
			continue
		}
		if t, ok := optMap["type"].(string); ok && t == "string" {
			return sanitizeSchemaValue(optMap)
		}
	}

	// 退而求其次，选择数组类型。
	for _, opt := range options {
		optMap, ok := opt.(map[string]interface{})
		if !ok {
			continue
		}
		if t, ok := optMap["type"].(string); ok && t == "array" {
			return sanitizeSchemaValue(optMap)
		}
	}

	// 都没有则取第一个有效 schema。
	for _, opt := range options {
		if optMap, ok := opt.(map[string]interface{}); ok {
			return sanitizeSchemaValue(optMap)
		}
	}
	return nil
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
