package integration

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildURL(t *testing.T) {
	assert.Equal(t, "https://example.com/api", BuildURL("https://example.com", "/api"))
	assert.Equal(t, "https://example.com/api", BuildURL("https://example.com/", "api"))
	assert.Equal(t, "https://other.test/tool", BuildURL("https://example.com", "https://other.test/tool"))
}

func TestAppendQueryParams(t *testing.T) {
	got := AppendQueryParams("https://example.com/search?existing=1", map[string]any{
		"city":  "Tokyo",
		"limit": 10,
	})
	assert.Contains(t, got, "existing=1")
	assert.Contains(t, got, "city=Tokyo")
	assert.Contains(t, got, "limit=10")
}

func TestApplyAuth(t *testing.T) {
	req, err := http.NewRequest(http.MethodGet, "https://example.com", nil)
	assert.NoError(t, err)

	ApplyAuth(req, AuthConfig{
		Type: "api_key",
		Config: map[string]string{
			"api_key": "secret",
			"header":  "X-Test-Key",
		},
	})

	assert.Equal(t, "secret", req.Header.Get("X-Test-Key"))
}
