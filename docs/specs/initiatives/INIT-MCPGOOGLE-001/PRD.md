# PRD — Review-aware tools — comments, suggestions, and opt-in upload over gogoogle

**Initiative:** `INIT-MCPGOOGLE-001`
**Repository:** `github.com/plexusone/mcp-google`
**Workflow:** pbhq-lite
**Status:** Draft — pending stakeholder review (pbhq-lite gate: after PRD)

## 1. Problem

mcp-google gives AI assistants read-only access to Google Docs, Sheets, and
Slides. It is a thin MCP wrapper over the
[gogoogle](https://github.com/grokify/gogoogle) SDK. When people review a
shared document, most of their feedback lives where mcp-google cannot see it:

1. **Comments.** Comments come from the Drive API, which mcp-google does not
   call. An assistant asked "what did reviewers say?" sees none of them.
2. **Suggestions.** `get_document_text` and related tools flatten suggested
   insertions and deletions into the original text, so proposed edits are
   indistinguishable from the document itself, and deletions read as if they
   were still present.
3. **Publishing.** An assistant cannot publish a generated file for review.
   Users upload by hand and convert the file to a native Google format by
   hand.

There is also a structural issue. The Docs skill calls the Docs API directly
(`documents.Get`) instead of going through gogoogle's `docsutil`, so it does
not benefit from gogoogle improvements.

And there is a security issue. The per-identity tool policy maps tools to
scopes by matching substrings of tool names, and it **allows** any tool that
matches no pattern. New tools that do not contain "document", "sheet", or
"slide" in their names would bypass scope checks.

**Who is affected:**

- AI assistants and agents that run document review loops.
- Operators who rely on per-identity tool policy.
- Contributors who expect all Google API logic to live in gogoogle.

**Desired state:**

- Assistants can read comments and suggestions in structured form.
- Assistants can optionally upload files, gated by explicit opt-in and
  scopes.
- All Google API calls go through gogoogle.
- Tool authorization is deny-by-default with an explicit scope per tool.

## 2. Goals and Non-Goals

### 2.1 Goals

| ID | Goal | Success criterion |
|----|------|-------------------|
| G-1 | Thin wrapper | No skill calls a Google API client directly; every call goes through a gogoogle package (test: no `google.golang.org/api/*` service method calls in `skills/`) |
| G-2 | Read suggestions | `get_document_suggestions` returns segments classified original / insertion / deletion with original and accepted text per block |
| G-3 | Read comments | `get_file_comments` returns all comments for a Doc, Sheet, or Slides file with quoted anchor text, resolved state, authors, and replies |
| G-4 | Optional publishing | `upload_file` and `update_file` exist only when the server is started with `--enable-write`, and require the `drive:write` policy scope over HTTP for tokens that carry a `scope` claim |
| G-5 | Deny-by-default authorization | Every registered tool has an explicit required scope; an unmapped tool is denied; a test fails if any tool lacks a mapping |
| G-6 | Read-only by default | With default flags, the server requests only read-only Google scopes and registers no write tools |

### 2.2 Non-Goals

- Writing comments, replying, resolving, or accepting suggestions.
- Sharing / permission management.
- Re-implementing any Google API logic in this repository (it belongs in
  gogoogle).
- Converting Markdown or generating documents.

## 3. User Stories

**US-1 — Read what reviewers proposed.** As an *AI assistant*, I want a
document's suggestions separated from its text, so that I can summarize or
apply proposed edits accurately.
Acceptance: `get_document_suggestions{document_id}` on a fixture returns
blocks whose segments match gogoogle's `SuggestionContent`.

**US-2 — Read what reviewers said.** As an *AI assistant*, I want every
comment with the text it refers to, so that I can map feedback to sections.
Acceptance: `get_file_comments{file_id, include_deleted?}` returns all pages
of comments with quoted text, resolved flag, author display names, and
replies.

**US-3 — Publish for review (opt-in).** As an *operator*, I want to allow an
assistant to upload a generated `.docx` / `.pptx` / `.csv` as a native Google
file or update an existing one, so that the review loop needs no manual
upload — but only when I explicitly enable it.
Acceptance: without `--enable-write`, the tools are absent from `tools/list`.
With it, `upload_file{path, name?, folder_id?, convert_to}` returns file ID
and link. Over HTTP, a token that carries a `scope` claim without `drive:write`
is denied and the denial is audit-logged. A token with no `scope` claim is not
gated, as for every other tool.

**US-4 — Trust the policy.** As an *operator*, I want new tools to be denied
unless explicitly mapped to a scope, so that adding tools never silently
widens access.
Acceptance: registry test asserts every tool has a mapping; a tool with no
mapping is denied at call time.

## 4. Functional Requirements

- **PRD-FR-1** All skills MUST call Google APIs only through gogoogle
  packages (`docsutil`, `sheetsutil`, `slidesutil`, `driveutil`).
- **PRD-FR-2** The Docs skill MUST add `get_document_suggestions`.
- **PRD-FR-3** A new Drive skill MUST add `get_file_comments`, usable on any
  file ID.
- **PRD-FR-4** The Drive skill MUST add `upload_file` and `update_file` only
  when `--enable-write` (or `MCP_GOOGLE_ENABLE_WRITE=true`) is set. In that
  mode the server additionally requests the `drive.file` Google scope.
- **PRD-FR-5** Tool authorization MUST use an explicit tool → scope map:
  - `docs:read`, `sheets:read`, `slides:read`, `drive:read` for read tools;
  - `drive:write` for write tools;
  - unmapped tools MUST be denied.
- **PRD-FR-6** Path inputs to write tools MUST be confined to a configured
  root (`--upload-root`, default: working directory), rejecting `..` escapes
  and symlinks that leave the root.
- **PRD-FR-7** Documentation (tools reference, configuration, credentials
  scopes) and release notes MUST cover the new tools, flags, and scopes.
- **PRD-FR-8** Each write call MUST open its local file once, inside the
  root, through an open the operating system confines, check the file it
  holds, and send that file's content with the number of bytes capped at the
  size limit. A link swapped in after the open MUST NOT be followed, and a
  file that grows MUST be cut off and refused.

## 5. Success Metrics (required)

| Metric | Baseline | Target | Measurement Method |
|--------|----------|--------|--------------------|
| Skills calling Google API clients directly | 1 (Docs skill) | 0 | Static test over `skills/` |
| Review feedback types readable (text, suggestions, comments) | 1 of 3 | 3 of 3 | Tool tests with fixtures |
| Tools without explicit scope mapping | All (pattern-based, allow-by-default) | 0; unmapped denied | Registry test; denial test |
| Write tools registered in default mode | 0 | 0 (unchanged) | `tools/list` test without `--enable-write` |
| Test coverage of new code | 0 % | ≥ 80 % | `go test -cover ./...` |

## 6. Out of Scope / Future

- Comment writing / resolution and suggestion acceptance.
- Permission management.
- Other Google products (Gmail, Calendar).
