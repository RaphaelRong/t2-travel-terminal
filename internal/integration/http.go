package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// AuthConfig describes the small auth vocabulary supported by Project integrations.
type AuthConfig struct {
	Type   string
	Config map[string]string
}

// NewJSONRequest builds a JSON request with common headers.
func NewJSONRequest(ctx context.Context, method string, rawURL string, payload any) (*http.Request, error) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(encoded)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if payload != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

// BuildURL joins a base URL and path while preserving absolute paths.
func BuildURL(baseURL, path string) string {
	if path == "" {
		return baseURL
	}
	if strings.HasPrefix(path, "http://") || strings.HasPrefix(path, "https://") {
		return path
	}
	if strings.HasPrefix(path, "/") {
		return strings.TrimRight(baseURL, "/") + path
	}
	return strings.TrimRight(baseURL, "/") + "/" + path
}

// AppendQueryParams appends args as query parameters.
func AppendQueryParams(rawURL string, params map[string]any) string {
	if len(params) == 0 {
		return rawURL
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	q := u.Query()
	for key, value := range params {
		q.Set(key, fmt.Sprint(value))
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// ApplyHeaders applies non-empty headers to the request.
func ApplyHeaders(header http.Header, headers map[string]string) {
	for key, value := range headers {
		if strings.TrimSpace(key) != "" {
			header.Set(key, value)
		}
	}
}

// ApplyRawHeaders decodes and applies a JSON object of headers.
func ApplyRawHeaders(header http.Header, raw json.RawMessage) {
	headers := map[string]string{}
	if err := json.Unmarshal(raw, &headers); err == nil {
		ApplyHeaders(header, headers)
	}
}

// ApplyAuth applies the configured auth scheme to req.
func ApplyAuth(req *http.Request, auth AuthConfig) {
	switch auth.Type {
	case "bearer":
		if token := strings.TrimSpace(auth.Config["token"]); token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	case "api_key":
		key := strings.TrimSpace(auth.Config["key"])
		if key == "" {
			key = strings.TrimSpace(auth.Config["api_key"])
		}
		header := strings.TrimSpace(auth.Config["header"])
		if header == "" {
			header = "X-API-Key"
		}
		if key != "" {
			req.Header.Set(header, key)
		}
	case "basic":
		username := auth.Config["username"]
		password := auth.Config["password"]
		if username != "" || password != "" {
			req.SetBasicAuth(username, password)
		}
	}
}

// RawAuthConfig creates AuthConfig from a JSON object.
func RawAuthConfig(authType string, raw json.RawMessage) AuthConfig {
	config := map[string]string{}
	_ = json.Unmarshal(raw, &config)
	return AuthConfig{Type: authType, Config: config}
}
