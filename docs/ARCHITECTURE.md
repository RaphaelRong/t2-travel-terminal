# T2 Architecture

This document describes the high-level architecture of T2 — Travel Terminal.

## Design Principles

1. **MCP / Skill / Agent First** — All capabilities are exposed through standardized interfaces.
2. **Open by Default** — Data schemas, APIs, and governance are open and community-driven.
3. **Pluggable** — New data sources and agents can be added without changing existing clients.
4. **Developer Experience** — Sandbox, debugger, documentation, and one-click publish.
5. **Security & Isolation** — OAuth2, API keys, and fine-grained scopes for multi-tenant access.

## System Layers

### 1. Presentation Layer (React + Vite)

- `apps/web/` — Web terminal dashboard
- Embedded widgets for RMS/PMS/BI
- Chat integrations (Claude, Cursor, Kimi Code, Slack)
- State management via Zustand, data fetching via TanStack Query

### 2. API & Integration Layer (Go)

- `internal/api/` — HTTP routes and handlers (Gin)
- `internal/server/` — HTTP server lifecycle and graceful shutdown
- `internal/config/` — Configuration management (Viper)
- API Gateway with key management
- MCP server registry (`pkg/mcp/`)
- OAuth2 / API Key authentication
- Rate limiting and audit logging

### 3. Agent Runtime Layer (Go)

- `internal/runtime/` — LLM / agent runtime
- Context, session, and tension detection
- Token measurement and audit reports
- Skill composition and orchestration

### 4. Data Layer (Go + External Sources)

- `internal/datastore/` — Database adapters and caching
- DerbySoft core systems (Content Suite, Property Connector, Go, Exchange, BI)
- Open global data sources (PredictHQ, Ticketmaster, Open-Meteo, etc.)
- Community-contributed MCP servers (`mcp-servers/`)
- Regional indices and vector search

## Data Flow

```
User / Agent
    │
    ▼
API Gateway ──► Auth / Scope Check
    │
    ▼
Agent Runtime ──► Skill / MCP Selection
    │
    ▼
Data Adapters ──► DerbySoft APIs + Open Data Sources
    │
    ▼
Response (index, insight, recommendation)
```

## Repository Layout

```
├── apps/
│   └── web/            # React + TypeScript + Vite frontend
├── cmd/
│   └── server/         # Go backend entry point
├── internal/           # Go internal packages
│   ├── api/            # HTTP handlers and routes
│   ├── config/         # Configuration
│   ├── datastore/      # Database/cache adapters
│   ├── mcp/            # MCP registry and management
│   ├── runtime/        # Agent / Skill runtime
│   └── server/         # HTTP server lifecycle
├── pkg/                # Go public/shared packages
│   ├── mcp/            # MCP protocol types and registry
│   ├── schemas/        # Shared data schemas
│   └── sdk/            # Official Go SDK
├── mcp-servers/        # Official MCP server implementations
├── agents/             # Reference agent implementations
├── skills/             # Reusable skill templates
├── docs/               # Documentation
├── deployments/        # Docker, K8s, and infra configs
└── .github/            # CI/CD and issue templates
```

## Security Model

- Tenant isolation at the API Gateway level
- Fine-grained scopes per MCP / Skill / Agent
- Audit logs for all tool calls and data access
- PII and commercial data handled per data partner agreements
