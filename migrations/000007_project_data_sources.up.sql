-- 000007_project_data_sources.up.sql
-- 将 Project 扩展为 Agent 可使用的数据源/工具集合基础模型。

ALTER TABLE projects
    ADD COLUMN IF NOT EXISTS source_scope text NOT NULL DEFAULT 'tenant',
    ADD COLUMN IF NOT EXISTS kind text NOT NULL DEFAULT 'data_source',
    ADD COLUMN IF NOT EXISTS status text NOT NULL DEFAULT 'draft',
    ADD COLUMN IF NOT EXISTS source_type text NOT NULL DEFAULT 'api',
    ADD COLUMN IF NOT EXISTS endpoint_url text,
    ADD COLUMN IF NOT EXISTS request_method text NOT NULL DEFAULT 'GET',
    ADD COLUMN IF NOT EXISTS request_path text,
    ADD COLUMN IF NOT EXISTS request_headers jsonb NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS request_body_template jsonb NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS auth_type text NOT NULL DEFAULT 'none',
    ADD COLUMN IF NOT EXISTS auth_config jsonb NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS capability_summary text,
    ADD COLUMN IF NOT EXISTS last_published_at timestamptz;

INSERT INTO tenants (id, name, slug, status)
VALUES (
    '00000000-0000-0000-0000-000000000001',
    'T2 System',
    'system',
    'active'
)
ON CONFLICT (id) DO NOTHING;

UPDATE tenants
SET name = 'T2 System',
    slug = 'system',
    status = 'active',
    updated_at = now()
WHERE id = '00000000-0000-0000-0000-000000000001';

DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'projects_source_scope_check'
          AND conrelid = 'public.projects'::regclass
    ) THEN
        ALTER TABLE projects
            ADD CONSTRAINT projects_source_scope_check
            CHECK (source_scope IN ('system', 'tenant'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'projects_kind_check'
          AND conrelid = 'public.projects'::regclass
    ) THEN
        ALTER TABLE projects
            ADD CONSTRAINT projects_kind_check
            CHECK (kind IN ('data_source', 'agent_bundle', 'workspace'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'projects_status_check'
          AND conrelid = 'public.projects'::regclass
    ) THEN
        ALTER TABLE projects
            ADD CONSTRAINT projects_status_check
            CHECK (status IN ('draft', 'online', 'offline', 'archived'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'projects_source_type_check'
          AND conrelid = 'public.projects'::regclass
    ) THEN
        ALTER TABLE projects
            ADD CONSTRAINT projects_source_type_check
            CHECK (source_type IN ('api', 'mcp', 'skill', 'mixed'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'projects_auth_type_check'
          AND conrelid = 'public.projects'::regclass
    ) THEN
        ALTER TABLE projects
            ADD CONSTRAINT projects_auth_type_check
            CHECK (auth_type IN ('none', 'api_key', 'bearer', 'oauth2', 'basic', 'custom'));
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'projects_scope_tenant_check'
          AND conrelid = 'public.projects'::regclass
    ) THEN
        ALTER TABLE projects
            ADD CONSTRAINT projects_scope_tenant_check
            CHECK (
                (source_scope = 'system' AND tenant_id = '00000000-0000-0000-0000-000000000001'::uuid)
                OR
                (source_scope = 'tenant' AND tenant_id <> '00000000-0000-0000-0000-000000000001'::uuid)
            );
    END IF;
END $$;

CREATE TABLE IF NOT EXISTS project_capabilities (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    kind text NOT NULL CHECK (kind IN ('api', 'tool', 'skill')),
    name text NOT NULL,
    description text,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    request_method text,
    request_path text,
    input_schema jsonb NOT NULL DEFAULT '{}'::jsonb,
    output_schema jsonb NOT NULL DEFAULT '{}'::jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_project_capabilities_project_id
    ON project_capabilities(project_id);

CREATE INDEX IF NOT EXISTS idx_projects_source_scope_status
    ON projects(source_scope, status);

CREATE OR REPLACE FUNCTION public.app_current_user_is_superadmin()
RETURNS boolean AS $$
BEGIN
    RETURN EXISTS (
        SELECT 1
        FROM users
        WHERE id = public.app_current_user_id()
          AND is_superadmin = true
    );
EXCEPTION WHEN OTHERS THEN
    RETURN false;
END;
$$ LANGUAGE plpgsql STABLE;

DROP POLICY IF EXISTS project_tenant_isolation ON projects;
DROP POLICY IF EXISTS project_select_access ON projects;
DROP POLICY IF EXISTS project_insert_access ON projects;
DROP POLICY IF EXISTS project_update_access ON projects;
DROP POLICY IF EXISTS project_delete_access ON projects;

CREATE POLICY project_select_access ON projects
    FOR SELECT
    USING (
        public.app_current_user_is_superadmin()
        OR tenant_id = public.app_current_tenant_id()
        OR (source_scope = 'system' AND status = 'online')
    );

CREATE POLICY project_insert_access ON projects
    FOR INSERT
    WITH CHECK (
        (source_scope = 'tenant' AND tenant_id = public.app_current_tenant_id())
        OR (
            source_scope = 'system'
            AND tenant_id = '00000000-0000-0000-0000-000000000001'::uuid
            AND public.app_current_user_is_superadmin()
        )
    );

CREATE POLICY project_update_access ON projects
    FOR UPDATE
    USING (
        (source_scope = 'tenant' AND tenant_id = public.app_current_tenant_id())
        OR (source_scope = 'system' AND public.app_current_user_is_superadmin())
    )
    WITH CHECK (
        (source_scope = 'tenant' AND tenant_id = public.app_current_tenant_id())
        OR (
            source_scope = 'system'
            AND tenant_id = '00000000-0000-0000-0000-000000000001'::uuid
            AND public.app_current_user_is_superadmin()
        )
    );

CREATE POLICY project_delete_access ON projects
    FOR DELETE
    USING (
        (source_scope = 'tenant' AND tenant_id = public.app_current_tenant_id())
        OR (source_scope = 'system' AND public.app_current_user_is_superadmin())
    );

ALTER TABLE project_capabilities ENABLE ROW LEVEL SECURITY;
ALTER TABLE project_capabilities FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS project_capabilities_select_access ON project_capabilities;
DROP POLICY IF EXISTS project_capabilities_insert_access ON project_capabilities;
DROP POLICY IF EXISTS project_capabilities_update_access ON project_capabilities;
DROP POLICY IF EXISTS project_capabilities_delete_access ON project_capabilities;

CREATE POLICY project_capabilities_select_access ON project_capabilities
    FOR SELECT
    USING (
        EXISTS (
            SELECT 1 FROM projects p
            WHERE p.id = project_capabilities.project_id
              AND (
                public.app_current_user_is_superadmin()
                OR p.tenant_id = public.app_current_tenant_id()
                OR (p.source_scope = 'system' AND p.status = 'online')
              )
        )
    );

CREATE POLICY project_capabilities_insert_access ON project_capabilities
    FOR INSERT
    WITH CHECK (
        EXISTS (
            SELECT 1 FROM projects p
            WHERE p.id = project_capabilities.project_id
              AND (
                (p.source_scope = 'tenant' AND p.tenant_id = public.app_current_tenant_id())
                OR (p.source_scope = 'system' AND public.app_current_user_is_superadmin())
              )
        )
    );

CREATE POLICY project_capabilities_update_access ON project_capabilities
    FOR UPDATE
    USING (
        EXISTS (
            SELECT 1 FROM projects p
            WHERE p.id = project_capabilities.project_id
              AND (
                (p.source_scope = 'tenant' AND p.tenant_id = public.app_current_tenant_id())
                OR (p.source_scope = 'system' AND public.app_current_user_is_superadmin())
              )
        )
    );

CREATE POLICY project_capabilities_delete_access ON project_capabilities
    FOR DELETE
    USING (
        EXISTS (
            SELECT 1 FROM projects p
            WHERE p.id = project_capabilities.project_id
              AND (
                (p.source_scope = 'tenant' AND p.tenant_id = public.app_current_tenant_id())
                OR (p.source_scope = 'system' AND public.app_current_user_is_superadmin())
              )
        )
    );
