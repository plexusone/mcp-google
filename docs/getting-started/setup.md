# Setup

## Create a Google Cloud Service Account

1. Go to the [Google Cloud Console](https://console.cloud.google.com/)
2. Create a new project or select an existing one
3. Enable the **Google Slides API** and **Google Docs API**:
   - Go to APIs & Services > Library
   - Search for "Google Slides API" and enable it
   - Search for "Google Docs API" and enable it
4. Create a service account:
   - Go to APIs & Services > Credentials
   - Click "Create Credentials" > "Service Account"
   - Give it a name (e.g., "mcp-server")
   - No special roles needed for read-only access
5. Download the JSON credentials file:
   - Click on the service account
   - Go to "Keys" tab
   - Add Key > Create new key > JSON
   - Save the downloaded file securely

## Share Documents with the Service Account

The service account email (found in the credentials JSON as `client_email`) needs access to documents:

1. Open the Google Slides presentation or Google Doc
2. Click "Share"
3. Add the service account email (e.g., `mcp-server@your-project.iam.gserviceaccount.com`)
4. Grant "Viewer" access (read-only is sufficient)

!!! tip "Sharing with Google Workspace"
    If you're using Google Workspace, you can share entire folders or drives with the service account for broader access.

## Required OAuth Scopes

The server uses these read-only scopes:

- `https://www.googleapis.com/auth/presentations.readonly` - Read Slides presentations
- `https://www.googleapis.com/auth/documents.readonly` - Read Docs documents
- `https://www.googleapis.com/auth/drive.readonly` - Access files shared with the service account

These scopes are automatically configured when using the server.
