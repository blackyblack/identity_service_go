# Agents.md - Session memory

Session memory is stored as `session.md` on a separate remote branch:
- Code branch: <task-branch>
- Session branch: session/<task-branch>
Rule: `session.md` MUST NOT be committed to the code branch.

Agent workflow:
1) Start of task session:
   - Run: `./scripts/session_pull.sh`
   - Outcome:
     - If session branch exists, `session.md` is populated.
     - If not, `session.md` is created empty (new task forked from main).

2) End of task session:
   - Update `session.md` with your notes.
   - Run: `./scripts/session_push.sh`
   - Outcome:
     - Session branch is updated with a single rolling commit (no history growth).

Cleanup:
- Session branches are deleted automatically when the PR is merged (GitHub Actions).
- Manual cleanup (if needed): run workflow "Session branch cleanup" or delete:
  `git push origin --delete session/<task-branch>`

Notes:
- If `session_push.sh` fails with a lease/conflict error, run pull -> merge your notes -> push again.
