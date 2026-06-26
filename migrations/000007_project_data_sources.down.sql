-- 000007_project_data_sources.down.sql

DROP TABLE IF EXISTS project_capabilities;

DROP POLICY IF EXISTS project_select_access ON projects;
DROP POLICY IF EXISTS project_insert_access ON projects;
DROP POLICY IF EXISTS project_update_access ON projects;
DROP POLICY IF EXISTS project_delete_access ON projects;

CREATE POLICY project_tenant_isolation ON projects
    FOR ALL
    USING (tenant_id = public.app_current_tenant_id());

DROP FUNCTION IF EXISTS public.app_current_user_is_superadmin();

ALTER TABLE projects
    DROP CONSTRAINT IF EXISTS projects_scope_tenant_check,
    DROP CONSTRAINT IF EXISTS projects_auth_type_check,
    DROP CONSTRAINT IF EXISTS projects_source_type_check,
    DROP CONSTRAINT IF EXISTS projects_status_check,
    DROP CONSTRAINT IF EXISTS projects_kind_check,
    DROP CONSTRAINT IF EXISTS projects_source_scope_check;

ALTER TABLE projects
    DROP COLUMN IF EXISTS last_published_at,
    DROP COLUMN IF EXISTS capability_summary,
    DROP COLUMN IF EXISTS auth_config,
    DROP COLUMN IF EXISTS auth_type,
    DROP COLUMN IF EXISTS request_body_template,
    DROP COLUMN IF EXISTS request_headers,
    DROP COLUMN IF EXISTS request_path,
    DROP COLUMN IF EXISTS request_method,
    DROP COLUMN IF EXISTS endpoint_url,
    DROP COLUMN IF EXISTS source_type,
    DROP COLUMN IF EXISTS status,
    DROP COLUMN IF EXISTS kind,
    DROP COLUMN IF EXISTS source_scope;
