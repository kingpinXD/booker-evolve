#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# evolve-cron.sh — Cron wrapper for booker-agent evolution
#
# Designed for macOS launchd. Handles:
#   - Lock file to prevent overlapping runs
#   - Git pull before evolution
#   - Output logging
# =============================================================================

LOCK_FILE="/tmp/booker-evolve.lock"
REPO_DIR="${BOOKER_EVOLVE_DIR:-/Users/tanmay/IdeaProjects/kingpinXD/booker-evolve}"
NETWORK_WAIT_INTERVAL=30   # seconds between retries
NETWORK_WAIT_MAX=20        # max retries (20 × 30s = 10 minutes)

# --- Network Check ---

wait_for_network() {
  for attempt in $(seq 1 "$NETWORK_WAIT_MAX"); do
    if curl -sfm 5 https://github.com > /dev/null 2>&1; then
      echo "[$(date -u +%H:%M:%S)] Network available."
      return 0
    fi
    echo "[$(date -u +%H:%M:%S)] No network (attempt $attempt/$NETWORK_WAIT_MAX). Retrying in ${NETWORK_WAIT_INTERVAL}s..."
    sleep "$NETWORK_WAIT_INTERVAL"
  done
  echo "[$(date -u +%H:%M:%S)] Network unavailable after $((NETWORK_WAIT_MAX * NETWORK_WAIT_INTERVAL))s. Skipping session."
  return 1
}

# --- Lock Management ---

cleanup() {
  rm -f "$LOCK_FILE"
}
trap cleanup EXIT

if [[ -f "$LOCK_FILE" ]]; then
  OLD_PID=$(cat "$LOCK_FILE" 2>/dev/null || echo "")
  if [[ -n "$OLD_PID" ]] && kill -0 "$OLD_PID" 2>/dev/null; then
    echo "Another evolution is running (PID $OLD_PID). Exiting."
    exit 0
  fi
  echo "Stale lock file found (PID $OLD_PID). Removing."
  rm -f "$LOCK_FILE"
fi

echo $$ > "$LOCK_FILE"

# --- Verify Repo ---

if [[ ! -d "$REPO_DIR" ]]; then
  echo "ERROR: Repo not found at $REPO_DIR"
  echo "Clone it: git clone git@github.com:kingpinXD/booker-evolve.git $REPO_DIR"
  exit 1
fi

cd "$REPO_DIR"

# --- Wait for Network ---

wait_for_network || exit 0

# --- Sync ---

echo "[$(date -u +%H:%M:%S)] Pulling latest changes..."
git pull --rebase origin main || echo "WARNING: git pull failed, continuing with local state"

# --- Run Evolution ---

mkdir -p logs

echo "[$(date -u +%H:%M:%S)] Starting evolution session..."
bash scripts/evolve.sh 2>&1 | tee -a logs/cron.log

echo "[$(date -u +%H:%M:%S)] Cron session complete."
