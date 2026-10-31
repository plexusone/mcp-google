# PLAN — Review-aware tools — comments, suggestions, and opt-in upload over gogoogle

**Initiative:** `INIT-MCPGOOGLE-001`
**Repository:** `github.com/plexusone/mcp-google`
**Version:** 0.1
**Status:** Draft — pending tech-lead review

Non-normative sequencing of [PRD.md](PRD.md) and [TRD.md](TRD.md); execution
state lives in [ROADMAP.md](ROADMAP.md).

## 1. Approach

Authorization comes first. Deny-by-default with an explicit scope map has to
be in place before any new tool is registered; otherwise the new tools would
bypass policy.

Next, route the Docs skill through gogoogle and bump to the gogoogle release
that adds the Drive package and suggestion extraction.

Then add the read tools (suggestions, comments), and only after that the
opt-in write tools. Docs and release come last.

## 2. Affected Components

| Component | Change |
|-----------|--------|
| `internal/emauth` | Explicit tool → scope map |
| `skills/docs` | gogoogle-only access; new suggestions tool |
| `skills/drive` | New skill |
| `internal/auth`, `cmd/mcp-google` | Mode-dependent scopes; new flags |
| `docs/`, `README.md`, `mkdocs.yml`, `CHANGELOG.json` | Updated |

## 3. Work Breakdown

| Step | Approach | Satisfies | Verification | Depends on |
|------|----------|-----------|--------------|------------|
| P-02 | Explicit tool → scope map; deny unmapped when a scope claim is present | PRD-FR-5; TRD-005 | Registry and denial tests | — |
| P-03 | Docs skill uses `docsutil.Service`; bump gogoogle | PRD-FR-1; TRD-001 | Static test | P-02; gogoogle release |
| P-04 | `get_document_suggestions` | PRD-FR-2; TRD-002 | Fixture test | P-03 |
| P-05 | Drive skill with `get_file_comments` | PRD-FR-3; TRD-003 | Paged fake-server test | P-03 |
| P-06 | Opt-in `upload_file` / `update_file`, mode-dependent scopes, path confinement | PRD-FR-4, 6; TRD-004 | `tools/list` and confinement tests | P-05 |
| P-07 | Docs, release notes, v0.5.0 | PRD-FR-7; TRD-006 | CI green, then tag | P-04, P-06, P-08 |
| P-08 | Open uploads once through oscompat's confined, size-checked open and stream the handle to gogoogle | PRD-FR-8; TRD-007 | Growth, swapped-path, and link tests | P-06; gogoogle and oscompat releases |

(Step numbers follow RMI numbers; `RMI-MCPGOOGLE-001` belongs to an earlier
initiative.)

## 4. Sequencing and Dependencies

```
P-02 → P-03 → {P-04, P-05};  P-05 → P-06 → P-08;  {P-04, P-08} → P-07
```

**Cross-repository dependency:** P-03 waits for the gogoogle release that
contains `driveutil/v3` and `ExtractSuggestionContent`. P-02 can proceed
immediately. P-08 waits for the oscompat release that adds
`OpenRegularInRoot` and `CapReader`, and the gogoogle release that uploads and
updates from an `io.Reader`; the two are independent of each other.

## 5. Data and Migration

Not applicable — no persistence.

Deployments that issue scoped tokens need policy updates for the new tools
(TRD §10).

## 6. Testing Strategy

Per TRD §12. Coverage target ≥ 80 % on new code.

## 7. Rollout and Rollback

- **Release.** v0.5.0, with write mode off by default.
- **Tagging.** Tag after CI is green.
- **Rollback.** Pin v0.4.0.

## 8. Risks

| ID | Risk | Likelihood / Impact | Mitigation |
|----|------|---------------------|------------|
| PR-1 | gogoogle release slips | Medium / Medium | P-02 is independent. A temporary relative `replace` is allowed locally and must be removed before push |
| PR-2 | Scoped-token users surprised by deny-unmapped | Low / Medium | Release notes; denial messages name the scope |
| PR-3 | The oscompat or gogoogle release for P-08 slips | Low / Medium | The path-based upload from P-06 remains correct for a trusted upload root, and is documented as such until P-08 lands |
