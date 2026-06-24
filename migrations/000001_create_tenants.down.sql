-- 000001_create_tenants.down.sql

DROP POLICY IF EXISTS project_tenant_isolation ON projects;
DROP POLICY IF EXISTS membership_user_isolation ON memberships;

ALTER TABLE IF EXISTS projects DISABLE ROW LEVEL SECURITY;
ALTER TABLE IF EXISTS memberships DISABLE ROW LEVEL SECURITY;

DROP TABLE IF EXISTS projects;
DROP TABLE IF EXISTS memberships;
DROP TABLE IF EXISTS users;
DROP TABLE IF EXISTS tenants;

DROP FUNCTION IF EXISTS public.app_current_tenant_id();
DROP FUNCTION IF EXISTS public.app_current_user_id();
