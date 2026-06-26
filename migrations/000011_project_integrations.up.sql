-- 000011_project_integrations.up.sql
-- Project 下配置多个能力来源：API 文档源、HTTP MCP、Skill 文档源。
-- project_capabilities 作为系统从来源同步/解析后的能力清单。

CREATE TABLE IF NOT EXISTS project_integrations (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('api', 'mcp', 'skill')),
    name text NOT NULL,
    description text,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    endpoint_url text,
    documentation_url text,
    transport text NOT NULL DEFAULT 'http' CHECK (transport IN ('http')),
    auth_type text NOT NULL DEFAULT 'inherit' CHECK (auth_type IN ('inherit', 'none', 'api_key', 'bearer', 'oauth2', 'basic', 'custom')),
    request_headers jsonb NOT NULL DEFAULT '{}'::jsonb,
    auth_config jsonb NOT NULL DEFAULT '{}'::jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    last_synced_at timestamptz,
    sync_status text NOT NULL DEFAULT 'idle' CHECK (sync_status IN ('idle', 'success', 'failed')),
    sync_error text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_project_integrations_project_id
    ON project_integrations(project_id);

ALTER TABLE project_capabilities
    ADD COLUMN IF NOT EXISTS integration_id uuid REFERENCES project_integrations(id) ON DELETE CASCADE,
    ADD COLUMN IF NOT EXISTS external_name text;

CREATE INDEX IF NOT EXISTS idx_project_capabilities_integration_id
    ON project_capabilities(integration_id);

ALTER TABLE project_integrations ENABLE ROW LEVEL SECURITY;
ALTER TABLE project_integrations FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS project_integrations_select_access ON project_integrations;
DROP POLICY IF EXISTS project_integrations_insert_access ON project_integrations;
DROP POLICY IF EXISTS project_integrations_update_access ON project_integrations;
DROP POLICY IF EXISTS project_integrations_delete_access ON project_integrations;

CREATE POLICY project_integrations_select_access ON project_integrations
    FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM projects p
            WHERE p.id = project_integrations.project_id
              AND (
                public.app_current_user_is_superadmin()
                OR p.tenant_id = public.app_current_tenant_id()
                OR (p.source_scope = 'system' AND p.status = 'online')
              )
        )
    );

CREATE POLICY project_integrations_insert_access ON project_integrations
    FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM projects p
            WHERE p.id = project_integrations.project_id
              AND (
                (p.source_scope = 'tenant' AND p.tenant_id = public.app_current_tenant_id())
                OR (p.source_scope = 'system' AND public.app_current_user_is_superadmin())
              )
        )
    );

CREATE POLICY project_integrations_update_access ON project_integrations
    FOR UPDATE
    USING (
        EXISTS (
            SELECT 1 FROM projects p
            WHERE p.id = project_integrations.project_id
              AND (
                (p.source_scope = 'tenant' AND p.tenant_id = public.app_current_tenant_id())
                OR (p.source_scope = 'system' AND public.app_current_user_is_superadmin())
              )
        )
    );

CREATE POLICY project_integrations_delete_access ON project_integrations
    FOR DELETE
    USING (
        EXISTS (
            SELECT 1 FROM projects p
            WHERE p.id = project_integrations.project_id
              AND (
                (p.source_scope = 'tenant' AND p.tenant_id = public.app_current_tenant_id())
                OR (p.source_scope = 'system' AND public.app_current_user_is_superadmin())
              )
        )
    );
