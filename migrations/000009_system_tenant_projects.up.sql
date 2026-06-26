-- 000009_system_tenant_projects.up.sql
-- 确保系统 Project 使用固定 System Tenant，而不是 tenant_id = NULL。

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

DROP POLICY IF EXISTS project_insert_access ON projects;
DROP POLICY IF EXISTS project_update_access ON projects;
ALTER TABLE projects DROP CONSTRAINT IF EXISTS projects_scope_tenant_check;

UPDATE projects
SET tenant_id = '00000000-0000-0000-0000-000000000001'
WHERE source_scope = 'system'
  AND tenant_id IS NULL;

ALTER TABLE projects
    ALTER COLUMN tenant_id SET NOT NULL;

ALTER TABLE projects
    ADD CONSTRAINT projects_scope_tenant_check
    CHECK (
        (source_scope = 'system' AND tenant_id = '00000000-0000-0000-0000-000000000001'::uuid)
        OR
        (source_scope = 'tenant' AND tenant_id <> '00000000-0000-0000-0000-000000000001'::uuid)
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
