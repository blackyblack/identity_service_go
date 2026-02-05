#!/usr/bin/env bash
set -euo pipefail

# Usage:
#   ./scripts/session_pull.sh [task_branch]
#
# If task_branch is omitted, uses the current git branch.

SESSION_FILE="${SESSION_FILE:-session.md}"
REMOTE="${REMOTE:-origin}"

TASK_BRANCH="${1:-$(git rev-parse --abbrev-ref HEAD)}"
if [[ "$TASK_BRANCH" == "HEAD" ]]; then
  echo "ERROR: detached HEAD. Pass task branch name: $0 <task_branch>" >&2
  exit 2
fi

SESSION_BRANCH="session/${TASK_BRANCH}"

# Try to fetch session branch. If it doesn't exist, memory is empty.
if git fetch "$REMOTE" "$SESSION_BRANCH" --quiet 2>/dev/null; then
  if git show "FETCH_HEAD:${SESSION_FILE}" > "$SESSION_FILE" 2>/dev/null; then
    echo "Loaded session memory from ${REMOTE}/${SESSION_BRANCH} into ${SESSION_FILE}"
  else
    : > "$SESSION_FILE"
    echo "Session branch exists but ${SESSION_FILE} missing - created empty ${SESSION_FILE}"
  fi
else
  : > "$SESSION_FILE"
  echo "No prior session memory for ${REMOTE}/${SESSION_BRANCH} - created empty ${SESSION_FILE}"
fi
