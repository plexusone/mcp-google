# Installation

## Requirements

- Go 1.26+
- Google Cloud service account with Slides and Docs API access

## Install from Source

```bash
go install github.com/plexusone/mcp-google/cmd/mcp-google@latest
```

## Build from Source

```bash
git clone https://github.com/plexusone/mcp-google.git
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
go get github.com/plexusone/mcp-google@latest
```

Then import the skills:

```go
import (
    "github.com/plexusone/mcp-google/skills/slides"
    "github.com/plexusone/mcp-google/skills/docs"
)
```

See [Using Skills](../skills/using-skills.md) for details.
