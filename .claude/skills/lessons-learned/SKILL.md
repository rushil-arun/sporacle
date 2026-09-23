---
name: lessons-learned
description: Repo-specific mistakes past Claude sessions made in sporacle and how to avoid them. Check before non-trivial changes to areas mentioned below (backend concurrency/game state, shared constants generation, trivia data, deployment). Also the place to record a new entry after a significant mistake in this repo.
---

# Lessons learned in this repo

This file accumulates specific, non-obvious mistakes made by past Claude Code sessions
working in sporacle, so future sessions don't repeat them. It is not a general coding
style guide — that belongs in CLAUDE.md. This file is for concrete "I did X, it broke Y,
here's what to do instead" entries.

## When to add an entry

Only after a **significant** mistake — one that broke a build/test, shipped incorrect
behavior, required real rework, or the user explicitly corrected as a non-obvious repo
convention. Do not log: typos caught immediately, normal back-and-forth refinement,
or anything already obvious from reading the code/CLAUDE.md.

## How to add an entry

1. Add a dated entry below using the template.
2. Keep it short: what happened, why it happened (root cause, not just symptom), what
   to do differently.
3. If the lesson is broad enough to be a standing rule (not just a one-off), also
   consider whether it belongs in CLAUDE.md instead/additionally.
4. If this file grows past a handful of unrelated entries, split it into topic-specific
   skills (e.g. `lessons-learned-concurrency`, `lessons-learned-deploy`) rather than
   letting one file sprawl.

### Entry template

```
### YYYY-MM-DD: <short title>
**What happened:** <the mistake and its concrete consequence>
**Root cause:** <why it happened>
**Do instead:** <the corrected approach>
```

## Entries

### 2026-09-22: trivia.db needs `deploy.sh --trivia`, not just a plain deploy
**What happened:** After migrating trivia data from `trivia/*.json` to `trivia/trivia.db`
(SQLite, read via `server/triviadb`), a deploy that only rebuilds/uploads the binary
would leave the EC2 instance's `/opt/sporacle/trivia` without `trivia.db` — the server
would have no categories to serve even though the binary itself is up to date.
**Root cause:** `scripts/deploy.sh` only syncs the `trivia/` directory (which now
contains `trivia.db`, the actual runtime data source) when passed `--trivia`; a plain
`scripts/deploy.sh` only ships the binary.
**Do instead:** Any deploy after a change to `trivia/trivia.db` (schema, seed data, or
first-time migration) must use `scripts/deploy.sh --trivia`. If `trivia.db` ever stops
being committed to git, this becomes required on *every* deploy, not just ones that
touch trivia data.
