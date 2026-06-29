# Agent Package

T2-TravelTerminal 的用户级 AI Agent 运行时。

## 职责

- **Soul & Memory**：管理用户 Agent 人格、偏好与业务记忆。
- **Session**：用户可随时发起新会话或继续历史会话。
- **Agent Run**：类似 Hermes 的动态规划、工具调用、迭代预算、护栏。
- **Project 接入**：把 Project Capabilities（MCP / API / Skill）转换为 Agent 可用工具。
- **God 范围控制**：系统级配置限制 Agent 能做什么，防止功能外溢。
- **用户业务数据**：通过 PostgreSQL 表按用户隔离，支持上传 CSV/JSON 并做 SQL 分析。

## 包结构

```
internal/agent/
├── domain/     # 领域模型（纯结构，无外部依赖）
├── store/      # 数据访问层（PostgreSQL 实现）
├── runtime/    # Agent 运行时（待实现）
├── tools/      # Agent 工具抽象与实现（待实现）
├── workflow/   # Workflow 扩展接口（第一版仅占位）
├── userdata/   # 用户业务数据管理（待实现）
├── god/        # God 范围控制加载与校验（待实现）
├── api/        # HTTP handlers（待实现）
├── queries/    # SQL 语句常量
└── tests/      # 集成测试（待实现）
```

## 设计约束

- 核心逻辑尽量集中在 `internal/agent` 内。
- LLM 调用复用 `user_llm_profiles`。
- Workflow 引擎第一版不实现，但保留接口。
- 用户业务数据使用 PostgreSQL RLS 隔离，不引入外部临时库。
- 路由注册是唯一的必要外部接触点。

## 数据库迁移

迁移文件位于项目根目录的 `migrations/`：

- `000014_agent_core.up.sql`
- `000014_agent_core.down.sql`

运行：

```bash
make migrate-up
```

## 测试

```bash
go test ./internal/agent/...
```
