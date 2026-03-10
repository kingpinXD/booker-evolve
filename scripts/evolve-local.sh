#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# evolve-local.sh — Run evolution in an isolated git worktree
#
# Protects your main branch by running evolve.sh in a temporary worktree.
# All changes stay on an ephemeral branch until you review them.
# =============================================================================

REPO_ROOT=$(git rev-parse --show-toplevel)
WORKTREE_DIR="$REPO_ROOT/.worktrees/evolve-$(date +%s)"
BRANCH_NAME="evolve-local-$(date +%Y%m%d-%H%M%S)"

echo "=== Booker Local Evolution ==="
echo "Creating isolated worktree at $WORKTREE_DIR..."
echo "Branch: $BRANCH_NAME"
echo ""

git worktree add "$WORKTREE_DIR" -b "$BRANCH_NAME"

cleanup() {
  echo ""
  echo "Cleaning up worktree..."
  cd "$REPO_ROOT"
  git worktree remove "$WORKTREE_DIR" --force 2>/dev/null || true
  git branch -D "$BRANCH_NAME" 2>/dev/null || true
  echo "Cleanup complete."
}
trap cleanup EXIT

cd "$WORKTREE_DIR"

# Copy .env if it exists (needed for API keys)
[[ -f "$REPO_ROOT/.env" ]] && cp "$REPO_ROOT/.env" .

# Run evolution in local mode (no push)
export EVOLVE_LOCAL=1
export EVOLVE_NO_PUSH=1
bash scripts/evolve.sh

echo ""
echo "=== Local evolution complete ==="
echo "Changes are on branch: $BRANCH_NAME"
echo "Review with: git log main..$BRANCH_NAME --oneline"
echo "Diff with:   git diff main..$BRANCH_NAME"
echo ""
echo "Worktree will be cleaned up on exit."
