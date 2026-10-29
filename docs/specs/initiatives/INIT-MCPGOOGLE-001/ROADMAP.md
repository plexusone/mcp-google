# ROADMAP — Review-aware tools — comments, suggestions, and opt-in upload over gogoogle

**Initiative:** `INIT-MCPGOOGLE-001`
**Repository:** `github.com/plexusone/mcp-google`
**Status:** Planned — 0 of 2 phases completed

> RMI IDs are stable and permanent. Commits implementing an item carry the trailer `Refs: RMI-MCPGOOGLE-<NNN>`. Phase status is derived from member RMIs. RMI `NNN` corresponds to PLAN step `P-NN`.

## Phase 1 — Policy, Thin Wrapper, and Read Tools

**Theme:** Deny-by-default tool authorization, all Google access through gogoogle, and tools for reading suggestions and comments
**Status:** Planned — 0 of 4 items completed

- [ ] `RMI-MCPGOOGLE-002` Explicit tool-to-scope authorization map; deny unmapped tools when a scope claim is present
  - Type: capability · Status: planned
  - Satisfies: PRD-FR-5, TRD-005
  - Acceptance: registry test asserts every tool is mapped; unmapped tool denied and audit-logged
- [ ] `RMI-MCPGOOGLE-003` Route the Docs skill through gogoogle docsutil and bump gogoogle
  - Type: capability · Status: planned
  - Depends on: `RMI-MCPGOOGLE-002`, `RMI-GOGOOGLE-007`
  - Satisfies: PRD-FR-1, TRD-001
  - Acceptance: static test finds no direct Google API client calls in `skills/`
- [ ] `RMI-MCPGOOGLE-004` get_document_suggestions tool
  - Type: capability · Status: planned
  - Depends on: `RMI-MCPGOOGLE-003`
  - Satisfies: PRD-FR-2, TRD-002
  - Acceptance: fixture test output matches gogoogle SuggestionContent
- [ ] `RMI-MCPGOOGLE-005` Drive skill with get_file_comments
  - Type: capability · Status: planned
  - Depends on: `RMI-MCPGOOGLE-003`
  - Satisfies: PRD-FR-3, TRD-003
  - Acceptance: paged fake-server test returns all comments with quoted text, resolved state, and replies

## Phase 2 — Opt-in Publishing and Release

**Theme:** Upload and update tools behind explicit opt-in, scope gating, and path confinement; docs and v0.5.0
**Status:** Planned — 0 of 2 items completed

- [ ] `RMI-MCPGOOGLE-006` Opt-in upload_file and update_file with drive:write scope, drive.file OAuth scope, and upload-root confinement
  - Type: capability · Status: planned
  - Depends on: `RMI-MCPGOOGLE-005`
  - Satisfies: PRD-FR-4, PRD-FR-6, TRD-004
  - Acceptance: tools absent without `--enable-write`; path escapes rejected; denial without `drive:write`
- [ ] `RMI-MCPGOOGLE-007` Tools reference, configuration and scope docs, release notes, v0.5.0
  - Type: release · Status: planned
  - Depends on: `RMI-MCPGOOGLE-004`, `RMI-MCPGOOGLE-006`
  - Satisfies: PRD-FR-7, TRD-006
  - Acceptance: CI green before tag; release notes call out the authorization change
