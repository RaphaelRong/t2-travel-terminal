package project

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/t2-travel-terminal/t2-travel-terminal/internal/integration"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/jsonx"
	"gopkg.in/yaml.v3"
)

// Project is the sync service's minimal view of a T2 project.
type Project struct {
	RequestHeaders json.RawMessage
	AuthType       string
	AuthConfig     json.RawMessage
}

// Integration is the sync service's minimal view of a project integration.
type Integration struct {
	Kind             string
	EndpointURL      string
	DocumentationURL string
	AuthType         string
	RequestHeaders   json.RawMessage
	AuthConfig       json.RawMessage
}

// Capability is a synchronized Project capability.
type Capability struct {
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

// Syncer turns Integration manifests into capabilities.
type Syncer struct {
	client *http.Client
}

// NewSyncer creates a sync service with conservative network timeouts.
func NewSyncer() *Syncer {
	return &Syncer{
		client: &http.Client{Timeout: 20 * time.Second},
	}
}

// SyncCapabilities loads the Integration's manifest and converts it to capabilities.
func (s *Syncer) SyncCapabilities(ctx context.Context, project Project, integration Integration) ([]Capability, error) {
	switch integration.Kind {
	case "mcp":
		return s.syncMCPIntegration(ctx, project, integration)
	case "api":
		return s.syncAPIIntegration(ctx, project, integration)
	case "skill":
		return s.syncSkillIntegration(ctx, project, integration)
	default:
		return nil, fmt.Errorf("unsupported integration kind: %s", integration.Kind)
	}
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

func (s *Syncer) syncMCPIntegration(ctx context.Context, project Project, integration Integration) ([]Capability, error) {
	if strings.TrimSpace(integration.EndpointURL) == "" {
		return nil, fmt.Errorf("mcp endpoint_url is required")
	}

	tools, err := s.fetchMCPTools(ctx, project, integration)
	if err != nil {
		return nil, err
	}

	result := make([]Capability, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if name == "" {
			continue
		}
		result = append(result, Capability{
			Kind:          "tool",
			Name:          name,
			ExternalName:  name,
			Description:   tool.Description,
			RequestMethod: "POST",
			InputSchema:   jsonx.Object(tool.InputSchema),
			OutputSchema:  json.RawMessage(`{}`),
			Metadata:      jsonx.MustMarshalObject(map[string]any{"source": "mcp-tools-list"}),
		})
	}
	return result, nil
}

func (s *Syncer) fetchMCPTools(ctx context.Context, project Project, integration Integration) ([]struct {
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
	req, err := integrationRequest(ctx, http.MethodPost, strings.TrimSpace(integration.EndpointURL), payload)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, text/event-stream")
	applyIntegrationAuth(req, project, integration)

	resp, err := s.client.Do(req)
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

func (s *Syncer) syncAPIIntegration(ctx context.Context, project Project, integration Integration) ([]Capability, error) {
	docURL := integrationURL(integration.DocumentationURL, integration.EndpointURL)
	if docURL == "" {
		return nil, fmt.Errorf("api documentation_url is required")
	}

	doc, err := s.fetchStructuredDocument(ctx, project, integration, docURL)
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

func (s *Syncer) syncSkillIntegration(ctx context.Context, project Project, integration Integration) ([]Capability, error) {
	docURL := integrationURL(integration.DocumentationURL, integration.EndpointURL)
	if docURL == "" {
		return nil, fmt.Errorf("skill documentation_url is required")
	}

	doc, err := s.fetchSkillDocument(ctx, project, integration, docURL)
	if err != nil {
		return nil, err
	}

	tools := getSlice(doc, "tools")
	if len(tools) == 0 {
		return nil, fmt.Errorf("skill manifest must include tools")
	}

	result := make([]Capability, 0, len(tools))
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
		result = append(result, Capability{
			Kind:          "skill",
			Name:          name,
			ExternalName:  name,
			Description:   firstNonEmpty(getString(tool, "description"), getString(tool, "summary")),
			RequestMethod: method,
			RequestPath:   path,
			InputSchema:   rawObjectField(tool, "input_schema", "inputSchema", "schema"),
			OutputSchema:  rawObjectField(tool, "output_schema", "outputSchema"),
			Metadata: jsonx.MustMarshalObject(map[string]any{
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

func (s *Syncer) fetchSkillDocument(ctx context.Context, project Project, integration Integration, docURL string) (map[string]any, error) {
	switch docURL {
	case "/api/v1/hub/skills/ticketmaster/manifest", "/hub/skills/ticketmaster/manifest", "builtin:ticketmaster":
		return TicketmasterSkillManifestDocument(), nil
	default:
		return s.fetchStructuredDocument(ctx, project, integration, docURL)
	}
}

// TicketmasterSkillManifestDocument returns the built-in Ticketmaster Skill manifest.
func TicketmasterSkillManifestDocument() map[string]any {
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

func (s *Syncer) fetchStructuredDocument(ctx context.Context, project Project, integration Integration, docURL string) (map[string]any, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, docURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json, application/yaml, text/yaml, */*")
	applyIntegrationAuth(req, project, integration)

	resp, err := s.client.Do(req)
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

func integrationRequest(ctx context.Context, method string, rawURL string, payload any) (*http.Request, error) {
	req, err := integration.NewJSONRequest(ctx, method, rawURL, payload)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func applyIntegrationAuth(req *http.Request, project Project, item Integration) {
	integration.ApplyRawHeaders(req.Header, project.RequestHeaders)
	if item.AuthType == "inherit" {
		integration.ApplyAuth(req, integration.RawAuthConfig(project.AuthType, project.AuthConfig))
	}
	integration.ApplyRawHeaders(req.Header, item.RequestHeaders)
	if item.AuthType != "inherit" {
		integration.ApplyAuth(req, integration.RawAuthConfig(item.AuthType, item.AuthConfig))
	}
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

func capabilitiesFromOpenAPI3(doc map[string]any) ([]Capability, error) {
	paths := getMap(doc, "paths")
	if len(paths) == 0 {
		return nil, fmt.Errorf("openapi document has no paths")
	}

	result := []Capability{}
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
			result = append(result, Capability{
				Kind:          "api",
				Name:          name,
				ExternalName:  firstNonEmpty(getString(operation, "operationId"), name),
				Description:   firstNonEmpty(getString(operation, "summary"), getString(operation, "description")),
				RequestMethod: strings.ToUpper(method),
				RequestPath:   path,
				InputSchema:   inputSchema,
				OutputSchema:  outputSchema,
				Metadata: jsonx.MustMarshalObject(map[string]any{
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

func capabilitiesFromSwagger2(doc map[string]any) ([]Capability, error) {
	paths := getMap(doc, "paths")
	if len(paths) == 0 {
		return nil, fmt.Errorf("swagger document has no paths")
	}

	basePath := getString(doc, "basePath")
	result := []Capability{}
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
			result = append(result, Capability{
				Kind:          "api",
				Name:          name,
				ExternalName:  firstNonEmpty(getString(operation, "operationId"), name),
				Description:   firstNonEmpty(getString(operation, "summary"), getString(operation, "description")),
				RequestMethod: strings.ToUpper(method),
				RequestPath:   basePath + path,
				InputSchema:   inputSchema,
				OutputSchema:  outputSchema,
				Metadata: jsonx.MustMarshalObject(map[string]any{
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
			return jsonx.MustMarshalObject(schema)
		}
		if schema := resolveSchemaRef(doc, getMap(response, "schema")); len(schema) > 0 {
			return jsonx.MustMarshalObject(schema)
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
	return jsonx.MustMarshalObject(schema)
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

func integrationURL(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
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
			return jsonx.MustMarshalObject(value)
		}
	}
	return json.RawMessage(`{}`)
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
