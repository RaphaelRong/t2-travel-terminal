-- 000005_fix_rls_casts.down.sql
-- 回退 RLS 策略到原始定义

DROP POLICY IF EXISTS membership_user_isolation ON memberships;
CREATE POLICY membership_user_isolation ON memberships
    FOR ALL
    USING (user_id = public.app_current_user_id());

DROP POLICY IF EXISTS project_tenant_isolation ON projects;
CREATE POLICY project_tenant_isolation ON projects
    FOR ALL
    USING (tenant_id = public.app_current_tenant_id());
