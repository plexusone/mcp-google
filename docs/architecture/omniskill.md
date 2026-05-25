# OmniSkill Integration

[OmniSkill](https://github.com/plexusone/omniskill) is a Go framework for building composable MCP servers. Google MCP Server uses omniskill to provide a modular, reusable architecture.

## What is OmniSkill?

OmniSkill provides:

- **Skill interface**: A standard way to package MCP tools
- **Runtime**: MCP server with multiple transport options (stdio, HTTP, SSE)
- **Registry**: Skill discovery and management
- **Library mode**: Direct invocation without MCP protocol overhead

## Skill Interface

Each Google skill implements the `skill.Skill` interface:

```go
type Skill interface {
    Name() string                    // Skill identifier
    Description() string             // Human-readable description
    Tools() []Tool                   // List of tools
    Init(ctx context.Context) error  // Initialize before use
    Close() error                    // Cleanup on shutdown
}
```

## Google Skills Implementation

### Slides Skill

```go
package slides

type Skill struct {
    httpClient    *http.Client
    presentations *gslides.PresentationsService
}

func New(httpClient *http.Client) *Skill {
    return &Skill{httpClient: httpClient}
}

func (s *Skill) Name() string { return "slides" }

func (s *Skill) Tools() []skill.Tool {
    return []skill.Tool{
        s.getPresentationTool(),
        s.listSlidesTool(),
        s.getSlideTool(),
        s.getSlideNotesTool(),
        s.getPresentationContentTool(),
    }
}

func (s *Skill) Init(ctx context.Context) error {
    svc, err := gslides.NewService(ctx, option.WithHTTPClient(s.httpClient))
    if err != nil {
        return err
    }
    s.presentations = gslides.NewPresentationsService(svc)
    return nil
}
```

### Docs Skill

```go
package docs

type Skill struct {
    httpClient *http.Client
    documents  *docs.DocumentsService
}

func New(httpClient *http.Client) *Skill {
    return &Skill{httpClient: httpClient}
}

func (s *Skill) Name() string { return "docs" }

func (s *Skill) Tools() []skill.Tool {
    return []skill.Tool{
        s.getDocumentTool(),
        s.getDocumentContentTool(),
        s.getDocumentTextTool(),
        s.getDocumentParagraphsTool(),
    }
}
```

## Tool Definition

Tools use `skill.NewTool()` for clean definitions:

```go
func (s *Skill) getDocumentTool() skill.Tool {
    return skill.NewTool(
        "get_document",
        "Get metadata about a Google Doc document",
        map[string]skill.Parameter{
            "document_id": {
                Type:        "string",
                Description: "The ID or URL of the Google Doc",
                Required:    true,
            },
        },
        func(ctx context.Context, params map[string]any) (any, error) {
            documentID := params["document_id"].(string)
            // ... implementation
            return result, nil
        },
    )
}
```

## Runtime Usage

The main server creates an omniskill Runtime and registers skills:

```go
import runtime "github.com/plexusone/omniskill/mcp/server"

// Create runtime
rt := runtime.New(&mcp.Implementation{
    Name:    "mcp-google",
    Version: "v0.4.0",
}, nil)

// Initialize and register skills
slidesSkill := slides.New(httpClient)
slidesSkill.Init(ctx)
rt.RegisterSkill(slidesSkill)

docsSkill := docs.New(httpClient)
docsSkill.Init(ctx)
rt.RegisterSkill(docsSkill)

// Serve via stdio
rt.ServeStdio(ctx)
```

## Skill Registration Options

### Without Prefix (Default)

Tools are registered with their original names:

```go
rt.RegisterSkill(slidesSkill)
// Tools: get_presentation, list_slides, get_slide, etc.
```

### With Prefix

Tools are prefixed with skill name to avoid conflicts:

```go
rt.RegisterSkillWithPrefix(slidesSkill)
// Tools: slides_get_presentation, slides_list_slides, etc.
```

This is useful when combining skills that might have tool name conflicts.

## Benefits of OmniSkill

1. **Composability**: Skills can be imported into any omniskill-based server
2. **Standardization**: Common interface for all skills
3. **Transport flexibility**: stdio, HTTP, SSE without code changes
4. **Library mode**: Call tools directly without MCP protocol
5. **Registry**: Centralized skill discovery
