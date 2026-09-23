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

_(none yet)_
