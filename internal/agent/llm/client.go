package llm

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/t2-travel-terminal/t2-travel-terminal/internal/agent/store"
)

// ClientFactory 根据 user_llm_profiles 创建对应的 Provider。
type ClientFactory struct{}

// NewClientFactory 创建 ClientFactory。
func NewClientFactory() *ClientFactory {
	return &ClientFactory{}
}

// LoadProfile 从数据库读取指定 profile。若 profileID 为零，则回退到该用户的第一个 active profile。
func (f *ClientFactory) LoadProfile(ctx context.Context, q store.Querier, userID, profileID uuid.UUID) (*Profile, error) {
	var profile Profile
	var authConfig map[string]interface{}
	var err error

	if profileID == uuid.Nil {
		err = q.QueryRow(ctx, `
			SELECT id, provider, COALESCE(base_url, ''), auth_config, COALESCE(default_model, '')
			FROM user_llm_profiles
			WHERE user_id = $1 AND status = 'active'
			ORDER BY updated_at DESC
			LIMIT 1
		`, userID).Scan(&profile.ID, &profile.Provider, &profile.BaseURL, &authConfig, &profile.DefaultModel)
	} else {
		err = q.QueryRow(ctx, `
			SELECT id, provider, COALESCE(base_url, ''), auth_config, COALESCE(default_model, '')
			FROM user_llm_profiles
			WHERE id = $1 AND user_id = $2
		`, profileID, userID).Scan(&profile.ID, &profile.Provider, &profile.BaseURL, &authConfig, &profile.DefaultModel)
	}
	if err != nil {
		return nil, fmt.Errorf("load llm profile: %w", err)
	}
	profile.APIKey = ExtractAPIKey(authConfig)
	return &profile, nil
}

// CreateProvider 根据 profile 创建 Provider。
func (f *ClientFactory) CreateProvider(profile *Profile) (Provider, error) {
	switch profile.Provider {
	case "openai", "custom":
		return NewOpenAIProvider(profile.BaseURL, profile.APIKey), nil
	case "anthropic":
		return NewAnthropicProvider(profile.BaseURL, profile.APIKey), nil
	default:
		return nil, fmt.Errorf("unsupported provider: %s", profile.Provider)
	}
}
