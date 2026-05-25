# Credentials

Google MCP Server supports multiple credential sources for authentication.

## Option 1: Google Service Account

The simplest option - use a standard Google Cloud service account JSON file.

### Setup

1. Create a service account in [Google Cloud Console](https://console.cloud.google.com/)
2. Download the JSON credentials file
3. Share documents with the service account email

### Usage

```bash
mcp-google --credentials /path/to/service-account.json
```

### JSON Format

```json
{
  "type": "service_account",
  "project_id": "your-project",
  "private_key_id": "key-id",
  "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
  "client_email": "mcp-server@your-project.iam.gserviceaccount.com",
  "client_id": "123456789",
  "auth_uri": "https://accounts.google.com/o/oauth2/auth",
  "token_uri": "https://oauth2.googleapis.com/token"
}
```

## Option 2: goauth CredentialsSet

Use a [goauth](https://github.com/grokify/goauth) CredentialsSet file for managing multiple credentials.

### Setup

Create a CredentialsSet JSON file with your Google credentials:

```json
{
  "credentials": {
    "myaccount": {
      "type": "gcpsa",
      "gcpsa": {
        "gcpCredentials": {
          "type": "service_account",
          "project_id": "your-project",
          "private_key_id": "key-id",
          "private_key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n",
          "client_email": "mcp-server@your-project.iam.gserviceaccount.com",
          "client_id": "123456789"
        },
        "scopes": [
          "https://www.googleapis.com/auth/presentations.readonly",
          "https://www.googleapis.com/auth/documents.readonly",
          "https://www.googleapis.com/auth/drive.readonly"
        ]
      }
    }
  }
}
```

### Usage

```bash
mcp-google \
  --goauth-credentials-file /path/to/credentials.json \
  --goauth-credentials-account myaccount
```

### Benefits

- Store multiple accounts in one file
- Pre-configure scopes
- Consistent credential format across tools

## Option 3: Vault-Backed Credentials

Use [omnitoken](https://github.com/plexusone/omnitoken) with vault backends for secure credential storage.

### Supported Vault URIs

| URI Pattern | Description | Requirements |
|-------------|-------------|--------------|
| `op://vault` | 1Password | `OP_SERVICE_ACCOUNT_TOKEN` env var |
| `bw://org-id` | Bitwarden | `BW_ACCESS_TOKEN` and `BW_ORGANIZATION_ID` env vars |
| `file:///path/to/dir` | File-based storage | None |
| `env://PREFIX_` | Environment variables with prefix | None |
| `memory://` | In-memory (testing only) | None |

### 1Password

Store your goauth credentials in 1Password and access them securely:

```bash
# Set 1Password service account token
export OP_SERVICE_ACCOUNT_TOKEN="ops_..."

# Use 1Password vault
mcp-google --vault op://MyVault --credentials-name google
```

The credential item in 1Password should contain the goauth Credentials JSON in a field.

### Bitwarden

Store credentials in Bitwarden Secrets Manager:

```bash
# Set Bitwarden credentials
export BW_ACCESS_TOKEN="..."
export BW_ORGANIZATION_ID="..."

# Use Bitwarden vault
mcp-google --vault bw://org-id --credentials-name google
```

### File Vault

```bash
mcp-google --vault file:///path/to/secrets --credentials-name google
```

#### File Vault Structure

```
/path/to/secrets/
└── google.json    # Contains goauth Credentials JSON
```

### Environment Vault

```bash
# With env:// vault URI, credentials are read from environment variables
export GOOGLE_CREDENTIALS='{"type":"gcpsa",...}'
mcp-google --vault env://GOOGLE_ --credentials-name CREDENTIALS
```

## Credential Priority

If multiple options are specified, the server returns an error. Only one credential source is allowed.

## Recommended Approach

| Use Case | Recommended Option |
|----------|-------------------|
| Local development | Google Service Account |
| Multiple Google accounts | goauth CredentialsSet |
| Production with secrets management | Vault-backed |
| Combining with other services | Vault-backed |

## Security Best Practices

1. **Never commit credentials** - Add credentials files to `.gitignore`
2. **Use file permissions** - `chmod 600 service-account.json`
3. **Principle of least privilege** - Only request needed scopes
4. **Rotate keys** - Periodically rotate service account keys
5. **Use vault backends** - For production, use proper secrets management
