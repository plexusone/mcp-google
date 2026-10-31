# TRD — Review-aware tools — comments, suggestions, and opt-in upload over gogoogle

**Initiative:** `INIT-MCPGOOGLE-001`
**Repository:** `github.com/plexusone/mcp-google`
**Version:** 0.1
**Status:** Draft

MUST / MUST NOT / SHOULD / MAY are used in the RFC 2119 sense.

## 1. Introduction

**Purpose.** Specify the new tools, the refactor routing all Google API
access through gogoogle, the opt-in write mode, and deny-by-default tool
authorization.

**Scope.** This repository's skills, server flags, and authorization only.

**References.** [PRD](PRD.md) · [PLAN](PLAN.md) · [ROADMAP](ROADMAP.md) ·
gogoogle `driveutil/v3`, `docsutil/v1` (`ExtractSuggestionContent`) —
released in the gogoogle version this work bumps to.

## 2. Technical Requirements and Traceability

| TRD ID | Requirement | Source | Verification |
|--------|-------------|--------|--------------|
| TRD-001 | Skills use only gogoogle packages for Google API calls | PRD-FR-1 | Static test: no `*docs.DocumentsService`, `*sheets.Service`, `*slides.Service`, or `*drive.Service` method calls in `skills/` |
| TRD-002 | `get_document_suggestions` per §6.1 | PRD-FR-2 | Tool test with recorded fixture |
| TRD-003 | Drive skill with `get_file_comments` per §6.2 | PRD-FR-3 | Paged fake-server test |
| TRD-004 | Opt-in write tools per §6.3 | PRD-FR-4, 6 | `tools/list` tests with and without flag; path-confinement tests |
| TRD-005 | Explicit tool → scope map, deny unmapped (§6.4) | PRD-FR-5 | Registry and denial tests |
| TRD-006 | Docs and release notes | PRD-FR-7 | Docs review; mkdocs build |
| TRD-007 | Write tools open each file once through a confined open, check the handle, and stream it capped at the limit (§6.3) | PRD-FR-8 | Growth, swapped-path, and link tests |

## 3. System Invariants

- **INV-1** In default mode the server requests only read-only Google
  scopes and registers no tool that can modify Google data.
- **INV-2** Every registered tool has exactly one required policy scope; a
  call to a tool without one is denied.
- **INV-3** Skills contain no Google API request logic; they translate MCP
  inputs and outputs around gogoogle calls.
- **INV-4** A write call opens its local file once. What is sent is the
  content of that open handle, never more than the size limit, and a link
  created at the path afterwards is never followed.

## 4. Architecture

```
MCP client ─► mcp-google (omniskill runtime, stdio or HTTP + ID-JAG policy)
                ├─ docs skill   ─► gogoogle/docsutil
                ├─ sheets skill ─► gogoogle/sheetsutil
                ├─ slides skill ─► gogoogle/slidesutil
                └─ drive skill (new) ─► gogoogle/driveutil  (write tools only with --enable-write)
```

| Component | Change |
|-----------|--------|
| `skills/docs` | Use `docsutil.Service` for all fetches; add `get_document_suggestions` |
| `skills/drive` (new) | `get_file_comments`; `upload_file`, `update_file` behind opt-in, reading the file through oscompat's confined, size-checked open and streaming the handle to gogoogle |
| `internal/auth` | Scope set computed from mode: read-only, plus `drive.file` when write is enabled |
| `internal/emauth` | Explicit tool → scope map; deny unmapped |
| `cmd/mcp-google` | `--enable-write`, `--upload-root` flags and env vars; register drive skill |

## 5. State and Lifecycle

Not applicable — stateless request/response tools.

## 6. Interface Contracts

### 6.1 `get_document_suggestions`

- **Input:** `{ "document_id": string }`.
- **Output:** gogoogle `SuggestionContent` JSON: `document_id`, `blocks[]`
  with `location`, `segments[]` (`kind`, `text`, `suggestion_ids`),
  `original` and `accepted`, plus `suggestion_ids`.
- **Fetch:** `docsutil.GetDocumentWithSuggestions(..., "SUGGESTIONS_INLINE")`.

### 6.2 `get_file_comments`

- **Input:** `{ "file_id": string, "include_deleted"?: bool }`.
- **Output:** `{ "file_id": string, "comments": [Comment] }`, where each
  gogoogle `Comment` has `quoted_text`, `resolved`, `author`, `replies[]`, and
  timestamps. The list is a field of an object because MCP structured tool
  results are objects.
- **Scope:** works for Docs, Sheets, and Slides files.

### 6.3 Write tools (opt-in)

| Tool | Input | Output |
|------|-------|--------|
| `upload_file` | `{ path, name?, folder_id?, convert_to: "doc"\|"sheet"\|"slides"\|"" }` | gogoogle `File` JSON (`id`, `name`, `mime_type`, `web_view_link`) |
| `update_file` | `{ file_id, path, convert_to? }` | gogoogle `File` JSON |

**Registration.** These tools are registered only when `--enable-write` is
set or `MCP_GOOGLE_ENABLE_WRITE=true`.

**Path handling.** `path` is taken from `--upload-root` and the file is
opened once, through oscompat's `OpenRegularInRoot`, which uses the
operating system's confined open (`os.Root`). A request is rejected, before
anything is sent to Google, when:

- the name has a `..` element, contains a NUL, or is rooted or
  drive-qualified without being inside the root (`path_outside_root`);
- the path is empty (`invalid_request`);
- any element of the path is a symlink that leads out of the root. The open
  itself refuses it, so there is no gap between the check and the open
  (`path_outside_root`);
- the path is not a regular file (`invalid_request`);
- the file is larger than 50 MB (`file_too_large`).

A symlink whose target is relative and stays inside the root is followed. A
symlink with an absolute target is refused even when it points back into the
root, because the confined open does not follow absolute links.

**Streaming.** The checks are made on the open handle, not on the path. The
handle is wrapped in oscompat's `CapReader` and passed to gogoogle as
`UploadRequest.Content` or `UpdateContent`, so gogoogle never opens the path.
A file that grows after the size check is cut off at the limit and refused as
`file_too_large`. A link swapped in at the path after the open is never
followed, because the content sent is the handle's.

**Audit.** Every write call is audit-logged with the identity, tool, file
name, and resulting file ID.

### 6.4 Authorization map

| Tool | Required policy scope |
|------|-----------------------|
| `get_document_*` (metadata, content, text, paragraphs, suggestions) | `docs:read` |
| `get_spreadsheet_metadata`, `list_sheets`, `get_sheet_values`, `get_sheet_data`, `get_multiple_ranges` | `sheets:read` |
| `get_presentation*`, `list_slides`, `get_slide*` | `slides:read` |
| `get_file_comments` | `drive:read` |
| `upload_file`, `update_file` | `drive:write` |

The map is a literal Go table keyed by exact tool name, replacing substring
matching. When a token carries no scope claim, behavior stays as today
(stdio and tokens without scopes are not gated). When a scope claim is
present, unmapped tools are denied.

### 6.5 Errors

- **Tool errors** return MCP `isError: true` with JSON
  `{ code, message }`. Codes:
  - `not_found`
  - `permission_denied`
  - `unsupported_conversion`
  - `path_outside_root`
  - `write_disabled`
  - `file_too_large`
  - `invalid_request`
  - `api_error`
- **Scope denials** keep the existing denial format and are audit-logged.

## 7. Data Design

No persistence. Tool outputs reuse gogoogle JSON types unchanged, so the
wrapper adds no schema of its own.

## 8. Operational Guarantees

| Category | Requirement | Measurement | Applicability |
|----------|-------------|-------------|---------------|
| Performance | Comment listing makes one API call per page | Fake-server request count | Applicable |
| Auditability | Every write call and every denial is audit-logged | Log-capture tests | Applicable |
| Availability, throughput | Not applicable — per-process tool server | — | — |

## 9. Security and Privacy

- **Least privilege.** Default mode requests only the read-only scopes
  (INV-1). Write mode adds only `drive.file`, never full `drive`.
- **Deny-by-default.** Explicit per-tool scopes (INV-2) close the current gap
  where unmapped tools are allowed.
- **Filesystem confinement.** Upload paths are confined to `--upload-root`
  by a single, operating-system-confined open (INV-4). The size check and the
  read cap apply to the file that is actually sent.
- **Personal data.** Comment author emails are passed through only when the
  API returns them. The docs advise treating them as personal data.
- **Not applicable.** Encryption at rest (no storage); in transit, Google
  client TLS and the existing HTTP transport settings apply.

## 10. Compatibility and Migration

- **Applicable.** Existing tool names, inputs, and outputs are unchanged.
- **Authorization change.** Deny-unmapped applies only to tokens that carry
  a scope claim. Deployments that issue scoped tokens must add `drive:read`
  and/or `drive:write` to use the new tools. This is noted in the release
  notes.

## 11. Dependencies

| Dependency | Version | Purpose |
|------------|---------|---------|
| `github.com/grokify/gogoogle` | v0.14.0 or later (adds `driveutil/v3`, `ExtractSuggestionContent`, and upload and update from an `io.Reader`) | Google API SDK |
| `github.com/grokify/oscompat` | v0.11.0 or later (adds `OpenRegularInRoot` and `CapReader`) | Confined, size-checked file open and capped reader |
| `github.com/plexusone/omniskill` | current (v0.12.0) | Skill runtime |
| `github.com/modelcontextprotocol/go-sdk` | current | MCP |

## 12. Testing Strategy

- **Tool tests.** Each tool is tested against gogoogle fake servers (via
  `option.WithEndpoint`).
- **Registry test.** Asserts INV-2 (every tool has a scope mapping).
- **Static test.** Asserts INV-3 (no direct Google API calls in skills).
- **Write mode.** `tools/list` snapshots with and without `--enable-write`.
- **Path confinement.** Tests cover `..`, absolute paths outside the root,
  and symlinks, including links with relative and absolute targets.
- **Single open.** Tests change the file after it is opened: one grows it past
  the limit, one replaces the path with a link to a file outside the root. The
  first is refused as `file_too_large`, and the second still sends the file
  that was opened (INV-4).

## 13. Deployment

- **Release.** Ship as a minor release (v0.5.0) after the gogoogle and
  oscompat releases are tagged.
- **Rollout.** Write mode is off by default.
- **Rollback.** Pin v0.4.0.

## 14. Alternatives Considered

| Alternative | Why rejected |
|-------------|--------------|
| Implement Drive calls directly in a new skill | Duplicates SDK logic; other Go programs and the gogoogle CLI need the same functions |
| Keep substring-based scope mapping and add "file" / "comment" patterns | Still allow-by-default; fragile as tool names grow |
| Always register write tools, relying on Google scopes to fail | Exposes write intent to every client and widens requested OAuth scopes by default |
| Check the path, then let gogoogle open it | Leaves a gap between the check and the open: a link swapped in is followed, and a file that grows is sent whole. Closing it needs the open handle, so the content is sent from the handle |
| Copy the file to a private temporary file and pass its path | Closes the same gap but copies up to 50 MB and adds temporary-file handling that streaming the handle does not need |

## 15. Risks and Open Questions

| ID | Risk / Question | Mitigation |
|----|-----------------|------------|
| R-1 | Scoped-token deployments lose access to new tools until policies are updated | Release notes; clear denial messages naming the missing scope |
| R-2 | Service accounts cannot see files not shared with them | Documented in setup guide (share the folder with the service account) |
| R-3 | A symlink with an absolute target is refused even when it points inside the root | Documented in the tools reference and release notes; a relative link inside the root works |
| R-4 | `drive.file` lets `update_file` reach only files the server's credentials created | Documented; publish with `upload_file` and update the copy it created |

No open question blocks implementation.

## Appendix: Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1 | 2026-10-29 | John Wang | Initial draft |
| 0.2 | 2026-10-31 | John Wang | Add the single confined open and streamed upload (INV-4, TRD-007) |
