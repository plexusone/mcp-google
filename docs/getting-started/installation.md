# Installation

## Requirements

- Go 1.24+
- Google Cloud service account with Slides and Docs API access

## Install from Source

```bash
go install github.com/grokify/mcp-google/cmd/mcp-google@latest
```

## Build from Source

```bash
git clone https://github.com/grokify/mcp-google.git
cd mcp-google
go build ./cmd/mcp-google
```

## Verify Installation

```bash
mcp-google version
# Output: mcp-google v0.4.0
```

## Using as a Library

To use the Google skills in your own omniskill-based server:

```bash
go get github.com/grokify/mcp-google@latest
```

Then import the skills:

```go
import (
    "github.com/grokify/mcp-google/skills/slides"
    "github.com/grokify/mcp-google/skills/docs"
)
```

See [Using Skills](../skills/using-skills.md) for details.
