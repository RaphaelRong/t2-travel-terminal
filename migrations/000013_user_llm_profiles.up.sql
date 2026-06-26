-- 000013_user_llm_profiles.up.sql
-- User-scoped LLM provider profiles for Playground model selection.

CREATE TABLE IF NOT EXISTS user_llm_profiles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider text NOT NULL CHECK (provider IN ('openai', 'anthropic', 'google', 'custom')),
    display_name text NOT NULL,
    base_url text,
    auth_config jsonb NOT NULL DEFAULT '{}'::jsonb,
    default_model text,
    models jsonb NOT NULL DEFAULT '[]'::jsonb,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_user_llm_profiles_user_id
    ON user_llm_profiles(user_id, updated_at DESC);

ALTER TABLE user_llm_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE user_llm_profiles FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS user_llm_profiles_select_own ON user_llm_profiles;
DROP POLICY IF EXISTS user_llm_profiles_insert_own ON user_llm_profiles;
DROP POLICY IF EXISTS user_llm_profiles_update_own ON user_llm_profiles;
DROP POLICY IF EXISTS user_llm_profiles_delete_own ON user_llm_profiles;

CREATE POLICY user_llm_profiles_select_own ON user_llm_profiles
    FOR SELECT
    USING (user_id = public.app_current_user_id());

CREATE POLICY user_llm_profiles_insert_own ON user_llm_profiles
    FOR INSERT
    WITH CHECK (user_id = public.app_current_user_id());

CREATE POLICY user_llm_profiles_update_own ON user_llm_profiles
    FOR UPDATE
    USING (user_id = public.app_current_user_id())
    WITH CHECK (user_id = public.app_current_user_id());

CREATE POLICY user_llm_profiles_delete_own ON user_llm_profiles
    FOR DELETE
    USING (user_id = public.app_current_user_id());
