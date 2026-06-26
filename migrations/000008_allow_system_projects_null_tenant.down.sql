-- 000008_allow_system_projects_null_tenant.down.sql

DELETE FROM project_capabilities
WHERE project_id IN (
    SELECT id
    FROM projects
    WHERE source_scope = 'system'
);

DELETE FROM projects
WHERE source_scope = 'system';

DELETE FROM tenants
WHERE id = '00000000-0000-0000-0000-000000000001';
