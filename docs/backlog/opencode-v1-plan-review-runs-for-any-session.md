---
worth: maybe
where: plugins/opencode/plugins/revdiff-plan-review.ts:server
added: 2026-10-02
---
# opencode v1 plan review runs for any session's idle event

The `event` hook reacts to every `session.idle` and reads `event.properties.sessionID` without checking
which session it is. Any session whose last assistant message is in plan mode opens a review, a subagent
or background session included, and nothing stops the same completed plan from opening a second review
when another idle event arrives for it.

No report of this from a user. Surfaced reviewing #369, where the v2 integration states the rule
explicitly: only the root session the client is showing, one review per completed plan. Deferred because
the v1 files stay unchanged while the v2 PR is open, and whether v1 is worth changing at all once v2
support exists is undecided.
