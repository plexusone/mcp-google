# ROADMAP — Review-aware tools — comments, suggestions, and opt-in upload over gogoogle

**Initiative:** `INIT-MCPGOOGLE-001`
**Repository:** `github.com/plexusone/mcp-google`
**Status:** In progress — 2 of 3 phases completed

> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-MCPGOOGLE-<NNN>`. Phase status is derived from member RMIs. RMI `NNN` corresponds to PLAN step `P-NN`.

## Phase 1 — Policy, Thin Wrapper, and Read Tools

**Theme:** Deny-by-default tool authorization, all Google access through gogoogle, and tools for reading suggestions and comments
**Status:** Completed — 4 of 4 items completed

- [x] `RMI-MCPGOOGLE-002` Explicit tool-to-scope authorization map; deny unmapped tools when a scope claim is present
  - Type: capability · Status: completed
  - Satisfies: PRD-FR-5, TRD-005
  - Acceptance: registry test asserts every tool is mapped; unmapped tool denied and audit-logged
- [x] `RMI-MCPGOOGLE-003` Route the Docs skill through gogoogle docsutil and bump gogoogle
  - Type: capability · Status: completed
  - Depends on: `RMI-MCPGOOGLE-002`, `RMI-GOGOOGLE-007`
  - Satisfies: PRD-FR-1, TRD-001
  - Acceptance: static test finds no direct Google API client calls in `skills/`
- [x] `RMI-MCPGOOGLE-004` get_document_suggestions tool
  - Type: capability · Status: completed
  - Depends on: `RMI-MCPGOOGLE-003`
  - Satisfies: PRD-FR-2, TRD-002
  - Acceptance: fixture test output matches gogoogle SuggestionContent
- [x] `RMI-MCPGOOGLE-005` Drive skill with get_file_comments
  - Type: capability · Status: completed
  - Depends on: `RMI-MCPGOOGLE-003`
  - Satisfies: PRD-FR-3, TRD-003
  - Acceptance: paged fake-server test returns all comments with quoted text, resolved state, and replies

## Phase 2 — Opt-in Publishing and Release

**Theme:** Upload and update tools behind explicit opt-in, scope gating, and path confinement; docs and v0.5.0
**Status:** In progress — 1 of 2 items completed

- [x] `RMI-MCPGOOGLE-006` Opt-in upload_file and update_file with drive:write scope, drive.file OAuth scope, and upload-root confinement
  - Type: capability · Status: completed
  - Depends on: `RMI-MCPGOOGLE-005`
  - Satisfies: PRD-FR-4, PRD-FR-6, TRD-004
  - Acceptance: tools absent without `--enable-write`; path escapes rejected; a token with a `scope` claim but no `drive:write` is denied (tokens with no `scope` claim are not gated)
- [ ] `RMI-MCPGOOGLE-007` Tools reference, configuration and scope docs, release notes, v0.5.0
  - Type: release · Status: planned
  - Depends on: `RMI-MCPGOOGLE-004`, `RMI-MCPGOOGLE-008`
  - Satisfies: PRD-FR-7, TRD-006
  - Acceptance: CI green before tag; release notes call out the authorization change

## Phase 3 — Confined Uploads

**Theme:** Open each upload once through a confined, size-checked handle and stream it to Drive
**Status:** Completed — 1 of 1 items completed

- [x] `RMI-MCPGOOGLE-008` Open uploads once through a confined, size-checked handle and stream them
  - Type: capability · Status: completed
  - Depends on: `RMI-MCPGOOGLE-006`, `RMI-GOGOOGLE-008`
  - Satisfies: PRD-FR-8, TRD-007
  - Acceptance: `upload_file` and `update_file` open the file once inside the upload root, refuse links out of it, and cap the stream at the size limit; a file that grows is refused as `file_too_large` and a link swapped in after the open is not followed
