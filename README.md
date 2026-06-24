# T2 — Travel Terminal

**An open-source, open-data intelligence terminal and ecosystem for the global hotel and travel industry.**

T2 (Travel Terminal) is inspired by the idea of an airport terminal: a central hub where data, agents, and applications converge. Our mission is to democratize hospitality market intelligence by combining DerbySoft core systems with free global data sources and community-built **MCP / Skill / Agent** layers.

> 中文简介：T2（Travel Terminal，旅行终端）是一个面向全球酒店与旅游业的开源、开放数据智能终端与生态平台。我们希望让每一个酒店从业者，无论规模大小，都能免费获取可信的市场指数和 AI 决策支持。

---

## 🌍 The Four Pillars

- **Open Source 开源** — Code, schemas, and official MCP servers are published on GitHub under transparent governance.
- **Open 开放** — Open APIs, open data standards, and an open marketplace. No vendor lock-in.
- **Share 分享** — Community-contributed Agents, Skills, and MCPs; shared regional indices and insights.
- **Integrate 融入** — Embeds into Claude, Cursor, Kimi Code, Slack, Excel, BI dashboards, RMS/PMS, and enterprise systems.

---

## 🚀 What T2 Provides

- **Global Hotel Market Intelligence** — real-time regional indices, pricing signals, channel health, and event impact.
- **MCP / Skill / Agent Runtime** — standardized interfaces for AI agents and legacy systems.
- **Developer Center + Sandbox** — mock data, visual debugger, and one-click publish to Marketplace.
- **Sharing Marketplace** — a hub for third-party Agents, Skills, and MCPs with revenue sharing.

---

## 📡 Open Data Layer

T2 connects DerbySoft core systems with major free and open global data sources:

- **DerbySoft Core**: Content Suite, Property Connector, Go, Exchange, BI
- **Events**: [PredictHQ](https://www.predicthq.com/), [Ticketmaster](https://developer.ticketmaster.com/), Eventbrite, Songkick
- **Mobility & Weather**: Open-Meteo, NOAA, aviation public data, GTFS
- **Tourism & Macro**: tourism board statistics, OpenStreetMap, exchange rates, visa policies

See [docs/DATA_SOURCES.md](./docs/DATA_SOURCES.md) for the full list.

---

## 🏗️ Architecture

See [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md) for the system design.

High-level stack:

```
┌─────────────────────────────────────────────┐
│  AI Agents · Chat · BI · RMS/PMS · Investors │
├─────────────────────────────────────────────┤
│        MCP / Skill / Agent Interface        │
├─────────────────────────────────────────────┤
│     API Gateway · Sandbox · Marketplace     │
├─────────────────────────────────────────────┤
│   DerbySoft Core  +  Open Global Data       │
└─────────────────────────────────────────────┘
```

---

## 🛠️ Getting Started

> This project is in early development. The instructions below will be updated as the first MVP is released.

### Tech Stack

- **Backend**: Go 1.23+ (Gin, Viper, Zap)
- **Frontend**: React 18 + TypeScript + Vite
- **Protocol**: MCP / Skill / Agent
- **License**: AGPL-3.0

### Prerequisites

- Go ≥ 1.23
- Node.js ≥ 20
- Git
- A GitHub account

### Clone the repository

```bash
git clone https://github.com/t2-travel-terminal/t2-travel-terminal.git
cd t2-travel-terminal
```

### Run the Go backend

```bash
go run ./cmd/server
# or
make run
```

The server starts on `http://localhost:8080`.

### Run the React frontend

```bash
cd apps/web
npm install
npm run dev
```

The web app starts on `http://localhost:3000` and proxies API calls to `http://localhost:8080`.

### Build everything

```bash
make build      # builds Go binary
make web        # runs frontend dev server
```

---

## 🤝 Contributing

We welcome contributions from developers, data scientists, hoteliers, and travel industry experts.

Please read [CONTRIBUTING.md](./CONTRIBUTING.md) and [CODE_OF_CONDUCT.md](./CODE_OF_CONDUCT.md) before submitting issues or pull requests.

---

## 📄 License

T2 — Travel Terminal is licensed under the **GNU Affero General Public License v3.0 (AGPL-3.0)**.

See [LICENSE](./LICENSE) for the full text.

> Because T2 is a network service (terminal/platform), AGPL ensures that anyone running a public instance must share their modifications.

---

## 🌐 Community

- [Discussions](https://github.com/t2-travel-terminal/t2-travel-terminal/discussions)
- [Issues](https://github.com/t2-travel-terminal/t2-travel-terminal/issues)
- [Twitter / X](https://x.com/T2TravelTerminal) *(placeholder)*

---

## 🙏 Acknowledgments

T2 is initiated by [DerbySoft](https://www.derbysoft.com/) and built with the global hospitality community.
