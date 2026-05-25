# Architecture Overview

Google MCP Server is built on a composable architecture using [omniskill](https://github.com/plexusone/omniskill), enabling its Google skills to be reused in other MCP servers.

## Design Principles

### 1. Composability First

Rather than being a monolithic MCP server, this project exposes its functionality as **skills** - modular units that can be:

- Used standalone in `mcp-google`
- Imported into other omniskill-based servers
- Combined with skills from other services (Slack, Jira, GitHub, etc.)

### 2. Separation of Concerns

```
mcp-google/
├── cmd/mcp-google/    # CLI entry point
├── skills/                   # Exportable omniskill modules
│   ├── slides/              # Google Slides skill
│   └── docs/                # Google Docs skill
└── internal/
    └── auth/                # Shared authentication
```

- **Skills**: Contain all business logic, tools, and API interactions
- **CLI**: Only handles configuration and skill orchestration
- **Auth**: Shared authentication layer using omnitoken

### 3. Skill Independence

Each skill is self-contained:

- Has its own `Init()` lifecycle
- Manages its own Google API client
- Defines its own tools
- Can be registered independently

## Component Diagram

```
┌─────────────────────────────────────────────────────┐
│                  mcp-google                   │
├─────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │   Slides    │  │    Docs     │  │    Auth     │  │
│  │   Skill     │  │   Skill     │  │  (omnitoken)│  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  │
│         │                │                │         │
│         └────────┬───────┴────────────────┘         │
│                  │                                  │
│         ┌────────▼────────┐                         │
│         │ omniskill Runtime│                         │
│         └────────┬────────┘                         │
│                  │                                  │
│         ┌────────▼────────┐                         │
│         │  MCP Protocol   │                         │
│         │  (stdio/HTTP)   │                         │
│         └─────────────────┘                         │
└─────────────────────────────────────────────────────┘
```

## Key Dependencies

| Package | Purpose |
|---------|---------|
| [omniskill](https://github.com/plexusone/omniskill) | Skill framework and MCP runtime |
| [omnitoken](https://github.com/plexusone/omnitoken) | Token management with vault backends |
| [gogoogle](https://github.com/grokify/gogoogle) | Google API utilities |
| [goauth](https://github.com/grokify/goauth) | Authentication utilities |

## Data Flow

1. **Initialization**:
   - CLI parses credentials configuration
   - omnitoken creates authenticated HTTP client
   - Skills are initialized with the HTTP client
   - Skills register with omniskill Runtime

2. **Request Handling**:
   - MCP client sends tool call request
   - omniskill Runtime routes to appropriate skill
   - Skill executes tool logic using Google APIs
   - Response returned via MCP protocol

3. **Shutdown**:
   - Skills Close() called for cleanup
   - Token manager releases resources
