-- 000012_hub_provider_credentials.up.sql
-- Tenant-scoped credentials for built-in third-party hub providers.

CREATE TABLE IF NOT EXISTS hub_provider_credentials (
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    provider_id text NOT NULL,
    auth_type text NOT NULL DEFAULT 'api_key' CHECK (auth_type IN ('api_key', 'bearer', 'oauth2', 'basic', 'custom')),
    auth_config jsonb NOT NULL DEFAULT '{}'::jsonb,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_by uuid REFERENCES users(id) ON DELETE SET NULL,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    PRIMARY KEY (tenant_id, provider_id)
);

ALTER TABLE hub_provider_credentials ENABLE ROW LEVEL SECURITY;
ALTER TABLE hub_provider_credentials FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS hub_provider_credentials_select_access ON hub_provider_credentials;
DROP POLICY IF EXISTS hub_provider_credentials_insert_access ON hub_provider_credentials;
DROP POLICY IF EXISTS hub_provider_credentials_update_access ON hub_provider_credentials;
DROP POLICY IF EXISTS hub_provider_credentials_delete_access ON hub_provider_credentials;

CREATE POLICY hub_provider_credentials_select_access ON hub_provider_credentials
    FOR SELECT
    USING (
        public.app_current_user_is_superadmin()
        OR tenant_id = public.app_current_tenant_id()
    );

CREATE POLICY hub_provider_credentials_insert_access ON hub_provider_credentials
    FOR INSERT
    WITH CHECK (
        public.app_current_user_is_superadmin()
        OR tenant_id = public.app_current_tenant_id()
    );

CREATE POLICY hub_provider_credentials_update_access ON hub_provider_credentials
    FOR UPDATE
    USING (
        public.app_current_user_is_superadmin()
        OR tenant_id = public.app_current_tenant_id()
    );

CREATE POLICY hub_provider_credentials_delete_access ON hub_provider_credentials
    FOR DELETE
    USING (
        public.app_current_user_is_superadmin()
        OR tenant_id = public.app_current_tenant_id()
    );
