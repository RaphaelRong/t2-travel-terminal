package tools

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/domain"
)

func TestProjectToolExecutesMCPToolCall(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "tools/call", payload["method"])

		params, ok := payload["params"].(map[string]any)
		require.True(t, ok)
		assert.Equal(t, "search_events", params["name"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"tool-call","result":{"content":[{"type":"text","text":"ok"}]}}`))
	}))
	defer server.Close()

	tool := NewProjectTool(
		&domain.ProjectCapability{
			Kind:         domain.ProjectCapabilityKindTool,
			Name:         "ticketmaster_search_events",
			ExternalName: "search_events",
			Metadata:     map[string]any{"source": "mcp-tools-list"},
		},
		&domain.Project{},
		&domain.ProjectIntegration{
			Kind:        "mcp",
			EndpointURL: server.URL,
			AuthType:    "inherit",
		},
		"",
	)

	result, err := tool.Execute(t.Context(), map[string]any{"city": "Tokyo"}, nil)
	require.NoError(t, err)

	resultMap, ok := result.(map[string]any)
	require.True(t, ok)
	assert.NotEmpty(t, resultMap["content"])
}
