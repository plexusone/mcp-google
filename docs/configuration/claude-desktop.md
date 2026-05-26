# Claude Desktop Configuration

Configure Claude Desktop to use Google MCP Server.

## Configuration File Location

| OS | Path |
|----|------|
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |

## Basic Configuration

### With Google Service Account

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "args": ["--credentials", "/path/to/service-account.json"]
    }
  }
}
```

### With goauth

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "args": [
        "--goauth-credentials-file", "/path/to/credentials.json",
        "--goauth-credentials-account", "myaccount"
      ]
    }
  }
}
```

### With File Vault

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "args": [
        "--vault", "file:///path/to/secrets",
        "--credentials-name", "google"
      ]
    }
  }
}
```

### With 1Password

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "env": {
        "OP_SERVICE_ACCOUNT_TOKEN": "ops_...",
        "OMNITOKEN_VAULT_URI": "op://MyVault",
        "OMNITOKEN_CREDENTIALS_NAME": "google"
      }
    }
  }
}
```

### With Bitwarden

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "env": {
        "BW_ACCESS_TOKEN": "...",
        "BW_ORGANIZATION_ID": "...",
        "OMNITOKEN_VAULT_URI": "bw://org-id",
        "OMNITOKEN_CREDENTIALS_NAME": "google"
      }
    }
  }
}
```

## Using Environment Variables

### Google Service Account

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "env": {
        "GOOGLE_CREDENTIALS_FILE": "/path/to/service-account.json"
      }
    }
  }
}
```

### goauth CredentialsSet

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "env": {
        "GOAUTH_CREDENTIALS_FILE": "/path/to/credentials.json",
        "GOAUTH_CREDENTIALS_ACCOUNT": "google-slides"
      }
    }
  }
}
```

## Environment Variables Reference

| Variable | Description |
|----------|-------------|
| `GOOGLE_CREDENTIALS_FILE` | Path to Google service account JSON |
| `GOAUTH_CREDENTIALS_FILE` | Path to goauth CredentialsSet JSON |
| `GOAUTH_CREDENTIALS_ACCOUNT` | Account key in goauth file |
| `OMNITOKEN_VAULT_URI` | Vault URI (e.g., `op://MyVault`, `bw://org-id`) |
| `OMNITOKEN_CREDENTIALS_NAME` | Credential name in vault (default: `google`) |
| `OP_SERVICE_ACCOUNT_TOKEN` | 1Password service account token |
| `BW_ACCESS_TOKEN` | Bitwarden access token |
| `BW_ORGANIZATION_ID` | Bitwarden organization ID |

## Multiple Servers

You can run multiple MCP servers alongside Google:

```json
{
  "mcpServers": {
    "google": {
      "command": "/path/to/mcp-google",
      "args": ["--credentials", "/path/to/google-service-account.json"]
    },
    "github": {
      "command": "/path/to/github-mcp-server",
      "args": ["--token", "ghp_xxx"]
    },
    "filesystem": {
      "command": "/path/to/filesystem-mcp-server",
      "args": ["--root", "/home/user/documents"]
    }
  }
}
```

## Finding the Binary Path

### If installed with `go install`

```bash
# Find GOPATH
go env GOPATH

# Binary is at $GOPATH/bin/mcp-google
# Typically: ~/go/bin/mcp-google
```

### If built from source

Use the full path to where you built it:

```bash
/path/to/mcp-google/mcp-google
```

## Troubleshooting

### Server Not Starting

Check the Claude Desktop logs:

- macOS: `~/Library/Logs/Claude/`
- Windows: `%APPDATA%\Claude\logs\`

Common issues:

1. **Binary not found**: Verify the path is correct
2. **Credentials not found**: Check the credentials file path
3. **Permission denied**: Ensure the binary is executable (`chmod +x`)

### Verifying Configuration

Test the server manually:

```bash
# Should start and wait for input (Ctrl+C to exit)
/path/to/mcp-google --credentials /path/to/creds.json
```

### JSON Syntax Errors

Validate your JSON:

```bash
# On macOS/Linux
cat ~/Library/Application\ Support/Claude/claude_desktop_config.json | python3 -m json.tool
```

## Available Tools in Claude

Once configured, you can ask Claude to:

- "Read the Google Slides presentation at [URL]"
- "Get the text from this Google Doc: [URL or ID]"
- "List all slides in presentation [ID]"
- "Show me the speaker notes for slide 3"

Claude will use the appropriate tools:

- `get_presentation` / `get_presentation_content` for Slides
- `get_document_metadata` / `get_document_text` for Docs
