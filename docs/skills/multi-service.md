# Building Multi-Service MCP Servers

This guide shows how to combine Google skills with other services to create a unified MCP server.

## Architecture

```
┌──────────────────────────────────────────────────────┐
│                 work-mcp-server                       │
│                                                       │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐  │
│  │   Google    │  │    Slack    │  │    Jira     │  │
│  │  (slides,   │  │ (messages,  │  │  (issues,   │  │
│  │   docs)     │  │  channels)  │  │  projects)  │  │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘  │
│         │                │                │         │
│         └────────┬───────┴────────┬───────┘         │
│                  │                │                 │
│         ┌────────▼────────────────▼────────┐        │
│         │       omniskill Runtime          │        │
│         └──────────────────────────────────┘        │
└──────────────────────────────────────────────────────┘
```

## Implementation Steps

### 1. Define Dependencies

```go
// go.mod
module github.com/example/work-mcp-server

require (
    github.com/grokify/mcp-google v0.4.0
    github.com/plexusone/omniskill v0.1.0
    github.com/plexusone/omnitoken v0.1.0
    // Add other skill packages
)
```

### 2. Create Skill Providers

Encapsulate skill creation with authentication:

```go
package providers

import (
    "context"
    "net/http"

    "github.com/grokify/mcp-google/skills/slides"
    "github.com/grokify/mcp-google/skills/docs"
    "github.com/plexusone/omniskill/skill"
)

type GoogleProvider struct {
    client      *http.Client
    slidesSkill *slides.Skill
    docsSkill   *docs.Skill
}

func NewGoogleProvider(client *http.Client) *GoogleProvider {
    return &GoogleProvider{client: client}
}

func (p *GoogleProvider) Init(ctx context.Context) error {
    p.slidesSkill = slides.New(p.client)
    if err := p.slidesSkill.Init(ctx); err != nil {
        return err
    }

    p.docsSkill = docs.New(p.client)
    if err := p.docsSkill.Init(ctx); err != nil {
        return err
    }

    return nil
}

func (p *GoogleProvider) Skills() []skill.Skill {
    return []skill.Skill{p.slidesSkill, p.docsSkill}
}

func (p *GoogleProvider) Close() error {
    _ = p.slidesSkill.Close()
    _ = p.docsSkill.Close()
    return nil
}
```

### 3. Configure Credentials

Use omnitoken for unified credential management:

```go
package auth

import (
    "context"
    "net/http"

    "github.com/plexusone/omnitoken"
)

type Credentials struct {
    mgr *omnitoken.TokenManager
}

func NewCredentials(vaultURI string) (*Credentials, error) {
    mgr, err := omnitoken.NewFromVaultURI(vaultURI)
    if err != nil {
        return nil, err
    }
    return &Credentials{mgr: mgr}, nil
}

func (c *Credentials) GoogleClient(ctx context.Context) (*http.Client, error) {
    return c.mgr.GetClient(ctx, "google")
}

func (c *Credentials) SlackClient(ctx context.Context) (*http.Client, error) {
    return c.mgr.GetClient(ctx, "slack")
}

func (c *Credentials) JiraClient(ctx context.Context) (*http.Client, error) {
    return c.mgr.GetClient(ctx, "jira")
}

func (c *Credentials) Close() error {
    return c.mgr.Close()
}
```

### 4. Main Server

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/example/work-mcp-server/auth"
    "github.com/example/work-mcp-server/providers"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    runtime "github.com/plexusone/omniskill/mcp/server"
)

func main() {
    ctx := context.Background()

    // Initialize credentials
    creds, err := auth.NewCredentials(os.Getenv("VAULT_URI"))
    if err != nil {
        log.Fatal(err)
    }
    defer creds.Close()

    // Create runtime
    rt := runtime.New(&mcp.Implementation{
        Name:    "work-mcp-server",
        Version: "v1.0.0",
    }, nil)

    // Initialize Google provider
    googleClient, _ := creds.GoogleClient(ctx)
    googleProvider := providers.NewGoogleProvider(googleClient)
    if err := googleProvider.Init(ctx); err != nil {
        log.Fatal(err)
    }
    defer googleProvider.Close()

    // Register Google skills with prefix
    for _, skill := range googleProvider.Skills() {
        rt.RegisterSkillWithPrefix(skill)
    }

    // Add other providers (Slack, Jira, etc.)
    // ...

    // Serve
    if err := rt.ServeStdio(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Tool Naming Strategies

### Strategy 1: Prefixed (Recommended for Multi-Service)

```go
rt.RegisterSkillWithPrefix(slidesSkill)  // slides_get_presentation
rt.RegisterSkillWithPrefix(docsSkill)    // docs_get_document
rt.RegisterSkillWithPrefix(slackSkill)   // slack_send_message
```

### Strategy 2: No Prefix (When Tool Names Don't Conflict)

```go
rt.RegisterSkill(slidesSkill)  // get_presentation
rt.RegisterSkill(docsSkill)    // get_document
rt.RegisterSkill(slackSkill)   // send_message
```

### Strategy 3: Custom Tool Names

Build wrapper skills with custom tool names:

```go
type CustomGoogleSkill struct {
    slides *slides.Skill
    docs   *docs.Skill
}

func (s *CustomGoogleSkill) Tools() []skill.Tool {
    // Create tools with custom names
    return []skill.Tool{
        skill.NewTool("google_slides_read", "...", params, handler),
        skill.NewTool("google_docs_fetch", "...", params, handler),
    }
}
```

## Configuration

### Environment Variables

```bash
# Vault URI for all credentials
export VAULT_URI=file:///path/to/secrets

# Or individual credential names
export GOOGLE_CREDS_NAME=google
export SLACK_CREDS_NAME=slack
export JIRA_CREDS_NAME=jira
```

### Claude Desktop Config

```json
{
  "mcpServers": {
    "work": {
      "command": "/path/to/work-mcp-server",
      "env": {
        "VAULT_URI": "file:///path/to/secrets"
      }
    }
  }
}
```

## Error Handling

### Graceful Degradation

If one service fails to initialize, you may want to continue with others:

```go
// Try Google
googleClient, err := creds.GoogleClient(ctx)
if err != nil {
    log.Printf("Warning: Google not available: %v", err)
} else {
    googleProvider := providers.NewGoogleProvider(googleClient)
    if err := googleProvider.Init(ctx); err != nil {
        log.Printf("Warning: Google init failed: %v", err)
    } else {
        for _, skill := range googleProvider.Skills() {
            rt.RegisterSkillWithPrefix(skill)
        }
    }
}

// Continue with Slack, Jira, etc.
```

### Health Checks

Consider adding a health check tool:

```go
func healthCheckTool(providers map[string]bool) skill.Tool {
    return skill.NewTool(
        "health_check",
        "Check which services are available",
        nil,
        func(ctx context.Context, params map[string]any) (any, error) {
            return providers, nil
        },
    )
}
```
