#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# install-cron.sh — Install/uninstall booker evolution launchd job
# =============================================================================

PLIST_NAME="com.booker.evolve"
PLIST_SRC="$(git rev-parse --show-toplevel)/${PLIST_NAME}.plist"
PLIST_DST="$HOME/Library/LaunchAgents/${PLIST_NAME}.plist"
LOCK_FILE="/tmp/booker-evolve.lock"
GUI_DOMAIN="gui/$(id -u)"

usage() {
  echo "Usage: $0 {install|uninstall|status}"
  exit 1
}

cmd_install() {
  if [[ ! -f "$PLIST_SRC" ]]; then
    echo "ERROR: Plist not found at $PLIST_SRC"
    exit 1
  fi

  mkdir -p "$HOME/Library/LaunchAgents"
  cp "$PLIST_SRC" "$PLIST_DST"
  echo "Copied plist to $PLIST_DST"

  # Unload first if already loaded (ignore errors)
  launchctl bootout "$GUI_DOMAIN/$PLIST_NAME" 2>/dev/null || true

  launchctl bootstrap "$GUI_DOMAIN" "$PLIST_DST"
  echo "Loaded $PLIST_NAME into launchd."
  echo ""
  echo "Trigger manually:  launchctl kickstart $GUI_DOMAIN/$PLIST_NAME"
  echo "Check status:      $0 status"
}

cmd_uninstall() {
  launchctl bootout "$GUI_DOMAIN/$PLIST_NAME" 2>/dev/null || true
  rm -f "$PLIST_DST"
  echo "Unloaded and removed $PLIST_NAME."
}

cmd_status() {
  echo "=== launchd status ==="
  launchctl print "$GUI_DOMAIN/$PLIST_NAME" 2>/dev/null || echo "Not loaded."
  echo ""
  echo "=== Lock file ==="
  if [[ -f "$LOCK_FILE" ]]; then
    PID=$(cat "$LOCK_FILE" 2>/dev/null || echo "?")
    if kill -0 "$PID" 2>/dev/null; then
      echo "Running (PID $PID)"
    else
      echo "Stale lock (PID $PID, process not running)"
    fi
  else
    echo "No lock file. Not running."
  fi
}

case "${1:-}" in
  install)   cmd_install ;;
  uninstall) cmd_uninstall ;;
  status)    cmd_status ;;
  *)         usage ;;
esac
