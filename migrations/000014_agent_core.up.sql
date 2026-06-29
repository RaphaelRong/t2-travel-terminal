-- 000014_agent_core.up.sql
-- Agent package 核心表：Soul、Memory、Session、Message、God 配置、用户业务数据

-- 1. God 全局范围配置（系统级，仅 SuperAdmin 可管理）
CREATE TABLE IF NOT EXISTS god_configs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,
    is_active boolean NOT NULL DEFAULT false,
    allowed_domains jsonb NOT NULL DEFAULT '[]'::jsonb,
    forbidden_domains jsonb NOT NULL DEFAULT '[]'::jsonb,
    allowed_tools jsonb NOT NULL DEFAULT '[]'::jsonb,
    forbidden_tools jsonb NOT NULL DEFAULT '[]'::jsonb,
    require_approval_tools jsonb NOT NULL DEFAULT '[]'::jsonb,
    max_iterations int NOT NULL DEFAULT 30,
    can_delegate boolean NOT NULL DEFAULT false,
    can_run_workflow boolean NOT NULL DEFAULT false,
    rules text NOT NULL DEFAULT '',
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

ALTER TABLE god_configs ENABLE ROW LEVEL SECURITY;
ALTER TABLE god_configs FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS god_configs_select_all ON god_configs;
DROP POLICY IF EXISTS god_configs_modify_by_superadmin ON god_configs;

CREATE POLICY god_configs_select_all ON god_configs
    FOR SELECT USING (true);
CREATE POLICY god_configs_modify_by_superadmin ON god_configs
    FOR ALL USING (public.app_current_user_is_superadmin());

-- 2. Soul 模板与用户 Soul
CREATE TABLE IF NOT EXISTS agent_souls (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope text NOT NULL DEFAULT 'user' CHECK (scope IN ('system', 'user')),
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,
    identity_text text NOT NULL,
    voice_text text,
    values_text text,
    allowed_domains jsonb NOT NULL DEFAULT '[]'::jsonb,
    forbidden_domains jsonb NOT NULL DEFAULT '[]'::jsonb,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

ALTER TABLE agent_souls ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_souls FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS agent_souls_select ON agent_souls;
DROP POLICY IF EXISTS agent_souls_user_modify ON agent_souls;

CREATE POLICY agent_souls_select ON agent_souls
    FOR SELECT USING (
        scope = 'system'
        OR user_id = public.app_current_user_id()
    );
CREATE POLICY agent_souls_user_modify ON agent_souls
    FOR ALL USING (user_id = public.app_current_user_id());

-- 3. 每个用户的 Agent 配置
CREATE TABLE IF NOT EXISTS agent_user_profiles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    soul_id uuid REFERENCES agent_souls(id),
    default_llm_profile_id uuid REFERENCES user_llm_profiles(id),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused')),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

ALTER TABLE agent_user_profiles ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_user_profiles FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS agent_user_profiles_own ON agent_user_profiles;

CREATE POLICY agent_user_profiles_own ON agent_user_profiles
    FOR ALL USING (user_id = public.app_current_user_id());

-- 4. 记忆库
CREATE TABLE IF NOT EXISTS agent_memories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category text NOT NULL DEFAULT 'general' CHECK (category IN ('preference', 'project', 'fact', 'skill')),
    content text NOT NULL,
    source_session_id uuid,
    source_message_id uuid,
    confidence float NOT NULL DEFAULT 1.0,
    expires_at timestamptz,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agent_memories_user_category ON agent_memories(user_id, category);
CREATE INDEX IF NOT EXISTS idx_agent_memories_created_at ON agent_memories(user_id, created_at DESC);

ALTER TABLE agent_memories ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_memories FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS agent_memories_own ON agent_memories;

CREATE POLICY agent_memories_own ON agent_memories
    FOR ALL USING (user_id = public.app_current_user_id());

-- 5. 会话
CREATE TABLE IF NOT EXISTS agent_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id uuid REFERENCES tenants(id) ON DELETE SET NULL,
    title text,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'deleted')),
    parent_session_id uuid REFERENCES agent_sessions(id),
    context_summary text,
    context_summary_at timestamptz,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agent_sessions_user_status ON agent_sessions(user_id, status, updated_at DESC);

ALTER TABLE agent_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_sessions FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS agent_sessions_own ON agent_sessions;

CREATE POLICY agent_sessions_own ON agent_sessions
    FOR ALL USING (user_id = public.app_current_user_id());

-- 6. 消息
CREATE TABLE IF NOT EXISTS agent_messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('system', 'user', 'assistant', 'tool')),
    content text,
    tool_calls jsonb,
    tool_call_id text,
    tool_name text,
    tool_result jsonb,
    reasoning_content text,
    token_count integer,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_agent_messages_session_created ON agent_messages(session_id, created_at);

ALTER TABLE agent_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_messages FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS agent_messages_session_owner ON agent_messages;

CREATE POLICY agent_messages_session_owner ON agent_messages
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM agent_sessions s
            WHERE s.id = agent_messages.session_id
              AND s.user_id = public.app_current_user_id()
        )
    );

-- 7. 会话关联的 Project
CREATE TABLE IF NOT EXISTS agent_session_projects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    added_at timestamptz DEFAULT now(),
    UNIQUE (session_id, project_id)
);

ALTER TABLE agent_session_projects ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_session_projects FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS agent_session_projects_owner ON agent_session_projects;

CREATE POLICY agent_session_projects_owner ON agent_session_projects
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM agent_sessions s
            WHERE s.id = agent_session_projects.session_id
              AND s.user_id = public.app_current_user_id()
        )
    );

-- 8. 用户业务数据集
CREATE TABLE IF NOT EXISTS agent_user_datasets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,
    description text,
    schema jsonb NOT NULL DEFAULT '{}'::jsonb,
    row_count int NOT NULL DEFAULT 0,
    source text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE INDEX IF NOT EXISTS idx_agent_user_datasets_user ON agent_user_datasets(user_id, updated_at DESC);

ALTER TABLE agent_user_datasets ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_user_datasets FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS agent_user_datasets_own ON agent_user_datasets;

CREATE POLICY agent_user_datasets_own ON agent_user_datasets
    FOR ALL USING (user_id = public.app_current_user_id());

-- 9. 用户业务数据集行数据
CREATE TABLE IF NOT EXISTS agent_user_dataset_rows (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id uuid NOT NULL REFERENCES agent_user_datasets(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    row_index int NOT NULL,
    data jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    UNIQUE (dataset_id, row_index)
);

CREATE INDEX IF NOT EXISTS idx_agent_user_dataset_rows_dataset ON agent_user_dataset_rows(dataset_id, row_index);

ALTER TABLE agent_user_dataset_rows ENABLE ROW LEVEL SECURITY;
ALTER TABLE agent_user_dataset_rows FORCE ROW LEVEL SECURITY;

DROP POLICY IF EXISTS agent_user_dataset_rows_own ON agent_user_dataset_rows;

CREATE POLICY agent_user_dataset_rows_own ON agent_user_dataset_rows
    FOR ALL USING (user_id = public.app_current_user_id());
