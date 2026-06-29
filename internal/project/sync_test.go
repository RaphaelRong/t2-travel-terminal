package project

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSyncerSyncMCPIntegration(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)

		var payload map[string]any
		require.NoError(t, json.NewDecoder(r.Body).Decode(&payload))
		assert.Equal(t, "tools/list", payload["method"])

		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"jsonrpc":"2.0","id":"tools-list","result":{"tools":[{"name":"search_events","description":"Search events","inputSchema":{"type":"object"}}]}}`))
	}))
	defer server.Close()

	caps, err := NewSyncer().SyncCapabilities(t.Context(), Project{}, Integration{
		Kind:        "mcp",
		EndpointURL: server.URL,
		AuthType:    "inherit",
	})
	require.NoError(t, err)
	require.Len(t, caps, 1)
	assert.Equal(t, "tool", caps[0].Kind)
	assert.Equal(t, "search_events", caps[0].Name)
	assert.JSONEq(t, `{"source":"mcp-tools-list"}`, string(caps[0].Metadata))
}

func TestSyncerSyncBuiltinSkillManifest(t *testing.T) {
	caps, err := NewSyncer().SyncCapabilities(t.Context(), Project{}, Integration{
		Kind:             "skill",
		DocumentationURL: "builtin:ticketmaster",
	})
	require.NoError(t, err)
	require.Len(t, caps, 1)
	assert.Equal(t, "skill", caps[0].Kind)
	assert.Equal(t, "ticketmaster_search_events", caps[0].Name)
	assert.Equal(t, "POST", caps[0].RequestMethod)
	assert.Equal(t, "/api/v1/hub/skills/ticketmaster/search-events", caps[0].RequestPath)
}
