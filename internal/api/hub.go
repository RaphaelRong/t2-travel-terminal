package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/config"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/hub"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

type hubHandler struct {
	ticketmaster *hub.TicketmasterClient
}

type hubProvider struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Status       string   `json:"status"`
	AuthType     string   `json:"auth_type"`
	ManifestURL  string   `json:"manifest_url"`
	Description  string   `json:"description"`
	Capabilities []string `json:"capabilities"`
}

type providerCredentialReq struct {
	AuthType   string          `json:"auth_type"`
	AuthConfig json.RawMessage `json:"auth_config"`
	Status     string          `json:"status"`
}

func newHubHandler(cfg *config.Config) *hubHandler {
	return &hubHandler{
		ticketmaster: hub.NewTicketmasterClient(cfg.TicketmasterAPIKey),
	}
}

func builtinHubProviders() []hubProvider {
	return []hubProvider{
		{
			ID:           "ticketmaster",
			Name:         "TicketMaster Events",
			Type:         "builtin_skill",
			Status:       "available",
			AuthType:     "api_key",
			ManifestURL:  "/api/v1/hub/skills/ticketmaster/manifest",
			Description:  "Fetch city-level event data from Ticketmaster Discovery API.",
			Capabilities: []string{"ticketmaster_search_events"},
		},
	}
}

func (h *hubHandler) listProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"providers": builtinHubProviders(),
	})
}

func (h *hubHandler) listProviderCredentials(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)

	rows, err := conn.Query(c.Request.Context(),
		`SELECT provider_id, auth_type, status, updated_at
		 FROM hub_provider_credentials
		 WHERE tenant_id = $1
		 ORDER BY provider_id`,
		t.ID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	type credentialSummary struct {
		ProviderID string    `json:"provider_id"`
		AuthType   string    `json:"auth_type"`
		Status     string    `json:"status"`
		Configured bool      `json:"configured"`
		UpdatedAt  time.Time `json:"updated_at"`
	}

	result := []credentialSummary{}
	for rows.Next() {
		var item credentialSummary
		if err := rows.Scan(&item.ProviderID, &item.AuthType, &item.Status, &item.UpdatedAt); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		item.Configured = item.Status == "active"
		result = append(result, item)
	}
	c.JSON(http.StatusOK, gin.H{"credentials": result})
}

func (h *hubHandler) upsertProviderCredential(c *gin.Context) {
	providerID := c.Param("provider_id")
	if !knownHubProvider(providerID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "unknown provider"})
		return
	}

	var req providerCredentialReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.AuthType == "" {
		req.AuthType = "api_key"
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if len(req.AuthConfig) == 0 || !json.Valid(req.AuthConfig) {
		req.AuthConfig = json.RawMessage(`{}`)
	}

	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)
	userID := tenant.UserIDFromContext(c)

	_, err := conn.Exec(c.Request.Context(),
		`INSERT INTO hub_provider_credentials (
		     tenant_id, provider_id, auth_type, auth_config, status, created_by, updated_at
		 )
		 VALUES ($1, $2, $3, $4, $5, $6, now())
		 ON CONFLICT (tenant_id, provider_id)
		 DO UPDATE SET
		     auth_type = EXCLUDED.auth_type,
		     auth_config = EXCLUDED.auth_config,
		     status = EXCLUDED.status,
		     updated_at = now()`,
		t.ID, providerID, req.AuthType, req.AuthConfig, req.Status, userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "provider credential saved"})
}

func (h *hubHandler) deleteProviderCredential(c *gin.Context) {
	providerID := c.Param("provider_id")
	conn := tenant.ConnFromContext(c)
	t, _ := tenant.TenantFromContext(c)
	_, err := conn.Exec(c.Request.Context(),
		`DELETE FROM hub_provider_credentials WHERE tenant_id = $1 AND provider_id = $2`,
		t.ID, providerID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "provider credential deleted"})
}

func (h *hubHandler) ticketmasterManifest(c *gin.Context) {
	manifest := ticketmasterSkillManifestDocument()
	manifest["provider"] = gin.H{
		"id":          "ticketmaster",
		"name":        "Ticketmaster",
		"type":        "builtin",
		"auth":        "server_managed",
		"configured":  h.ticketmaster.Configured(),
		"description": "The Ticketmaster API key is configured in T2 server environment variables.",
	}
	c.JSON(http.StatusOK, manifest)
}

func (h *hubHandler) ticketmasterSearchEvents(c *gin.Context) {
	var req hub.TicketmasterSearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	apiKey, err := h.providerAPIKey(c, "ticketmaster")
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	result, err := h.ticketmaster.SearchEventsWithAPIKey(c.Request.Context(), req, apiKey)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, result)
}

func (h *hubHandler) providerAPIKey(c *gin.Context, providerID string) (string, error) {
	conn := tenant.ConnFromContext(c)
	if t, ok := tenant.TenantFromContext(c); ok {
		var authConfig json.RawMessage
		err := conn.QueryRow(c.Request.Context(),
			`SELECT auth_config
			 FROM hub_provider_credentials
			 WHERE tenant_id = $1 AND provider_id = $2 AND status = 'active'`,
			t.ID, providerID,
		).Scan(&authConfig)
		if err == nil {
			if key := apiKeyFromConfig(authConfig); key != "" {
				return key, nil
			}
		} else if err != pgx.ErrNoRows {
			return "", err
		}
	}

	if providerID == "ticketmaster" && h.ticketmaster.Configured() {
		return h.ticketmaster.DefaultAPIKey(), nil
	}
	return "", nil
}

func knownHubProvider(providerID string) bool {
	for _, provider := range builtinHubProviders() {
		if provider.ID == providerID {
			return true
		}
	}
	return false
}

func apiKeyFromConfig(raw json.RawMessage) string {
	config := map[string]string{}
	_ = json.Unmarshal(raw, &config)
	for _, key := range []string{"api_key", "key", "token"} {
		if value := strings.TrimSpace(config[key]); value != "" {
			return value
		}
	}
	return ""
}
