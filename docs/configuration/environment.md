# Environment Variables

All command-line flags can be set via environment variables.

## Available Variables

### Credential Configuration

| Variable | Flag | Description |
|----------|------|-------------|
| `GOOGLE_CREDENTIALS_FILE` | `--credentials` | Path to Google service account JSON |
| `GOAUTH_CREDENTIALS_FILE` | `--goauth-credentials-file` | Path to goauth CredentialsSet JSON |
| `GOAUTH_CREDENTIALS_ACCOUNT` | `--goauth-credentials-account` | Account key in CredentialsSet |
| `OMNITOKEN_VAULT_URI` | `--vault` | Vault URI for credentials |
| `OMNITOKEN_CREDENTIALS_NAME` | `--credentials-name` | Name of credentials in vault |

### Vault Provider Authentication

| Variable | Description |
|----------|-------------|
| `OP_SERVICE_ACCOUNT_TOKEN` | 1Password service account token |
| `BW_ACCESS_TOKEN` | Bitwarden access token |
| `BW_ORGANIZATION_ID` | Bitwarden organization ID |

## Precedence

Command-line flags take precedence over environment variables.

```bash
# Environment variable is used
export GOOGLE_CREDENTIALS_FILE=/path/to/env-creds.json
mcp-google
# Uses: /path/to/env-creds.json

# Flag overrides environment
export GOOGLE_CREDENTIALS_FILE=/path/to/env-creds.json
mcp-google --credentials /path/to/flag-creds.json
# Uses: /path/to/flag-creds.json
```

## Examples

### Google Service Account

```bash
export GOOGLE_CREDENTIALS_FILE=/path/to/service-account.json
mcp-google
```

### goauth CredentialsSet

```bash
export GOAUTH_CREDENTIALS_FILE=/path/to/credentials.json
export GOAUTH_CREDENTIALS_ACCOUNT=myaccount
mcp-google
```

### Vault-Backed (File)

```bash
export OMNITOKEN_VAULT_URI=file:///path/to/secrets
export OMNITOKEN_CREDENTIALS_NAME=google
mcp-google
```

### 1Password

```bash
export OP_SERVICE_ACCOUNT_TOKEN="ops_..."
export OMNITOKEN_VAULT_URI=op://MyVault
export OMNITOKEN_CREDENTIALS_NAME=google
mcp-google
```

### Bitwarden

```bash
export BW_ACCESS_TOKEN="..."
export BW_ORGANIZATION_ID="..."
export OMNITOKEN_VAULT_URI=bw://org-id
export OMNITOKEN_CREDENTIALS_NAME=google
mcp-google
```

## Shell Configuration

### Bash/Zsh

Add to `~/.bashrc` or `~/.zshrc`:

```bash
# Google MCP Server credentials
export GOOGLE_CREDENTIALS_FILE="$HOME/.config/google-mcp/service-account.json"
```

### Fish

Add to `~/.config/fish/config.fish`:

```fish
set -gx GOOGLE_CREDENTIALS_FILE "$HOME/.config/google-mcp/service-account.json"
```

## Dotenv Files

For project-specific configuration, use a `.env` file:

```bash
# .env
GOOGLE_CREDENTIALS_FILE=/path/to/service-account.json
```

Load with `source` or a dotenv tool:

```bash
source .env
mcp-google
```

Or use a tool like `direnv` for automatic loading.

## Docker

```dockerfile
FROM golang:1.24 AS builder
# ... build steps ...

FROM alpine:latest
COPY --from=builder /app/mcp-google /usr/local/bin/
ENV GOOGLE_CREDENTIALS_FILE=/secrets/service-account.json
CMD ["mcp-google"]
```

Run with mounted secrets:

```bash
docker run -v /path/to/secrets:/secrets mcp-google
```
