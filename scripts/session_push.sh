#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   ./scripts/session_push.sh [task_branch]
#
# If task_branch is omitted, uses the current git branch.
#
# Writes session memory to remote branch copilot/session/<task_branch>
# Keeps a SINGLE rolling commit (amends each update), so branch history does not grow.
# session.md MUST NOT be committed to the code branch.

SESSION_FILE="${SESSION_FILE:-session.md}"
REMOTE="${REMOTE:-origin}"

TASK_BRANCH="${1:-$(git rev-parse --abbrev-ref HEAD)}"
if [[ "$TASK_BRANCH" == "HEAD" ]]; then
  echo "ERROR: detached HEAD. Pass task branch name: $0 <task_branch>" >&2
  exit 2
fi

SESSION_BRANCH="copilot/session/${TASK_BRANCH}"

# Check that session file exists
if [[ ! -f "$SESSION_FILE" ]]; then
  echo "ERROR: ${SESSION_FILE} not found. Nothing to push." >&2
  exit 1
fi

# Create a temporary directory for the orphan worktree
TEMP_DIR=$(mktemp -d)
trap 'rm -rf "$TEMP_DIR"' EXIT

# Try to fetch the session branch to get its current state
if git fetch "$REMOTE" "$SESSION_BRANCH" --quiet 2>/dev/null; then
  # Session branch exists, create worktree from it
  git worktree add --quiet "$TEMP_DIR" FETCH_HEAD --detach
  cd "$TEMP_DIR"
else
  # Session branch doesn't exist, create orphan branch
  git worktree add --quiet "$TEMP_DIR" --detach
  cd "$TEMP_DIR"
  git checkout --orphan "$SESSION_BRANCH"
  git rm -rf . --quiet 2>/dev/null || true
fi

# Copy session file to the worktree
cp "${OLDPWD}/${SESSION_FILE}" "$SESSION_FILE"

# Stage and commit (amend if there's already a commit, otherwise create new)
git add "$SESSION_FILE"

if git rev-parse HEAD >/dev/null 2>&1; then
  # Amend the existing commit to keep history minimal
  git commit --amend --no-edit -m "Update session memory" --quiet
else
  # First commit on orphan branch
  git commit -m "Update session memory" --quiet
fi

# Force push to update the session branch (single rolling commit)
git push --force "$REMOTE" "HEAD:refs/heads/${SESSION_BRANCH}"

cd "$OLDPWD"
git worktree remove "$TEMP_DIR" --force 2>/dev/null || true

echo "Session memory pushed to ${REMOTE}/${SESSION_BRANCH}"
