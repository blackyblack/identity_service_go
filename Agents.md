# Agents.md - Session memory

Session memory is stored as `session.md` on branch `session/<task-branch>`.  
**Rule:** `session.md` MUST NOT be committed to the code branch.

## Workflow

**Start of session:** `./scripts/session_pull.sh`  
**End of session:** Update `session.md`, then `./scripts/session_push.sh`

`session_pull.sh` seeds `session.md` with a template if no prior memory exists. Fill it in as you work and keep it updated through the session.

Enable git hooks so `session_pull.sh` runs after pulls/merges and `session_push.sh` runs before pushes:
`git config core.hooksPath .githooks`

## What to store in session.md

1. **Task and plan** — Record the task description and initial plan first
2. **Solution steps** — Document each step taken to solve the task
3. **Failed attempts** — Log what didn't work and why
4. **Issues encountered** — Note blockers, errors, and how they were resolved

Persist all important steps so the task solution flow can be reproduced.

## Cleanup

Session branches are deleted automatically when the PR is merged.  
Manual: `git push origin --delete session/<task-branch>`
