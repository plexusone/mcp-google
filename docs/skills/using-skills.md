# Using Skills

The Google Slides and Docs skills can be imported into any omniskill-based MCP server.

## Installation

```bash
go get github.com/grokify/mcp-google@latest
```

## Import

```go
import (
    "github.com/grokify/mcp-google/skills/slides"
    "github.com/grokify/mcp-google/skills/docs"
)
```

## Skill Constructors

### Slides Skill

```go
func slides.New(httpClient *http.Client) *slides.Skill
```

Creates a new Slides skill. The HTTP client must be authenticated with Google OAuth credentials that have the `presentations.readonly` scope.

### Docs Skill

```go
func docs.New(httpClient *http.Client) *docs.Skill
```

Creates a new Docs skill. The HTTP client must be authenticated with Google OAuth credentials that have the `documents.readonly` scope.

## Lifecycle

### 1. Create

```go
slidesSkill := slides.New(httpClient)
docsSkill := docs.New(httpClient)
```

### 2. Initialize

```go
if err := slidesSkill.Init(ctx); err != nil {
    return fmt.Errorf("failed to initialize Slides skill: %w", err)
}

if err := docsSkill.Init(ctx); err != nil {
    return fmt.Errorf("failed to initialize Docs skill: %w", err)
}
```

`Init()` creates the Google API client. It must be called before registering with a runtime.

### 3. Register

```go
rt.RegisterSkill(slidesSkill)
rt.RegisterSkill(docsSkill)
```

### 4. Close

```go
defer slidesSkill.Close()
defer docsSkill.Close()
```

`Close()` is currently a no-op but should be called for forward compatibility.

## Authentication

The skills require an authenticated `*http.Client`. Options include:

### Using omnitoken

```go
import "github.com/plexusone/omnitoken"

// From service account file
mgr, _ := omnitoken.NewFromFile("/path/to/service-account.json")
client, _ := mgr.GetClient(ctx, "google")
```

### Using Google's oauth2 package

```go
import "golang.org/x/oauth2/google"

data, _ := os.ReadFile("/path/to/service-account.json")
config, _ := google.JWTConfigFromJSON(data,
    "https://www.googleapis.com/auth/presentations.readonly",
    "https://www.googleapis.com/auth/documents.readonly",
)
client := config.Client(ctx)
```

### Using goauth

```go
import "github.com/grokify/goauth"

client, _ := goauth.NewClient(ctx, "/path/to/credentials.json", "account-key")
```

## Required Scopes

Both skills require these scopes:

| Scope | Purpose |
|-------|---------|
| `presentations.readonly` | Read Slides presentations |
| `documents.readonly` | Read Docs documents |
| `drive.readonly` | Access files shared with service account |

## Tool Names

### Slides Skill (`Name() = "slides"`)

| Tool | Description |
|------|-------------|
| `get_presentation` | Get presentation metadata |
| `list_slides` | List all slides |
| `get_slide` | Get single slide content |
| `get_slide_notes` | Get speaker notes |
| `get_presentation_content` | Get all content |

### Docs Skill (`Name() = "docs"`)

| Tool | Description |
|------|-------------|
| `get_document` | Get document metadata |
| `get_document_content` | Get structured content |
| `get_document_text` | Get plain text |
| `get_document_paragraphs` | Get text by paragraphs |

## Complete Example

```go
package main

import (
    "context"
    "log"

    "github.com/grokify/mcp-google/skills/slides"
    "github.com/grokify/mcp-google/skills/docs"
    "github.com/modelcontextprotocol/go-sdk/mcp"
    "github.com/plexusone/omnitoken"
    runtime "github.com/plexusone/omniskill/mcp/server"
)

func main() {
    ctx := context.Background()

    // Get authenticated client
    mgr, err := omnitoken.NewFromFile("/path/to/service-account.json")
    if err != nil {
        log.Fatal(err)
    }
    defer mgr.Close()

    httpClient, err := mgr.GetClient(ctx, "google")
    if err != nil {
        log.Fatal(err)
    }

    // Create skills
    slidesSkill := slides.New(httpClient)
    if err := slidesSkill.Init(ctx); err != nil {
        log.Fatal(err)
    }
    defer slidesSkill.Close()

    docsSkill := docs.New(httpClient)
    if err := docsSkill.Init(ctx); err != nil {
        log.Fatal(err)
    }
    defer docsSkill.Close()

    // Create and configure runtime
    rt := runtime.New(&mcp.Implementation{
        Name:    "my-google-server",
        Version: "v1.0.0",
    }, nil)

    rt.RegisterSkill(slidesSkill)
    rt.RegisterSkill(docsSkill)

    // Run
    if err := rt.ServeStdio(ctx); err != nil {
        log.Fatal(err)
    }
}
```
