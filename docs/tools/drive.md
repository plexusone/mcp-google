# Google Drive Tools

Drive tools read the comments on any file and, only when you turn on
[write mode](#write-mode), upload files and replace their content. They work on
Docs, Sheets, and Slides files as well as any other Drive file.

| Tool | Mode | Scope | Purpose |
|------|------|-------|---------|
| `get_file_comments` | default | `drive:read` | Comments with quoted text, resolved state, and replies |
| `upload_file` | write mode only | `drive:write` | Upload a local file, optionally converting it to a Google file |
| `update_file` | write mode only | `drive:write` | Replace a Drive file's content, keeping its ID |

!!! note "Read-only by default"
    In default mode the server registers no tool that can change Google data and
    requests only read-only Google scopes. `upload_file` and `update_file` do not
    appear in `tools/list` until the server is started with `--enable-write`.

---

## get_file_comments

List the comments on a Drive file, with the text each is anchored to, whether it
is resolved, and its replies. Every page of results is returned, or an error:
never a partial list.

### Input

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_id` | string | Yes | The file ID, or a Drive, Docs, Sheets, or Slides URL |
| `include_deleted` | boolean | No | Include deleted comments (default: false) |

### Output

| Field | Type | Description |
|-------|------|-------------|
| `file_id` | string | The file ID that was read |
| `comments` | array | The comments, in the order Drive returns them |

Each comment has:

| Field | Type | Description |
|-------|------|-------------|
| `id` | string | Comment ID |
| `content` | string | The comment text |
| `quoted_text` | string | The text in the file the comment is anchored to |
| `anchor` | string | The anchor, when the API provides one |
| `author` | object | `display_name`, and `email_address` only when the API returns it |
| `created_time`, `modified_time` | string | RFC 3339 timestamps |
| `resolved` | boolean | Whether the thread is resolved |
| `deleted` | boolean | Whether the comment is deleted |
| `replies` | array | Replies, each with `id`, `content`, `action` (`resolve` or `reopen` when the reply did that), `author`, and `created_time` |

### Example

```json
{
  "name": "get_file_comments",
  "arguments": {
    "file_id": "https://docs.google.com/document/d/1BxiMVs0XRA5nFMdKvBdBZjgmUUqptlbs74OgvE2upms/edit"
  }
}
```

!!! warning "Treat author email addresses as personal data"
    Email addresses appear only when Google returns them. Do not pass them on to
    systems that do not need them.

!!! tip "Suggested edits are not comments"
    Use [`get_document_suggestions`](docs.md#get_document_suggestions) for the
    suggested edits in a Google Doc. Together with the document text and these
    comments, that covers all three kinds of review feedback.

---

## Write mode

Write mode lets an agent publish a document for review and keep it up to date
without a manual upload. It is off unless you ask for it.

### Turn it on

```bash
mcp-google serve --enable-write --upload-root ./to-publish
```

| Flag | Environment variable | Description |
|------|----------------------|-------------|
| `--enable-write` | `MCP_GOOGLE_ENABLE_WRITE=true` | Register `upload_file` and `update_file`. A service account file also requests the `drive.file` Google scope. A flag given on the command line, including `--enable-write=false`, wins over the variable |
| `--upload-root` | `MCP_GOOGLE_UPLOAD_ROOT` | The directory the tools may read local files from (default: the working directory) |

The upload root must exist when the server starts. The same flags apply to the
`upload-file` and `update-file` [CLI commands](../cli.md), which refuse to run
without `--enable-write`.

In write mode the server requests the `drive.file` Google scope in addition to
the read-only scopes. `drive.file` covers only the files the app creates or
opens, never the whole Drive. If your credentials come from a goauth file or a
vault, their configured scopes must include `drive.file`; for a service account
file the server adds it for you.

### upload_file

Upload a local file, optionally converting it to a Google Doc, Sheet, or Slides
file.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `path` | string | Yes | Path of the local file, relative to the upload root |
| `name` | string | No | Name for the Drive file (default: the file name, without its extension when converting) |
| `folder_id` | string | No | ID of the Drive folder to put the file in |
| `convert_to` | string | No | `doc`, `sheet`, or `slides`. Omit to store the file unchanged |

Output is the new file: `id`, `name`, `mime_type`, `web_view_link`,
`head_revision_id`, and `modified_time`.

```json
{
  "name": "upload_file",
  "arguments": {
    "path": "review/packet.docx",
    "convert_to": "doc",
    "folder_id": "1AbCdEfGh"
  }
}
```

### update_file

Replace the content of an existing Drive file with a local file. The file ID,
link, sharing, and comments are kept, so reviewers keep working in the same
document.

| Parameter | Type | Required | Description |
|-----------|------|----------|-------------|
| `file_id` | string | Yes | The ID or URL of the Drive file |
| `path` | string | Yes | Path of the local file, relative to the upload root |
| `convert_to` | string | No | `doc`, `sheet`, or `slides`. Omit to store the file unchanged |

!!! note "Which files can be updated"
    With the `drive.file` scope, `update_file` can change only files that the
    server's credentials created, or that a person opened with the app. A file
    that was merely shared with a service account is outside that scope, and
    Drive refuses the update as `permission_denied` or `not_found`. Publish the
    file with `upload_file` first, and update the copy the server created. The
    server does not ask for the broader `drive` scope, which would reach the
    whole Drive.

### Conversions

| `convert_to` | Accepted extensions | Result |
|--------------|---------------------|--------|
| `doc` | `.docx`, `.odt`, `.txt`, `.html` | Google Doc |
| `sheet` | `.csv`, `.xlsx`, `.ods` | Google Sheet |
| `slides` | `.pptx`, `.odp` | Google Slides file |

An extension that does not match the target is refused as `unsupported_conversion`.

### What a path may be

A path is confined to the upload root. These are refused with
`path_outside_root`, the same way on every platform, before anything is sent to
Google:

- a `..` element, with `/` or `\` as the separator;
- a rooted or drive-qualified path that is not inside the root, such as
  `/etc/passwd`, `C:x`, or a UNC path;
- a path that leads outside the root through a symbolic link.

The path must name a regular file no larger than **50 MB**.

!!! note "The check is made when the tool runs"
    The path, the file type, and the size are checked, and the file is then
    handed to the upload. A symbolic link created inside the upload root after
    the check, but before the file is read, could still be followed, and a file
    that grows in that interval could exceed the size limit. Keep the upload
    root writable only by people who are allowed to upload from anywhere the
    server can read.

### Errors

A failed write returns an error result whose text is JSON,
`{"code": "...", "message": "..."}`:

| Code | Meaning |
|------|---------|
| `path_outside_root` | The path leaves the upload root |
| `not_found` | No such local file, or Drive says the file does not exist or is not shared with the credentials |
| `permission_denied` | Drive refused the request. Check the credentials' access and that `drive.file` was granted |
| `unsupported_conversion` | `convert_to` is not `doc`, `sheet`, or `slides`, or the extension does not match |
| `file_too_large` | The file is over the 50 MB limit |
| `invalid_request` | A missing or malformed argument, or a path that is not a regular file |
| `write_disabled` | A write was requested with write mode off (the CLI commands) |
| `api_error` | Any other Drive failure |

### Audit log

Every write call is logged, whether it succeeds or fails, with who made it, which
tool, the file name, the Drive file an update targets, and the resulting file ID
or the failure code:

```json
{"level":"INFO","msg":"write_call","tool":"upload_file","name":"packet.docx","sub":"user@example.com","client_id":"client-1","scope":"drive:write","file_id":"1AbCdEf"}
{"level":"WARN","msg":"write_call_failed","tool":"upload_file","name":"x","sub":"user@example.com","code":"path_outside_root"}
{"level":"WARN","msg":"write_call_failed","tool":"update_file","name":"packet.docx","target":"1AbCdEf","sub":"user@example.com","code":"not_found"}
```

Over stdio there is no verified identity, so `sub` is `local`. File contents,
credentials, and paths on the server's disk are never logged.

### Service accounts

A service account sees only the files shared with it, and the files it creates
are owned by it. Share the target folder with the service account's email
address before uploading into it.
