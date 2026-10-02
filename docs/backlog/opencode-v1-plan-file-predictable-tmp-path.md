---
worth: maybe
where: plugins/opencode/plugins/revdiff-plan-review.ts:server
added: 2026-10-02
---
# opencode v1 plan review writes the plan to a predictable /tmp path

The plan text is written to `/tmp/revdiff-plan-${sessionID}.md` with `Bun.write` and removed after the
review. The name is fixed by the session id, so two reviews of the same session overwrite each other's
file, and on a shared machine another local user can pre-create the path.

No user-visible symptom on a single-user machine. A unique temp file (the `mktemp` shape the Claude
planning launcher uses) removes both. Surfaced reviewing #369; deferred for the same reason as
`opencode-v1-plan-review-runs-for-any-session.md`: the v1 files stay unchanged while the v2 PR is open.
