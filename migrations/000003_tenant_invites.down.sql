-- 000003_tenant_invites.down.sql

DROP TABLE IF EXISTS tenant_invites;

ALTER TABLE tenants
    DROP COLUMN IF EXISTS created_by;
