-- 000005_fix_rls_casts.up.sql
-- 重建 RLS 辅助函数和策略，确保 text/uuid 类型转换明确，避免部分环境下出现
-- "operator does not exist: text = uuid" 错误。

CREATE OR REPLACE FUNCTION public.app_current_user_id()
RETURNS uuid AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_user_id', true), '')::uuid;
EXCEPTION WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

CREATE OR REPLACE FUNCTION public.app_current_tenant_id()
RETURNS uuid AS $$
BEGIN
    RETURN NULLIF(current_setting('app.current_tenant_id', true), '')::uuid;
EXCEPTION WHEN OTHERS THEN
    RETURN NULL;
END;
$$ LANGUAGE plpgsql STABLE;

DROP POLICY IF EXISTS membership_user_isolation ON memberships;
CREATE POLICY membership_user_isolation ON memberships
    FOR ALL
    USING (user_id = public.app_current_user_id()::uuid);

DROP POLICY IF EXISTS project_tenant_isolation ON projects;
CREATE POLICY project_tenant_isolation ON projects
    FOR ALL
    USING (tenant_id = public.app_current_tenant_id()::uuid);
