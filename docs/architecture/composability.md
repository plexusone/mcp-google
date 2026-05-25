# Composability

One of the key design goals of Google MCP Server is **composability** - the ability to combine its skills with others to create multi-service MCP servers.

## Why Composability Matters

### Traditional Approach (Monolithic)

Without composability, you end up with many separate MCP servers:

```
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│ google-mcp-srv  │  │ slack-mcp-srv   │  │ jira-mcp-srv    │
│   9 tools       │  │   12 tools      │  │   8 tools       │
└─────────────────┘  └─────────────────┘  └─────────────────┘
         │                   │                    │
         └───────────┬───────┴────────────────────┘
                     │
              ┌──────▼──────┐
              │  MCP Client │
              │ (3 servers) │
              └─────────────┘
```

Problems:

- Multiple processes to manage
- Multiple connections for the client
- Harder configuration
- Resource overhead

### Composable Approach

With omniskill composability, you can build one unified server:

```
┌─────────────────────────────────────────────────┐
│              work-mcp-server                     │
│  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌────────┐ │
│  │ slides  │ │  docs   │ │  slack  │ │  jira  │ │
│  └─────────┘ └─────────┘ └─────────┘ └────────┘ │
│                    29 tools                      │
└─────────────────────┬───────────────────────────┘
                      │
               ┌──────▼──────┐
               │  MCP Client │
               │ (1 server)  │
               └─────────────┘
```

Benefits:

- Single process
- Single connection
- Unified configuration
- Lower resource usage

## How to Compose Skills

### Step 1: Import Skills

```go
import (
    "github.com/grokify/mcp-google/skills/slides"
    "github.com/grokify/mcp-google/skills/docs"
    // Skills from other repos
    "github.com/example/slack-mcp-skills/skills/messages"
    "github.com/example/jira-mcp-skills/skills/issues"
)
```

### Step 2: Create Authentication Clients

Each service needs its own authenticated client:

```go
// Google auth (shared between slides and docs)
googleClient, err := getGoogleHTTPClient(ctx)

// Slack auth
slackClient, err := getSlackHTTPClient(ctx)

// Jira auth
jiraClient, err := getJiraHTTPClient(ctx)
```

### Step 3: Initialize Skills

```go
slidesSkill := slides.New(googleClient)
slidesSkill.Init(ctx)

docsSkill := docs.New(googleClient)
docsSkill.Init(ctx)

slackSkill := messages.New(slackClient)
slackSkill.Init(ctx)

jiraSkill := issues.New(jiraClient)
jiraSkill.Init(ctx)
```

### Step 4: Register with Runtime

```go
rt := runtime.New(&mcp.Implementation{
    Name:    "work-mcp-server",
    Version: "v1.0.0",
}, nil)

// Register all skills
rt.RegisterSkill(slidesSkill)
rt.RegisterSkill(docsSkill)
rt.RegisterSkill(slackSkill)
rt.RegisterSkill(jiraSkill)

// Serve
rt.ServeStdio(ctx)
```

## Complete Example: Multi-Service Server

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/grokify/mcp-google/skills/slides"
    "github.com/grokify/mcp-google/skills/docs"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    runtime "github.com/plexusone/omniskill/mcp/server"
)

func main() {
    ctx := context.Background()

    // Get authenticated clients
    googleClient := getGoogleClient(ctx)

    // Initialize skills
    slidesSkill := slides.New(googleClient)
    if err := slidesSkill.Init(ctx); err != nil {
        log.Fatal(err)
    }
    defer slidesSkill.Close()

    docsSkill := docs.New(googleClient)
    if err := docsSkill.Init(ctx); err != nil {
        log.Fatal(err)
    }
    defer docsSkill.Close()

    // Create runtime
    rt := runtime.New(&mcp.Implementation{
        Name:    "work-mcp-server",
        Version: "v1.0.0",
    }, nil)

    // Register skills (use prefix to avoid conflicts)
    rt.RegisterSkillWithPrefix(slidesSkill)  // slides_get_presentation, etc.
    rt.RegisterSkillWithPrefix(docsSkill)    // docs_get_document, etc.

    // Run
    if err := rt.ServeStdio(ctx); err != nil {
        log.Fatal(err)
    }
}
```

## Best Practices

### 1. Use Prefixes When Combining

When combining multiple skills, use `RegisterSkillWithPrefix()` to avoid tool name conflicts:

```go
rt.RegisterSkillWithPrefix(slidesSkill)  // slides_*
rt.RegisterSkillWithPrefix(docsSkill)    // docs_*
```

### 2. Share Authentication Where Possible

Google Slides and Docs can share the same authenticated HTTP client:

```go
googleClient := getGoogleHTTPClient(ctx)
slidesSkill := slides.New(googleClient)
docsSkill := docs.New(googleClient)
```

### 3. Proper Cleanup

Always defer `Close()` for each skill:

```go
defer slidesSkill.Close()
defer docsSkill.Close()
```

### 4. Handle Initialization Errors

Skills may fail to initialize (network issues, invalid credentials):

```go
if err := slidesSkill.Init(ctx); err != nil {
    log.Fatalf("Failed to initialize Slides skill: %v", err)
}
```
