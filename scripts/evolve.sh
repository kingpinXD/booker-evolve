#!/usr/bin/env bash
set -euo pipefail

# =============================================================================
# evolve.sh — Main orchestrator for booker-agent self-evolution
#
# 3-phase pipeline: PLANNING → IMPLEMENTATION → REFLECTION
# With build verification and auto-rollback safety.
# =============================================================================

# --- Configuration ---
REPO_ROOT=$(git rev-parse --show-toplevel)
cd "$REPO_ROOT"

DAY=$(cat DAY_COUNT 2>/dev/null || echo 0)
START_SHA=$(git rev-parse HEAD)
TIMESTAMP=$(date -u +%H%M%S)
TIMESTAMP_PRETTY=$(date -u +%H:%M)
LOG_DIR="logs/day${DAY}"

PLAN_BUDGET=2
TASK_BUDGET=3
REFLECT_BUDGET=0.50
MAX_TASKS=5
TASK_TIMEOUT=900  # 15 minutes

# Local mode flags (set by evolve-local.sh)
NO_PUSH="${EVOLVE_NO_PUSH:-0}"

# --- Helper Functions ---

log() {
  local msg="[$(date -u +%H:%M:%S)] $*"
  echo "$msg"
  echo "$msg" >> "$LOG_DIR/evolve.log"
}

check_prerequisites() {
  local missing=()
  command -v claude >/dev/null 2>&1 || missing+=("claude")
  command -v go >/dev/null 2>&1    || missing+=("go")
  command -v git >/dev/null 2>&1   || missing+=("git")

  if [[ ${#missing[@]} -gt 0 ]]; then
    echo "ERROR: Missing required tools: ${missing[*]}"
    exit 1
  fi
}

increment_day() {
  local current
  current=$(cat DAY_COUNT 2>/dev/null || echo 0)
  echo $((current + 1)) > DAY_COUNT
  log "Day counter incremented to $((current + 1))"
}

load_skill() {
  local skill_file="$1"
  if [[ -f "$skill_file" ]]; then
    cat "$skill_file"
  else
    log "WARNING: Skill file not found: $skill_file"
    echo ""
  fi
}

build_verify() {
  log "Running build verification..."
  gofmt -w . 2>/dev/null || true

  if ! go build ./...; then
    log "FAIL: go build"
    return 1
  fi

  if ! go test ./...; then
    log "FAIL: go test"
    return 1
  fi

  if ! go vet ./...; then
    log "FAIL: go vet"
    return 1
  fi

  if command -v golangci-lint >/dev/null 2>&1; then
    if ! golangci-lint run; then
      log "FAIL: golangci-lint"
      return 1
    fi
  fi

  log "Build verification passed"
  return 0
}

safe_rollback() {
  log "Rolling back to $START_SHA"
  git reset --hard "$START_SHA"
}

# --- Setup ---

mkdir -p "$LOG_DIR"
check_prerequisites

if [[ "$NO_PUSH" != "1" ]]; then
  git pull --rebase origin main || log "WARNING: git pull failed, continuing"
fi

log "=== Evolution Session: Day $DAY ==="
log "Start SHA: $START_SHA"
log "Timestamp: $TIMESTAMP_PRETTY UTC"

# =============================================================================
# Phase A: PLANNING
# =============================================================================

log "=== Phase A: PLANNING ==="

# Gather context
IDENTITY=$(cat IDENTITY.md 2>/dev/null || echo "No IDENTITY.md found")
PERSONALITY=$(cat PERSONALITY.md 2>/dev/null || echo "No PERSONALITY.md found")
JOURNAL_TAIL=$(tail -50 JOURNAL.md 2>/dev/null || echo "No journal yet")
LEARNINGS_TAIL=$(tail -30 LEARNINGS.md 2>/dev/null || echo "No learnings yet")
SELF_ASSESS_SKILL=$(load_skill "skills/self-assess/SKILL.md")

# Fetch and format GitHub issues (if gh is available)
ISSUES_BLOCK=""
if command -v gh >/dev/null 2>&1; then
  ISSUES_RAW=$(gh issue list --limit 10 --json number,title,body,labels,comments 2>/dev/null || echo "[]")
  if [[ "$ISSUES_RAW" != "[]" ]] && [[ -f scripts/format_issues.py ]]; then
    ISSUES_BLOCK=$(echo "$ISSUES_RAW" | python3 scripts/format_issues.py 2>/dev/null || echo "")
  fi
fi

PLAN_PROMPT="You are booker-agent, a self-evolving agent for the booker Go project.

$IDENTITY

$PERSONALITY

$SELF_ASSESS_SKILL

Recent journal:
$JOURNAL_TAIL

Recent learnings:
$LEARNINGS_TAIL

${ISSUES_BLOCK:+Open issues:
$ISSUES_BLOCK}

Your task: Assess the booker codebase and create SESSION_PLAN.md with up to $MAX_TASKS tasks.
Run the self-assessment steps first (build, test, vet, coverage), then prioritize work.

Each task in SESSION_PLAN.md should follow this format:
## Task 1: [title]
**Priority:** [high|medium|low]
**Description:** [what to do]
**Files:** [likely files to touch]
**Test:** [how to verify success]

Write SESSION_PLAN.md when done."

claude -p "$PLAN_PROMPT" \
  --allowedTools "Bash(read-only:*),Read,Write,Edit,Glob,Grep" \
  --max-budget-usd "$PLAN_BUDGET" \
  --output-format text \
  2>&1 | tee "$LOG_DIR/phase-a.log"

if [[ ! -f SESSION_PLAN.md ]]; then
  log "ERROR: SESSION_PLAN.md was not created. Aborting."
  exit 1
fi

log "Planning complete. SESSION_PLAN.md created."

# =============================================================================
# Phase B: IMPLEMENTATION
# =============================================================================

log "=== Phase B: IMPLEMENTATION ==="

TASK_COUNT=$(grep -c "^## Task" SESSION_PLAN.md 2>/dev/null || echo 0)
if [[ "$TASK_COUNT" -gt "$MAX_TASKS" ]]; then
  TASK_COUNT=$MAX_TASKS
fi

log "Tasks found: $TASK_COUNT"

EVOLVE_SKILL=$(load_skill "skills/evolve/SKILL.md")

for i in $(seq 1 "$TASK_COUNT"); do
  log "--- Task $i of $TASK_COUNT ---"
  TASK_SHA=$(git rev-parse HEAD)

  TASK_PROMPT="You are booker-agent, a self-evolving agent for the booker Go project.

$IDENTITY

$PERSONALITY

$EVOLVE_SKILL

Read SESSION_PLAN.md and implement Task $i.
Follow TDD: write/update test first, verify it fails, then implement.
After changes, run the full verification:
  go build ./... && go test ./... && go vet ./... && golangci-lint run

Testing rules (line counts exclude generated files: docs, mocks, *_gen.go, *_string.go):
- Under 400 lines changed: unit tests are sufficient.
- 400-1000 lines changed: include at least one integration test.

Commit size rules:
- Keep each commit under 300 lines of non-generated code.
- If a task is larger, split into multiple small commits.
- Check your diff size before committing:
  git diff --cached --stat -- '*.go' ':!*_gen.go' ':!*_string.go' ':!*mock*' ':!docs/*'

Commit message rules:
- Format: type(scope): short summary, followed by a body paragraph.
- The body must explain WHY the change was made, the reasoning, and any trade-offs.
- Never write a one-liner commit message. Always include a body.

After completing each task (success OR failure):
- Append a brief entry to JOURNAL.md: ### Day $DAY, Task \$i -- [title] + 1-2 sentences.
- If you learned something generalizable, append to LEARNINGS.md.
- These updates are part of the task, not optional.

If tests fail, you have 3 attempts to fix. After 3 failures, revert with git checkout.
Do not modify any protected files (except JOURNAL.md and LEARNINGS.md which you must update)."

  timeout "$TASK_TIMEOUT" claude -p "$TASK_PROMPT" \
    --allowedTools "Bash,Read,Write,Edit,Glob,Grep" \
    --max-budget-usd "$TASK_BUDGET" \
    --output-format text \
    2>&1 | tee "$LOG_DIR/task-${i}.log" || {
      log "Task $i timed out or failed, reverting to $TASK_SHA"
      git reset --hard "$TASK_SHA"
    }

  log "Task $i complete"
done

log "Implementation phase complete"

# =============================================================================
# Phase C: REFLECTION
# =============================================================================

log "=== Phase C: REFLECTION ==="

COMM_SKILL=$(load_skill "skills/communicate/SKILL.md")
CHANGES_SINCE=$(git log --oneline "$START_SHA"..HEAD 2>/dev/null || echo "No changes")

REFLECT_PROMPT="You are booker-agent, a self-evolving agent for the booker Go project.

$IDENTITY

$PERSONALITY

$COMM_SKILL

Session: Day $DAY
Changes since session start:
$CHANGES_SINCE

Tasks planned (read SESSION_PLAN.md for details):
$(head -50 SESSION_PLAN.md 2>/dev/null || echo "No plan")

Do the following:
1. Append a journal entry to JOURNAL.md: ## Day $DAY -- $TIMESTAMP_PRETTY -- [title]
2. If you learned anything generalizable, append to LEARNINGS.md
3. If any GitHub issues were addressed, write ISSUE_RESPONSE.md"

claude -p "$REFLECT_PROMPT" \
  --allowedTools "Bash(read-only:*),Read,Write,Edit,Glob,Grep" \
  --max-budget-usd "$REFLECT_BUDGET" \
  --output-format text \
  2>&1 | tee "$LOG_DIR/phase-c.log"

# Post issue comments if ISSUE_RESPONSE.md exists
if [[ -f ISSUE_RESPONSE.md ]] && command -v gh >/dev/null 2>&1; then
  log "Posting issue responses..."
  while IFS= read -r line; do
    if [[ "$line" =~ ^##\ Issue\ \#([0-9]+) ]]; then
      ISSUE_NUM="${BASH_REMATCH[1]}"
      # Collect the section until next ## or EOF
      COMMENT=""
    elif [[ -n "${ISSUE_NUM:-}" ]]; then
      if [[ "$line" =~ ^## ]]; then
        # Post previous issue comment
        if [[ -n "$COMMENT" ]]; then
          gh issue comment "$ISSUE_NUM" --body "$COMMENT" 2>/dev/null || \
            log "Failed to post comment on issue #$ISSUE_NUM"
        fi
        if [[ "$line" =~ ^##\ Issue\ \#([0-9]+) ]]; then
          ISSUE_NUM="${BASH_REMATCH[1]}"
          COMMENT=""
        else
          ISSUE_NUM=""
        fi
      else
        COMMENT="${COMMENT}${line}
"
      fi
    fi
  done < ISSUE_RESPONSE.md
  # Post last issue comment
  if [[ -n "${ISSUE_NUM:-}" ]] && [[ -n "${COMMENT:-}" ]]; then
    gh issue comment "$ISSUE_NUM" --body "$COMMENT" 2>/dev/null || \
      log "Failed to post comment on issue #$ISSUE_NUM"
  fi
fi

log "Reflection phase complete"

# =============================================================================
# BUILD VERIFICATION
# =============================================================================

log "=== BUILD VERIFICATION ==="

VERIFIED=false
for attempt in 1 2 3; do
  log "Verification attempt $attempt/3"
  if build_verify; then
    VERIFIED=true
    break
  fi
  log "Verification failed on attempt $attempt"
  if [[ $attempt -lt 3 ]]; then
    gofmt -w .
    git add -A && git commit -m "chore: auto-fix formatting" 2>/dev/null || true
  fi
done

if [[ "$VERIFIED" == "true" ]]; then
  # Check total diff size (non-generated .go files only)
  DIFF_LINES=$(git diff --stat "$START_SHA"..HEAD -- '*.go' ':!*_gen.go' ':!*_string.go' ':!*mock*' ':!docs/*' | tail -1 | grep -oE '[0-9]+ insertion' | grep -oE '[0-9]+' || echo 0)
  log "Total non-generated lines changed since session start: $DIFF_LINES"

  TAG="day${DAY}-${TIMESTAMP}"
  git tag "$TAG"

  if [[ "$NO_PUSH" != "1" ]]; then
    git push origin main --tags
    log "Pushed to origin with tag: $TAG"
  else
    log "Local mode: skipping push. Tag: $TAG"
  fi
else
  log "BUILD VERIFICATION FAILED after 3 attempts. Rolling back."
  safe_rollback
  if [[ "$NO_PUSH" != "1" ]]; then
    git push origin main --force-with-lease
    log "Rolled back and pushed to $START_SHA"
  else
    log "Local mode: rolled back to $START_SHA"
  fi
fi

# --- Finalize ---
increment_day
log "=== Session complete ==="
