package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/tenant"
)

type llmProfileResponse struct {
	ID           string    `json:"id"`
	Provider     string    `json:"provider"`
	DisplayName  string    `json:"display_name"`
	BaseURL      string    `json:"base_url"`
	DefaultModel string    `json:"default_model"`
	Models       []string  `json:"models"`
	Status       string    `json:"status"`
	Configured   bool      `json:"configured"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type llmProfileRequest struct {
	Provider     string   `json:"provider"`
	DisplayName  string   `json:"display_name"`
	BaseURL      string   `json:"base_url"`
	APIKey       string   `json:"api_key"`
	DefaultModel string   `json:"default_model"`
	Models       []string `json:"models"`
	Status       string   `json:"status"`
}

func listLLMProfilesHandler(c *gin.Context) {
	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	rows, err := conn.Query(c.Request.Context(),
		`SELECT id, provider, display_name, COALESCE(base_url, ''), COALESCE(default_model, ''),
		        models, status, auth_config <> '{}'::jsonb AS configured, created_at, updated_at
		 FROM user_llm_profiles
		 WHERE user_id = $1
		 ORDER BY updated_at DESC, created_at DESC`,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer rows.Close()

	profiles := []llmProfileResponse{}
	for rows.Next() {
		var item llmProfileResponse
		var modelsRaw json.RawMessage
		if err := rows.Scan(
			&item.ID,
			&item.Provider,
			&item.DisplayName,
			&item.BaseURL,
			&item.DefaultModel,
			&modelsRaw,
			&item.Status,
			&item.Configured,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		item.Models = decodeModelList(modelsRaw)
		profiles = append(profiles, item)
	}
	if rows.Err() != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": rows.Err().Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"profiles": profiles})
}

func createLLMProfileHandler(c *gin.Context) {
	req, ok := bindLLMProfileRequest(c)
	if !ok {
		return
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)
	authConfig := llmAuthConfig(req.APIKey)
	modelsRaw, _ := json.Marshal(cleanModelList(req.Models))

	var item llmProfileResponse
	var models json.RawMessage
	err := conn.QueryRow(c.Request.Context(),
		`INSERT INTO user_llm_profiles (
		     user_id, provider, display_name, base_url, auth_config, default_model, models, status
		 )
		 VALUES ($1, $2, $3, NULLIF($4, ''), $5, NULLIF($6, ''), $7, $8)
		 RETURNING id, provider, display_name, COALESCE(base_url, ''), COALESCE(default_model, ''),
		           models, status, auth_config <> '{}'::jsonb AS configured, created_at, updated_at`,
		userID,
		req.Provider,
		req.DisplayName,
		req.BaseURL,
		authConfig,
		req.DefaultModel,
		json.RawMessage(modelsRaw),
		req.Status,
	).Scan(
		&item.ID,
		&item.Provider,
		&item.DisplayName,
		&item.BaseURL,
		&item.DefaultModel,
		&models,
		&item.Status,
		&item.Configured,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	item.Models = decodeModelList(models)
	c.JSON(http.StatusCreated, item)
}

func updateLLMProfileHandler(c *gin.Context) {
	profileID, err := uuid.Parse(c.Param("profile_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}
	req, ok := bindLLMProfileRequest(c)
	if !ok {
		return
	}

	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)
	modelsRaw, _ := json.Marshal(cleanModelList(req.Models))

	var item llmProfileResponse
	var models json.RawMessage
	var query string
	var args []any
	if strings.TrimSpace(req.APIKey) == "" {
		query = `UPDATE user_llm_profiles
			 SET provider = $1,
			     display_name = $2,
			     base_url = NULLIF($3, ''),
			     default_model = NULLIF($4, ''),
			     models = $5,
			     status = $6,
			     updated_at = now()
			 WHERE id = $7 AND user_id = $8
			 RETURNING id, provider, display_name, COALESCE(base_url, ''), COALESCE(default_model, ''),
			           models, status, auth_config <> '{}'::jsonb AS configured, created_at, updated_at`
		args = []any{req.Provider, req.DisplayName, req.BaseURL, req.DefaultModel, json.RawMessage(modelsRaw), req.Status, profileID, userID}
	} else {
		query = `UPDATE user_llm_profiles
			 SET provider = $1,
			     display_name = $2,
			     base_url = NULLIF($3, ''),
			     auth_config = $4,
			     default_model = NULLIF($5, ''),
			     models = $6,
			     status = $7,
			     updated_at = now()
			 WHERE id = $8 AND user_id = $9
			 RETURNING id, provider, display_name, COALESCE(base_url, ''), COALESCE(default_model, ''),
			           models, status, auth_config <> '{}'::jsonb AS configured, created_at, updated_at`
		args = []any{req.Provider, req.DisplayName, req.BaseURL, llmAuthConfig(req.APIKey), req.DefaultModel, json.RawMessage(modelsRaw), req.Status, profileID, userID}
	}

	err = conn.QueryRow(c.Request.Context(), query, args...).Scan(
		&item.ID,
		&item.Provider,
		&item.DisplayName,
		&item.BaseURL,
		&item.DefaultModel,
		&models,
		&item.Status,
		&item.Configured,
		&item.CreatedAt,
		&item.UpdatedAt,
	)
	if err == pgx.ErrNoRows {
		c.JSON(http.StatusNotFound, gin.H{"error": "llm profile not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	item.Models = decodeModelList(models)
	c.JSON(http.StatusOK, item)
}

func deleteLLMProfileHandler(c *gin.Context) {
	profileID, err := uuid.Parse(c.Param("profile_id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid profile id"})
		return
	}
	conn := tenant.ConnFromContext(c)
	userID := tenant.UserIDFromContext(c)

	tag, err := conn.Exec(c.Request.Context(),
		`DELETE FROM user_llm_profiles WHERE id = $1 AND user_id = $2`,
		profileID,
		userID,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if tag.RowsAffected() == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "llm profile not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "llm profile deleted"})
}

func bindLLMProfileRequest(c *gin.Context) (llmProfileRequest, bool) {
	var req llmProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return req, false
	}
	req.Provider = strings.TrimSpace(req.Provider)
	req.DisplayName = strings.TrimSpace(req.DisplayName)
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	req.DefaultModel = strings.TrimSpace(req.DefaultModel)
	req.Status = strings.TrimSpace(req.Status)
	if req.Provider == "" {
		req.Provider = "custom"
	}
	if req.Status == "" {
		req.Status = "active"
	}
	if req.DisplayName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "display_name is required"})
		return req, false
	}
	if !validLLMProvider(req.Provider) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return req, false
	}
	if req.Status != "active" && req.Status != "inactive" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported status"})
		return req, false
	}
	return req, true
}

func validLLMProvider(provider string) bool {
	switch provider {
	case "openai", "anthropic", "google", "custom":
		return true
	default:
		return false
	}
}

func llmAuthConfig(apiKey string) json.RawMessage {
	key := strings.TrimSpace(apiKey)
	if key == "" {
		return json.RawMessage(`{}`)
	}
	raw, _ := json.Marshal(map[string]string{"api_key": key})
	return json.RawMessage(raw)
}

func cleanModelList(models []string) []string {
	seen := map[string]bool{}
	result := []string{}
	for _, model := range models {
		name := strings.TrimSpace(model)
		if name == "" || seen[name] {
			continue
		}
		seen[name] = true
		result = append(result, name)
	}
	return result
}

func decodeModelList(raw json.RawMessage) []string {
	if len(raw) == 0 {
		return []string{}
	}
	var models []string
	if err := json.Unmarshal(raw, &models); err != nil {
		return []string{}
	}
	return cleanModelList(models)
}

type fetchLLMModelsRequest struct {
	Provider string `json:"provider"`
	BaseURL  string `json:"base_url"`
	APIKey   string `json:"api_key"`
}

type openAIModelsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
}

type googleModelsResponse struct {
	Models []struct {
		Name string `json:"name"`
	} `json:"models"`
}

func fetchLLMModelsHandler(c *gin.Context) {
	var req fetchLLMModelsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	req.Provider = strings.TrimSpace(req.Provider)
	req.BaseURL = strings.TrimSpace(req.BaseURL)
	req.APIKey = strings.TrimSpace(req.APIKey)
	if req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "api_key is required"})
		return
	}
	if req.Provider == "" {
		req.Provider = "custom"
	}
	if !validLLMProvider(req.Provider) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported provider"})
		return
	}

	models, err := fetchLLMModels(c.Request.Context(), req.Provider, req.BaseURL, req.APIKey)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"models": cleanModelList(models)})
}

func fetchLLMModels(ctx context.Context, provider, baseURL, apiKey string) ([]string, error) {
	switch provider {
	case "openai", "custom":
		return fetchOpenAICompatibleModels(ctx, baseURL, apiKey)
	case "anthropic":
		return fetchAnthropicModels(ctx, apiKey)
	case "google":
		return fetchGoogleModels(ctx, apiKey)
	default:
		return nil, fmt.Errorf("unsupported provider")
	}
}

func fetchOpenAICompatibleModels(ctx context.Context, baseURL, apiKey string) ([]string, error) {
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}
	endpoint := strings.TrimRight(baseURL, "/") + "/models"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	return doOpenAIModelsRequest(req)
}

func fetchAnthropicModels(ctx context.Context, apiKey string) ([]string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.anthropic.com/v1/models", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-api-key", apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
	return doOpenAIModelsRequest(req)
}

func doOpenAIModelsRequest(req *http.Request) ([]string, error) {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned HTTP %d: %s", resp.StatusCode, string(body))
	}
	var decoded openAIModelsResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("invalid model list response")
	}
	models := make([]string, 0, len(decoded.Data))
	for _, item := range decoded.Data {
		if item.ID != "" {
			models = append(models, item.ID)
		}
	}
	return models, nil
}

func fetchGoogleModels(ctx context.Context, apiKey string) ([]string, error) {
	endpoint := "https://generativelanguage.googleapis.com/v1beta/models?key=" + url.QueryEscape(apiKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("provider returned HTTP %d: %s", resp.StatusCode, string(body))
	}
	var decoded googleModelsResponse
	if err := json.Unmarshal(body, &decoded); err != nil {
		return nil, fmt.Errorf("invalid model list response")
	}
	models := make([]string, 0, len(decoded.Models))
	for _, item := range decoded.Models {
		name := strings.TrimPrefix(item.Name, "models/")
		if name != "" {
			models = append(models, name)
		}
	}
	return models, nil
}
