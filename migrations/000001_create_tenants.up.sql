-- 000001_create_tenants.up.sql
-- 多租户基础表 + RLS 策略

CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- 辅助函数：从当前会话变量读取当前用户/租户，RLS 策略使用
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

-- 租户表
CREATE TABLE IF NOT EXISTS tenants (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL,
    slug text UNIQUE,
    plan_id text DEFAULT 'free',
    status text DEFAULT 'active',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- 用户表（仅认证相关的最小字段，业务 profile 可另建表）
CREATE TABLE IF NOT EXISTS users (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email text UNIQUE NOT NULL,
    name text,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- 成员关系：用户和租户之间的多对多关系
CREATE TABLE IF NOT EXISTS memberships (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('owner', 'admin', 'member')),
    joined_at timestamptz DEFAULT now(),
    UNIQUE (tenant_id, user_id)
);
CREATE INDEX idx_memberships_user_id ON memberships(user_id);
CREATE INDEX idx_memberships_tenant_id ON memberships(tenant_id);

-- 示例业务表：projects
CREATE TABLE IF NOT EXISTS projects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id uuid NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    created_by uuid REFERENCES users(id),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);
CREATE INDEX idx_projects_tenant_id ON projects(tenant_id);
CREATE INDEX idx_projects_tenant_id_created_at ON projects(tenant_id, created_at DESC);

-- RLS：memberships 只能被自己所属用户看到
ALTER TABLE memberships ENABLE ROW LEVEL SECURITY;
ALTER TABLE memberships FORCE ROW LEVEL SECURITY;

CREATE POLICY membership_user_isolation ON memberships
    FOR ALL
    USING (user_id = public.app_current_user_id());

-- RLS：projects 只能被当前会话设置的 tenant 看到
ALTER TABLE projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE projects FORCE ROW LEVEL SECURITY;

CREATE POLICY project_tenant_isolation ON projects
    FOR ALL
    USING (tenant_id = public.app_current_tenant_id());

-- 初始化一些示例数据（可选，仅本地测试）
-- INSERT INTO users (id, email, name) VALUES (gen_random_uuid(), 'alice@example.com', 'Alice');
-- INSERT INTO tenants (id, name, slug) VALUES (gen_random_uuid(), 'Acme Corp', 'acme');
-- INSERT INTO memberships (tenant_id, user_id, role) VALUES (...);
