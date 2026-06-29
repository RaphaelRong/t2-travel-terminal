# Agent Package 开发计划

> 目标：在 T2-TravelTerminal 中新建 `internal/agent` package，为平台提供用户级 AI Agent 运行时。
> 文档位置：`internal/agent/DEVELOPMENT_PLAN.md`
> 状态：**Phase 4 已完成，后续为优化打磨**

---

## 1. 背景与目标

T2-TravelTerminal 已具备：

- 多租户用户体系（`users` / `tenants` / `memberships`）
- Project 及其 Capabilities（MCP / API / Skill）
- 用户级 LLM Profile（`user_llm_profiles`）
- PostgreSQL + RLS 的多租户隔离

本计划要新增的 Agent package 需支持：

1. **Soul & Memory（基于数据库）**：持久化用户的 Agent 人格、偏好、业务记忆。
2. **会话管理**：用户可随时发起新会话，或从任意历史会话继续。
3. **逻辑分析与任务拆解**：类似 Hermes 的动态规划、工具调用、迭代控制、护栏。
4. **Workflow 扩展接口**：保留未来接入 Workflow 引擎的接口，第一版不实现。
5. **Project 接入会话**：把已有 Project（MCP/Skill/API）作为会话可用工具。
6. **God 全局范围定义**：由系统级配置决定 Agent 能做什么、不能做什么，防止功能外溢。
7. **用户业务数据表**：每个用户在 PostgreSQL 中拥有隔离的数据表（通过 RLS），用于组合、实验业务数据。

---

## 1.1 已确认决策

| # | 问题 | 决策 |
|---|---|---|
| 1 | LLM 调用 | 直接复用 `user_llm_profiles`，Agent 只读取其中配置，不复刻 LLM 管理功能。 |
| 2 | God 配置 | 使用独立的 `god_configs` 系统表，不混入 `agent_souls`。 |
| 3 | Workflow 引擎 | 第一版**不实现** Workflow 引擎，但保留 `workflow/` 目录和接口，便于未来扩展。 |
| 4 | 临时数据库 | 不引入 SQLite/DuckDB 等外部临时库，统一用 PostgreSQL 中的 `agent_user_datasets` / `agent_user_dataset_rows` 存储用户业务数据。 |
| 5 | 代码边界 | 尽量只在 `internal/agent` 包内开发；向 `internal/api/router.go` 注册路由是必要但最小的外部接触点；不修改其他业务包逻辑。 |

---

## 2. 设计原则

| 原则 | 说明 |
|---|---|
| **用户隔离** | Soul、Memory、会话、用户业务数据都按 `user_id` 隔离；租户级数据走现有 RLS。 |
| **God 优先** | 所有 Agent 行为先经过 God 的范围校验，超出范围直接拒绝或要求确认。 |
| **Project 即工具** | Project Capabilities 对 Agent 来说就是 Tools，统一通过 Tool Registry 管理。 |
| **Workflow 未来扩展** | 第一版不实现 Workflow 引擎，但保留扩展接口。 |
| **包内自治** | 核心逻辑集中在 `internal/agent`；路由注册是唯一外部接触点。 |
| **会话可恢复** | 会话状态持久化，Gateway/CLI 每次请求都能水合。 |
| **渐进实现** | 先完成数据层和核心领域，再接入 LLM，最后做 TUI/前端。 |

---

## 3. 概念模型

```
┌─────────────────────────────────────────────────────────────────────┐
│                           God（全局范围定义）                          │
│  - 允许/禁止的功能域                                                  │
│  - 默认 Soul 模板                                                     │
│  - 全局工具白名单/黑名单                                               │
│  - 安全策略（如禁止执行某些命令）                                       │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         User Agent Profile                           │
│  每个用户一个，包含：                                                  │
│  - Soul（人格、沟通风格、角色定位）                                     │
│  - Memory（用户偏好、项目约定、学到的知识）                             │
│  - 默认 LLM Profile 引用                                                │
│  - 用户业务数据集引用                                                   │
└─────────────────────────────────────────────────────────────────────┘
                                    │
                ┌───────────────────┼───────────────────┐
                ▼                   ▼                   ▼
        ┌──────────────┐   ┌──────────────┐   ┌────────────────────┐
        │   Session    │   │   Memory     │   │  User Data Tables  │
        │  （会话记录）  │   │  （记忆库）   │   │  （PostgreSQL RLS）│
        └──────────────┘   └──────────────┘   └────────────────────┘
                │
                ▼
        ┌─────────────────────────────────────┐
        │  Agent Run（一次用户请求的处理）      │
        │  - Intent Analysis                  │
        │  - Planning                         │
        │  - Tool Execution                   │
        │  - Guardrails                       │
        └─────────────────────────────────────┘
```

---

## 4. 包结构建议

```
internal/agent/
├── DEVELOPMENT_PLAN.md          # 本计划
├── README.md                    # 包使用说明
├── domain/                      # 领域模型（纯结构，无外部依赖）
│   ├── soul.go                  # Soul 定义
│   ├── memory.go                # Memory 条目定义
│   ├── session.go               # Session / Message 定义
│   ├── plan.go                  # Plan / Task 定义
│   ├── run.go                   # AgentRun 状态机
│   └── god_scope.go             # God 范围定义
├── store/                       # 数据访问层
│   ├── soul_store.go            # Soul 读写
│   ├── memory_store.go          # Memory 读写
│   ├── session_store.go         # Session / Message 读写
│   ├── god_store.go             # God 配置读写
│   └── userdata_store.go        # 用户业务数据读写
├── runtime/                     # Agent 运行时
│   ├── agent.go                 # Agent 主结构体
│   ├── conversation.go          # 会话循环（类似 Hermes run_conversation）
│   ├── planner.go               # 任务规划器
│   ├── executor.go              # 工具执行器
│   ├── guardrails.go            # 护栏逻辑
│   └── budget.go                # 迭代预算
├── tools/                       # Agent 可用工具抽象
│   ├── registry.go              # Tool Registry
│   ├── tool.go                  # Tool 接口
│   ├── project_tool.go          # 把 Project Capability 包装为 Tool
│   ├── builtin_tools.go         # 内置工具（read_file / write_file / execute_sql 等）
│   └── mcp_client.go            # MCP 客户端
├── workflow/                    # Workflow 集成层（第一版仅保留接口，不实现）
│   ├── client.go                # Workflow 引擎客户端接口
│   └── noop_client.go           # 占位实现
├── userdata/                    # 用户业务数据管理（PostgreSQL 内）
│   ├── manager.go               # 数据集生命周期
│   └── query.go                 # 查询/写入辅助
├── god/                         # God 范围控制
│   ├── scope.go                 # 范围定义与校验
│   └── loader.go                # 从 god_configs + agent_souls 加载
├── api/                         # HTTP handlers（供 internal/api/router.go 注册）
│   ├── sessions.go              # 会话 CRUD / 继续
│   ├── messages.go              # 发送消息 / 获取消息
│   ├── soul.go                  # Soul 管理
│   ├── memory.go                # Memory 管理
│   ├── god.go                   # God 配置管理（SuperAdmin）
│   └── userdata.go              # 用户业务数据导入/查询
├── queries/                     # SQL 语句常量（与 internal/queries 风格一致）
│   └── agent.go
└── tests/                       # 单元/集成测试
```

---

## 5. 数据库设计建议

### 5.1 agent_user_profiles（每个用户一个 Agent 配置）

```sql
CREATE TABLE agent_user_profiles (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    soul_id uuid REFERENCES agent_souls(id),
    default_llm_profile_id uuid REFERENCES user_llm_profiles(id),
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'paused')),
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

ALTER TABLE agent_user_profiles ENABLE ROW LEVEL SECURITY;
CREATE POLICY agent_user_profiles_own ON agent_user_profiles
    FOR ALL USING (user_id = public.app_current_user_id());
```

### 5.2 agent_souls（Soul 模板 + 用户 Soul）

```sql
CREATE TABLE agent_souls (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    scope text NOT NULL DEFAULT 'user' CHECK (scope IN ('system', 'user')),
    user_id uuid REFERENCES users(id) ON DELETE CASCADE,  -- user scope 时必填
    name text NOT NULL,
    identity_text text NOT NULL,        -- Soul 人格描述
    voice_text text,                    -- 沟通风格
    values_text text,                   -- 价值观/约束
    allowed_domains jsonb NOT NULL DEFAULT '[]'::jsonb,  -- 允许的功能域
    forbidden_domains jsonb NOT NULL DEFAULT '[]'::jsonb, -- 禁止的功能域
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

ALTER TABLE agent_souls ENABLE ROW LEVEL SECURITY;
-- system scope 对所有用户可见；user scope 仅自己可见
CREATE POLICY agent_souls_select ON agent_souls
    FOR SELECT USING (
        scope = 'system'
        OR user_id = public.app_current_user_id()
    );
CREATE POLICY agent_souls_user_modify ON agent_souls
    FOR ALL USING (user_id = public.app_current_user_id());
```

### 5.3 god_configs（God 全局范围配置）

```sql
CREATE TABLE god_configs (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name text NOT NULL UNIQUE,          -- 如 "default"
    is_active boolean NOT NULL DEFAULT false,
    allowed_domains jsonb NOT NULL DEFAULT '[]'::jsonb,
    forbidden_domains jsonb NOT NULL DEFAULT '[]'::jsonb,
    allowed_tools jsonb NOT NULL DEFAULT '[]'::jsonb,
    forbidden_tools jsonb NOT NULL DEFAULT '[]'::jsonb,
    require_approval_tools jsonb NOT NULL DEFAULT '[]'::jsonb,
    max_iterations int NOT NULL DEFAULT 30,
    can_delegate boolean NOT NULL DEFAULT false,
    can_run_workflow boolean NOT NULL DEFAULT false,
    rules text NOT NULL DEFAULT '',     -- 自然语言规则，注入 system prompt
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

-- 仅 SuperAdmin 可管理；所有用户可读（用于 Agent 范围校验）
ALTER TABLE god_configs ENABLE ROW LEVEL SECURITY;
CREATE POLICY god_configs_select_all ON god_configs
    FOR SELECT USING (true);
CREATE POLICY god_configs_modify_by_superadmin ON god_configs
    FOR ALL USING (public.app_current_user_is_superadmin());
```

### 5.4 agent_memories（记忆库）

```sql
CREATE TABLE agent_memories (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    category text NOT NULL DEFAULT 'general' CHECK (category IN ('preference', 'project', 'fact', 'skill')),
    content text NOT NULL,
    source_session_id uuid,             -- 来源会话
    source_message_id uuid,             -- 来源消息
    confidence float NOT NULL DEFAULT 1.0,
    expires_at timestamptz,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX idx_agent_memories_user_category ON agent_memories(user_id, category);
CREATE INDEX idx_agent_memories_created_at ON agent_memories(user_id, created_at DESC);

ALTER TABLE agent_memories ENABLE ROW LEVEL SECURITY;
CREATE POLICY agent_memories_own ON agent_memories
    FOR ALL USING (user_id = public.app_current_user_id());
```

### 5.5 agent_sessions（会话）

```sql
CREATE TABLE agent_sessions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    tenant_id uuid REFERENCES tenants(id) ON DELETE SET NULL,
    title text,
    status text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'archived', 'deleted')),
    parent_session_id uuid REFERENCES agent_sessions(id),
    context_summary text,               -- 上下文压缩后的摘要
    context_summary_at timestamptz,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now()
);

CREATE INDEX idx_agent_sessions_user_status ON agent_sessions(user_id, status, updated_at DESC);

ALTER TABLE agent_sessions ENABLE ROW LEVEL SECURITY;
CREATE POLICY agent_sessions_own ON agent_sessions
    FOR ALL USING (user_id = public.app_current_user_id());
```

### 5.6 agent_messages（消息）

```sql
CREATE TABLE agent_messages (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('system', 'user', 'assistant', 'tool')),
    content text,
    tool_calls jsonb,                   -- assistant 的 tool_calls
    tool_call_id text,                  -- tool role 时对应 tool_call_id
    tool_name text,                     -- tool role 时工具名
    tool_result jsonb,                  -- tool role 时结果
    reasoning_content text,             -- 模型的 reasoning
    token_count integer,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now()
);

CREATE INDEX idx_agent_messages_session_created ON agent_messages(session_id, created_at);

ALTER TABLE agent_messages ENABLE ROW LEVEL SECURITY;
CREATE POLICY agent_messages_session_owner ON agent_messages
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM agent_sessions s
            WHERE s.id = agent_messages.session_id
              AND s.user_id = public.app_current_user_id()
        )
    );
```

### 5.7 agent_session_projects（会话关联的 Project）

```sql
CREATE TABLE agent_session_projects (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id uuid NOT NULL REFERENCES agent_sessions(id) ON DELETE CASCADE,
    project_id uuid NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    added_at timestamptz DEFAULT now(),
    UNIQUE (session_id, project_id)
);

ALTER TABLE agent_session_projects ENABLE ROW LEVEL SECURITY;
CREATE POLICY agent_session_projects_owner ON agent_session_projects
    FOR ALL USING (
        EXISTS (
            SELECT 1 FROM agent_sessions s
            WHERE s.id = agent_session_projects.session_id
              AND s.user_id = public.app_current_user_id()
        )
    );
```

### 5.8 agent_user_datasets / agent_user_dataset_rows（用户业务数据）

所有用户业务数据统一存放在主 PostgreSQL 中，通过 RLS 隔离，不引入外部临时库。

```sql
CREATE TABLE agent_user_datasets (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name text NOT NULL,                 -- 数据集名称，如 "q2_sales"
    description text,
    schema jsonb NOT NULL DEFAULT '{}'::jsonb,  -- 字段schema描述
    row_count int NOT NULL DEFAULT 0,
    source text,                        -- 来源：upload_csv, agent_generated, ...
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    UNIQUE (user_id, name)
);

CREATE INDEX idx_agent_user_datasets_user ON agent_user_datasets(user_id, updated_at DESC);

CREATE TABLE agent_user_dataset_rows (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    dataset_id uuid NOT NULL REFERENCES agent_user_datasets(id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,  -- 冗余，用于 RLS
    row_index int NOT NULL,
    data jsonb NOT NULL DEFAULT '{}'::jsonb,
    created_at timestamptz DEFAULT now(),
    updated_at timestamptz DEFAULT now(),
    UNIQUE (dataset_id, row_index)
);

CREATE INDEX idx_agent_user_dataset_rows_dataset ON agent_user_dataset_rows(dataset_id, row_index);

ALTER TABLE agent_user_datasets ENABLE ROW LEVEL SECURITY;
CREATE POLICY agent_user_datasets_own ON agent_user_datasets
    FOR ALL USING (user_id = public.app_current_user_id());

ALTER TABLE agent_user_dataset_rows ENABLE ROW LEVEL SECURITY;
CREATE POLICY agent_user_dataset_rows_own ON agent_user_dataset_rows
    FOR ALL USING (user_id = public.app_current_user_id());
```

设计说明：

- 行数据以 `jsonb` 存储，保持灵活性，便于用户上传不同结构的 CSV/JSON。
- Agent 通过 `execute_sql` 工具对数据集做只读或受控写入查询。
- 写入操作受 GodScope 限制，避免 Agent 误删用户数据。

---

## 6. 核心领域对象

### 6.1 Soul

```go
type Soul struct {
    ID              uuid.UUID
    Scope           string       // "system" | "user"
    UserID          *uuid.UUID    // user scope 时必填
    Name            string
    IdentityText    string
    VoiceText       string
    ValuesText      string
    AllowedDomains  []string
    ForbiddenDomains []string
    Metadata        map[string]any
}
```

### 6.2 UserProfile

```go
type UserProfile struct {
    ID               uuid.UUID
    UserID           uuid.UUID
    SoulID           *uuid.UUID
    DefaultLLMProfileID *uuid.UUID
    Status           string
    CreatedAt        time.Time
    UpdatedAt        time.Time
}
```

### 6.3 Memory

```go
type Memory struct {
    ID        uuid.UUID
    UserID    uuid.UUID
    Category  string   // "preference" | "project" | "fact" | "skill"
    Content   string
    SourceSessionID *uuid.UUID
    SourceMessageID *uuid.UUID
    Confidence float64
    ExpiresAt  *time.Time
    Metadata   map[string]any
}
```

### 6.4 Session

```go
type Session struct {
    ID                uuid.UUID
    UserID            uuid.UUID
    TenantID          *uuid.UUID
    Title             string
    Status            string
    ParentSessionID   *uuid.UUID
    ContextSummary    string
    ContextSummaryAt  *time.Time
    Metadata          map[string]any
}
```

### 6.5 Message

```go
type Message struct {
    ID              uuid.UUID
    SessionID       uuid.UUID
    Role            string
    Content         string
    ToolCalls       []ToolCall
    ToolCallID      string
    ToolName        string
    ToolResult      map[string]any
    ReasoningContent string
    TokenCount      int
}
```

### 6.6 AgentRun

```go
type AgentRun struct {
    SessionID       uuid.UUID
    UserID          uuid.UUID
    Profile         *UserProfile
    Soul            *Soul
    GodScope        *GodScope
    Tools           []Tool
    Messages        []Message
    Budget          *IterationBudget
    Guardrails      *GuardrailController
    State           string   // "planning" | "executing" | "waiting" | "completed" | "error"
}
```

### 6.7 GodConfig

```go
type GodConfig struct {
    ID                   uuid.UUID
    Name                 string
    IsActive             bool
    AllowedDomains       []string
    ForbiddenDomains     []string
    AllowedTools         []string
    ForbiddenTools       []string
    RequireApprovalTools []string
    MaxIterations        int
    CanDelegate          bool
    CanRunWorkflow       bool
    Rules                string
    CreatedAt            time.Time
    UpdatedAt            time.Time
}
```

### 6.8 UserDataset / UserDatasetRow

```go
type UserDataset struct {
    ID          uuid.UUID
    UserID      uuid.UUID
    Name        string
    Description string
    Schema      map[string]any
    RowCount    int
    Source      string
    Metadata    map[string]any
}

type UserDatasetRow struct {
    ID        uuid.UUID
    DatasetID uuid.UUID
    UserID    uuid.UUID
    RowIndex  int
    Data      map[string]any
}
```

---

## 7. God 范围控制设计

### 7.1 GodScope 定义

```go
type GodScope struct {
    AllowedDomains   []string          // 显式允许的功能域，空则全部允许
    ForbiddenDomains []string          // 显式禁止的功能域
    AllowedTools     []string          // 工具白名单
    ForbiddenTools   []string          // 工具黑名单
    RequireApproval  []string          // 需要用户确认的操作
    MaxIterations    int               // 最大迭代次数
    CanDelegate      bool              // 是否允许子代理委托
    CanRunWorkflow   bool              // 是否允许调用 Workflow
    Rules            []string          // 自然语言规则（注入 system prompt）
}
```

### 7.2 加载优先级

GodScope 按以下优先级合并：

1. **God 全局配置**（`god_configs` 中 `is_active=true` 的记录）
2. **系统默认 Soul**（`agent_souls.scope='system'`）
3. **用户 Soul**（`agent_souls.scope='user'`）
4. **会话级临时覆盖**（如用户显式禁用某工具）

校验顺序：**禁止优先于允许**。只要任一层面禁止，即禁止。

> 注：God 配置本身不允许用户修改，仅 SuperAdmin 通过管理接口维护。`loader.go` 负责读取 `god_configs` 并与 Soul 合并为最终的 `GodScope`。

### 7.3 运行时校验

每次 Agent 要调用工具前：

```go
func (s *GodScope) CanExecute(toolName string, domain string, args map[string]any) error {
    if s.isForbiddenTool(toolName) {
        return fmt.Errorf("tool %s is forbidden by God scope", toolName)
    }
    if s.isForbiddenDomain(domain) {
        return fmt.Errorf("domain %s is forbidden by God scope", domain)
    }
    if s.requiresApproval(toolName) {
        return ErrRequiresApproval
    }
    return nil
}
```

---

## 8. 会话生命周期

### 8.1 创建新会话

```
POST /api/v1/agent/sessions
{
  "title": "Q2 销售分析",
  "project_ids": ["uuid-1", "uuid-2"],
  "soul_id": "uuid-optional",
  "llm_profile_id": "uuid-optional"
}
```

流程：

1. 创建 `agent_sessions` 记录。
2. 关联 `agent_session_projects`。
3. 生成 system message，包含 Soul + God rules + Project tools schema。
4. 返回 `session_id`。

### 8.2 继续会话

```
GET /api/v1/agent/sessions/{session_id}/messages
POST /api/v1/agent/sessions/{session_id}/messages
```

流程：

1. 校验会话归属当前用户。
2. 拉取历史消息（若过长则触发上下文压缩）。
3. 水合 `AgentRun`。
4. 处理用户新消息。

### 8.3 会话列表与归档

```
GET /api/v1/agent/sessions?status=active&limit=20
PUT /api/v1/agent/sessions/{session_id}/archive
DELETE /api/v1/agent/sessions/{session_id}
```

---

## 9. Agent 运行循环

### 9.1 单次请求处理流程

```
用户消息
    │
    ▼
┌─────────────────┐
│ 1. 范围校验      │  GodScope 初筛
└────────┬────────┘
         ▼
┌─────────────────┐
│ 2. 意图分析      │  调用 LLM 判断任务类型、是否需要澄清
└────────┬────────┘
         ▼
┌─────────────────┐
│ 3. 记忆预取      │  MemoryStore.search(user_id, query)
└────────┬────────┘
         ▼
┌─────────────────┐
│ 4. 规划          │  生成 Plan（可选 todo、delegate）
└────────┬────────┘
         ▼
┌─────────────────┐
│ 5. 执行循环      │  while budget:
│                 │    - LLM 生成 tool_calls / 回复
│                 │    - 校验 GodScope
│                 │    - 执行 Tool
│                 │    - 护栏检查
│                 │    - 追加结果
└────────┬────────┘
         ▼
┌─────────────────┐
│ 6. 记忆沉淀      │  自动/半自动提取值得记忆的内容
└────────┬────────┘
         ▼
┌─────────────────┐
│ 7. 返回结果      │  文本 / 工具结果 / 工作流句柄
└─────────────────┘
```

### 9.2 迭代预算与护栏

类似 Hermes：

```go
type IterationBudget struct {
    MaxTotal int
    Used     int
}

type GuardrailController struct {
    ExactFailureCounts map[ToolSignature]int
    SameToolFailureCounts map[string]int
    NoProgressCounts    map[ToolSignature]int
}
```

---

## 10. Tool 集成

### 10.1 Tool 接口

```go
type Tool interface {
    Name() string
    Description() string
    InputSchema() map[string]any
    Domain() string
    Execute(ctx context.Context, args map[string]any, runCtx *RunContext) (any, error)
}
```

### 10.2 Project Capability → Tool

对于每个 session 关联的 Project：

1. 读取 Project 的 Capabilities。
2. 根据 `kind`（api/tool/skill）生成对应 Tool。
3. `Execute` 时根据 capability 的 `request_method` / `request_path` 发起 HTTP 调用。
4. MCP 类型通过 `mcp_client.go` 调用 `tools/call`。

### 10.3 Workflow 工具（未来扩展）

第一版不实现 Workflow 引擎，但 `workflow/` 目录保留接口：

```go
type WorkflowClient interface {
    Start(ctx context.Context, workflowName string, inputs map[string]any) (uuid.UUID, error)
    Wait(ctx context.Context, runID uuid.UUID, timeout time.Duration) (map[string]any, error)
    Status(ctx context.Context, runID uuid.UUID) (string, error)
}
```

当前使用 `noop_client.go` 返回“未启用”错误，避免 Agent 循环中尝试调用不存在的 Workflow。

### 10.4 Tool Registry

每个 `AgentRun` 实例拥有独立的 Tool Registry：

```go
type Registry struct {
    tools map[string]Tool
    god   *GodScope
}

func (r *Registry) Register(t Tool) error {
    if r.god.isForbiddenTool(t.Name()) {
        return fmt.Errorf("forbidden by God")
    }
    r.tools[t.Name()] = t
    return nil
}
```

---

## 11. 用户业务数据（PostgreSQL 内）

### 11.1 用途

- 用户上传 CSV/Excel 后做临时分析。
- Agent 生成中间表、视图、实验数据。
- 组合多个 Project 的数据做联合查询。

### 11.2 表结构

见 `5.8 agent_user_datasets / agent_user_dataset_rows`。

### 11.3 Manager 实现

```go
type UserDataManager struct {
    store userdata.Store
}

func (m *UserDataManager) CreateDataset(ctx context.Context, userID uuid.UUID, name, description string, schema map[string]any) (*UserDataset, error)
func (m *UserDataManager) ImportCSV(ctx context.Context, datasetID uuid.UUID, r io.Reader) error
func (m *UserDataManager) Query(ctx context.Context, userID uuid.UUID, sql string, args ...any) ([]map[string]any, error)
```

### 11.4 安全

- 数据集按 `user_id` 隔离，RLS 策略保证不可跨用户访问。
- Agent 通过 `execute_sql` 工具访问，SQL 先做只读/白名单校验，再执行。
- 写入类操作（INSERT/UPDATE/DELETE）受 GodScope 限制。
- 定期清理（如 30 天未访问自动归档）。

---

## 12. API 路由建议

在 `internal/api/router.go` 中新增：

```go
agentHandler := newAgentHandler(pool, logger)
agentGroup := api.Group("/agent")
agentGroup.Use(tenant.Middleware(pool, logger, tm))
{
    // Soul
    agentGroup.GET("/soul", agentHandler.getSoul)
    agentGroup.PUT("/soul", agentHandler.updateSoul)

    // Memory
    agentGroup.GET("/memory", agentHandler.listMemory)
    agentGroup.POST("/memory", agentHandler.createMemory)
    agentGroup.DELETE("/memory/:memory_id", agentHandler.deleteMemory)

    // Sessions
    agentGroup.GET("/sessions", agentHandler.listSessions)
    agentGroup.POST("/sessions", agentHandler.createSession)
    agentGroup.GET("/sessions/:session_id", agentHandler.getSession)
    agentGroup.PUT("/sessions/:session_id/archive", agentHandler.archiveSession)
    agentGroup.DELETE("/sessions/:session_id", agentHandler.deleteSession)

    // Messages / Chat
    agentGroup.GET("/sessions/:session_id/messages", agentHandler.listMessages)
    agentGroup.POST("/sessions/:session_id/messages", agentHandler.sendMessage)
    agentGroup.POST("/sessions/:session_id/interrupt", agentHandler.interruptRun)

    // Projects in session
    agentGroup.POST("/sessions/:session_id/projects", agentHandler.attachProject)
    agentGroup.DELETE("/sessions/:session_id/projects/:project_id", agentHandler.detachProject)

    // User data
    agentGroup.GET("/userdata/datasets", agentHandler.listUserDatasets)
    agentGroup.POST("/userdata/datasets", agentHandler.createUserDataset)
    agentGroup.POST("/userdata/datasets/:dataset_id/import-csv", agentHandler.importUserDatasetCSV)
    agentGroup.POST("/userdata/query", agentHandler.executeUserDataQuery)

    // God config (SuperAdmin only)
    agentGroup.GET("/god/config", requireSuperAdmin(), agentHandler.getGodConfig)
    agentGroup.PUT("/god/config", requireSuperAdmin(), agentHandler.updateGodConfig)
}
```

---

## 13. 实现阶段计划

### Phase 0：基础准备（1 周）

- [x] 创建 `internal/agent` 目录结构。
- [x] 编写数据库迁移文件（`migrations/000014_agent_core.up.sql`）。
  - 说明：迁移文件放在项目已有的 `migrations/` 目录，以便 `cmd/initdb` 自动加载；SQL 内容仅创建 agent 相关新表，不修改现有表结构。
- [x] 实现 `domain` 包中的核心结构体。
- [x] 实现 `store` 包中的基础 CRUD。
- [x] 补充单元测试。

### Phase 1：Soul + Memory + Session（1-2 周）

- [x] Soul CRUD API。
- [x] Memory CRUD API + 简单语义搜索（先用 LIKE/FTS，后续接向量）。
- [x] Session 创建、继续、列表、归档 API。
- [x] Message 存储与水合。
- [ ] 前端会话列表/聊天界面（最小可用）。

### Phase 2：God 范围控制（1 周）

- [x] GodScope 数据模型与加载逻辑。
- [x] 工具白名单/黑名单校验。
- [x] 需要审批的操作提示。
- [x] System prompt 中注入 God rules。

### Phase 3：Agent 运行循环（2-3 周）

- [x] 接入 LLM（OpenAI-compatible + Anthropic，复用 `user_llm_profiles`）。
- [x] 实现 IterationBudget。
- [x] 实现 Tool Registry。
- [x] 实现基础工具：memory_tool、execute_sql（针对用户数据集）、done。
- [x] 实现 Guardrails（重复调用检测、无进展检测）。
- [x] 实现 Plan / Todo 数据结构（先不强制使用）。
- [x] SSE 流式输出。

### Phase 4：Project 接入（1-2 周）

- [x] Project Capability → Tool 转换。
- [x] MCP client 实现（通过 integration 同步为 capability 后统一走 HTTP 执行）。
- [x] Session 关联 Project API。
- [x] 会话中动态加载/卸载 Project tools。

### Phase 5：Workflow 集成（未来扩展）

> 第一版跳过。保留 `workflow/client.go` 接口，后续可接入 Temporal / Windmill / 自研引擎。

### Phase 6：优化与打磨（持续）

- [ ] 上下文压缩。
- [ ] 向量记忆检索。
- [ ] 子代理委托（delegate，受 GodScope.CanDelegate 限制）。
- [ ] 前端实时流式输出。
- [ ] 可观测性（tracing、metrics）。

---

## 14. 已确认决策 & 仍待讨论的问题

### 已确认决策

| # | 问题 | 决策 |
|---|---|---|
| 1 | LLM 调用 | 复用 `user_llm_profiles`。 |
| 2 | God 配置 | 独立 `god_configs` 表，SuperAdmin 管理。 |
| 3 | Workflow 引擎 | 第一版不实现，仅保留扩展接口。 |
| 4 | 用户业务数据 | 使用 PostgreSQL 表 `agent_user_datasets` / `agent_user_dataset_rows`，不引入外部临时库。 |
| 5 | 代码边界 | 核心逻辑集中在 `internal/agent`，仅向 `internal/api/router.go` 注册路由。 |
| 6 | LLM Client | 每个 `AgentRun` 根据 `user_llm_profiles` 自己构造请求，不使用统一 client pool。 |
| 7 | Project 工具执行 | 后端转换执行；MCP 用 HTTP JSON-RPC，不长连接。 |
| 8 | 前端实时性 | 聊天输出使用 SSE。 |
| 9 | 记忆沉淀 | Agent 调用 memory tool；敏感记忆无需用户确认。 |
| 10 | 权限模型 | 走现有 RBAC；会话/记忆/数据按 user_id 隔离，租户成员不可见他人会话。 |
| 11 | 数据集 SQL 安全 | 允许 CREATE/DROP 临时视图；通过 RLS + SQL 白名单限制只能操作自己的数据。 |

### 仍待讨论的问题

1. **LLM Client 设计** ✅ **已确认：每个 `AgentRun` 自己构造请求**
   - 原因：不同用户/会话可能使用不同的 `user_llm_profiles`，自己构造更灵活。
   - 实现：在 `runtime/conversation.go` 中根据 `AgentRun.Profile.DefaultLLMProfileID` 读取 profile，构建对应的 HTTP client 和请求。

2. **Project 工具执行** ✅ **已确认：后端转换，MCP 不走长连接**
   - MCP 基于 HTTP JSON-RPC `tools/call`，不维护 SSE 长连接。
   - Capability 执行由 `project_tool.go` 在 `internal/agent` 内部转换后调用，再统一包装返回结果。
   - 实现分层：
     - `tools/mcp_client.go` 处理 MCP JSON-RPC
     - `tools/api_client.go` 处理 REST API
     - `tools/skill_client.go` 处理 T2 Skill

3. **前端实时性** ✅ **已确认：聊天输出使用 SSE**
   - 按流式事件推送 assistant 内容、tool 结果、错误信息。

4. **记忆沉淀策略** ✅ **已确认：Agent 调用 memory tool；敏感记忆无需用户确认**
   - 由 Agent 在合适时机主动调用 `memory_tool` 写入记忆。
   - 用户可在 Memory 管理界面事后查看/删除。

5. **权限模型** ✅ **已确认：走现有 RBAC；租户成员不可见他人会话**
   - Agent 执行工具时沿用 T2 现有的角色权限体系。
   - 会话、记忆、用户数据集均按 `user_id` 隔离，RLS 策略保证不可跨用户访问。
   - 租户成员不能查看或继续其他成员的会话。

6. **用户数据集 SQL 安全** ✅ **已确认：允许 CREATE/DROP 临时视图，但只能操作自己的数据**
   - `execute_sql` 不限于 SELECT，允许 DDL（如创建临时视图/表）。
   - 通过 RLS 强制限制只能访问 `user_id = current_user_id` 的数据。
   - 建议对 SQL 做语法白名单（禁止 DROP DATABASE / DELETE FROM other_user 等危险操作）作为第二层防护。

---

## 15. 下一步建议

建议先完成 **Phase 0 + Phase 1 的最小可用集**：

- 数据库表（`migrations/000014_agent_core.up.sql`）
- Soul / Memory / Session / Message 的 CRUD
- 一个简单的“继续聊天”API（此时 LLM 只是 echo 或固定回复，不真正执行工具）

这样可以在 1-2 周内得到一个可跑通前后端的原型，便于验证数据模型和 API 设计，再逐步叠加 Agent 运行循环、God、Project、用户数据集。

---

*计划创建时间：2026-06-26*
