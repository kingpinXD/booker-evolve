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

SESSION=$(cat SESSION_NUMBER 2>/dev/null || echo 0)
START_SHA=$(git rev-parse HEAD)
TIMESTAMP=$(date -u +%H%M%S)
TIMESTAMP_PRETTY=$(date -u +%H:%M)
LOG_DIR="logs/session${SESSION}"

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

increment_session() {
  local current
  current=$(cat SESSION_NUMBER 2>/dev/null || echo 0)
  echo $((current + 1)) > SESSION_NUMBER
  log "Session counter incremented to $((current + 1))"
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

log "=== Evolution Session: Session $SESSION ==="
log "Start SHA: $START_SHA"
log "Timestamp: $TIMESTAMP_PRETTY UTC"

# --- Check for BLOCKED.md ---
if [[ -f BLOCKED.md ]]; then
  log "BLOCKED.md found. Agent is blocked and waiting for human intervention."
  log "Contents:"
  cat BLOCKED.md | while IFS= read -r line; do log "  $line"; done
  log "Remove BLOCKED.md to unblock the agent. Skipping session."
  increment_session
  exit 0
fi

# =============================================================================
# Phase A: PLANNING
# =============================================================================

log "=== Phase A: PLANNING ==="

# Gather context
IDENTITY=$(cat IDENTITY.md 2>/dev/null || echo "No IDENTITY.md found")
PERSONALITY=$(cat PERSONALITY.md 2>/dev/null || echo "No PERSONALITY.md found")
VISION=$(cat VISION.md 2>/dev/null || echo "No VISION.md found")
JOURNAL_TAIL=$(tail -50 JOURNAL.md 2>/dev/null || echo "No journal yet")
LEARNINGS_TAIL=$(tail -30 LEARNINGS.md 2>/dev/null || echo "No learnings yet")
SELF_ASSESS_SKILL=$(load_skill "skills/self-assess/SKILL.md")

# Check for unfinished TODOs from a previous session
HAS_PENDING_TODOS=false
if [[ -f TODO.md ]]; then
  PENDING_COUNT=$(grep -c '^\- \[ \]' TODO.md 2>/dev/null || true)
  PENDING_COUNT=${PENDING_COUNT:-0}
  PENDING_TASKS=$(grep -c '^\*\*Status:\*\* \(pending\|in-progress\)' TODO.md 2>/dev/null || true)
  PENDING_TASKS=${PENDING_TASKS:-0}
  if [[ "$PENDING_TASKS" -gt 0 || "$PENDING_COUNT" -gt 0 ]]; then
    HAS_PENDING_TODOS=true
    log "Found existing TODO.md with $PENDING_TASKS unfinished tasks ($PENDING_COUNT unchecked steps). Resuming."
  fi
fi

if [[ "$HAS_PENDING_TODOS" == "true" ]]; then
  # Resume: use existing TODO.md and SESSION_PLAN.md, skip fresh planning
  EXISTING_TODO=$(cat TODO.md)
  EXISTING_PLAN=$(cat SESSION_PLAN.md 2>/dev/null || echo "No previous SESSION_PLAN.md")

  RESUME_PROMPT="You are booker-agent, a self-evolving agent for the booker Go project.

$IDENTITY

$PERSONALITY

$VISION

$SELF_ASSESS_SKILL

Recent journal:
$JOURNAL_TAIL

Recent learnings:
$LEARNINGS_TAIL

You have unfinished work from a previous session. Here is the existing TODO.md:

$EXISTING_TODO

And the previous SESSION_PLAN.md:

$EXISTING_PLAN

Your task:
1. Run the self-assessment steps (build, test, vet, coverage) to understand current state.
2. Review the existing TODO.md. Tasks marked pending or in-progress are your priority.
3. If any pending tasks are no longer relevant (e.g., already fixed, obsolete), mark them skipped with a reason.
4. You may add up to 2 NEW tasks if the self-assessment reveals urgent issues (build failures, test failures), but unfinished tasks come first.
5. Update SESSION_PLAN.md to reflect the resumed plan. Keep the same format.
6. Update TODO.md to reflect any changes.

Do NOT discard previous work. Continuity matters."

  claude -p "$RESUME_PROMPT" \
    --allowedTools "Bash(read-only:*),Read,Write,Edit,Glob,Grep" \
    --output-format text \
    2>&1 | tee "$LOG_DIR/phase-a.log"

  log "Resumed planning from previous session."

else
  # Fresh planning: no pending TODOs
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

$VISION

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

Write SESSION_PLAN.md when done.

Then create TODO.md from the plan. Format:
# TODO
## Task 1: [title]
**Status:** pending
**Plan:** [to be filled during implementation]
- [ ] [step from the plan]
- [ ] [step from the plan]

Every task in SESSION_PLAN.md must have a corresponding entry in TODO.md."

  claude -p "$PLAN_PROMPT" \
    --allowedTools "Bash(read-only:*),Read,Write,Edit,Glob,Grep" \
    --output-format text \
    2>&1 | tee "$LOG_DIR/phase-a.log"

  log "Fresh planning complete."
fi

if [[ ! -f SESSION_PLAN.md ]]; then
  log "ERROR: SESSION_PLAN.md was not created. Aborting."
  exit 1
fi

if [[ ! -f TODO.md ]]; then
  log "ERROR: TODO.md was not created. Aborting."
  exit 1
fi

log "Planning phase complete. SESSION_PLAN.md and TODO.md ready."

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
SESSION_TIMEOUT=1500  # 25 minutes
IMPL_SHA=$(git rev-parse HEAD)

IMPL_PROMPT="You are booker-agent, a self-evolving agent for the booker Go project.

$IDENTITY

$PERSONALITY

$EVOLVE_SKILL

Read SESSION_PLAN.md and TODO.md. You have $TASK_COUNT tasks to implement this session.

Session limits — stop when EITHER is reached:
- 25 minutes of wall-clock time
- 70% context window usage (if context feels large or responses compress, wrap up)
When hitting a limit, commit current work, update TODO.md with remaining tasks, and stop.

Parallel execution:
- Identify tasks that are independent (different packages, unrelated features).
- For independent tasks, use git worktrees and the Task tool to work in parallel:
    git worktree add .worktrees/task-N -b evolve-task-N
  Spawn a parallel agent per worktree using the Task tool.
- Tasks that touch the same files must run sequentially.
- At the end, rebase all branches into main:
    git checkout main && git rebase evolve-task-N
  Then clean up: git worktree remove .worktrees/task-N && git branch -d evolve-task-N

For each task:
1. Read the task in SESSION_PLAN.md and the corresponding TODO.md entry.
2. Set the task status to in-progress in TODO.md.
3. Write your plan under the task entry: what files to change, what approach, what tests.
4. Break the work into checkbox steps in TODO.md if not already done.
5. Implement using TDD: write/update test first, verify it fails, then implement.
6. Check off TODO steps as you complete them.
7. Run the full verification: go build ./... && go test ./... && go vet ./... && golangci-lint run
8. Commit after each task with a descriptive message. Always include TODO.md, JOURNAL.md, and
   LEARNINGS.md in the commit (git add TODO.md JOURNAL.md LEARNINGS.md before committing).
   Include 'Session: task N/M' at the end of every commit message (N = current task, M = total).

Testing rules (line counts only include .go files — exclude .md files, generated files, mocks, docs):
- Under 400 lines of Go code changed: unit tests are sufficient.
- 400-1000 lines of Go code changed: include at least one integration test.

After completing each task (success OR failure):
- Set the task status in TODO.md to done or skipped (with reason).
- Append a brief entry to JOURNAL.md: ### Session $SESSION, Task N -- [title] + 1-2 sentences.
- If you learned something generalizable, append to LEARNINGS.md.

If tests fail, you have 3 attempts to fix. After 3 failures:
- Revert your code changes with git checkout.
- Mark the task as skipped in TODO.md with the failure reason.
- If the failure blocks ALL future work, create BLOCKED.md (see evolve skill for format).
- If just one task is hard, skip it and move on.

Do not modify any protected files (except JOURNAL.md, LEARNINGS.md, and TODO.md which you must update)."

# macOS lacks GNU timeout — use background process + kill.
# Run claude directly (no pipe) so $! captures the real PID.
claude -p "$IMPL_PROMPT" \
  --allowedTools "Bash,Read,Write,Edit,Glob,Grep,Task" \
  --output-format text \
  > "$LOG_DIR/phase-b.log" 2>&1 &
CLAUDE_PID=$!

# Kill after SESSION_TIMEOUT seconds if still running
(sleep "$SESSION_TIMEOUT" && kill "$CLAUDE_PID" 2>/dev/null) &
TIMER_PID=$!

wait "$CLAUDE_PID" 2>/dev/null || {
  log "Implementation phase timed out or failed"
}

# Clean up timer if claude finished before timeout
kill "$TIMER_PID" 2>/dev/null || true
wait "$TIMER_PID" 2>/dev/null || true

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

Session: $SESSION
Changes since session start:
$CHANGES_SINCE

Tasks planned (read SESSION_PLAN.md for details):
$(head -50 SESSION_PLAN.md 2>/dev/null || echo "No plan")

Do the following:
1. Append a journal entry to JOURNAL.md: ## Session $SESSION -- $TIMESTAMP_PRETTY -- [title]
2. If you learned anything generalizable, append to LEARNINGS.md
3. If any GitHub issues were addressed, write ISSUE_RESPONSE.md"

claude -p "$REFLECT_PROMPT" \
  --allowedTools "Bash(read-only:*),Read,Write,Edit,Glob,Grep" \
  --output-format text \
  2>&1 | tee "$LOG_DIR/phase-c.log"

# Post issue comments and close fixed issues
if [[ -f ISSUE_RESPONSE.md ]] && command -v gh >/dev/null 2>&1; then
  log "Processing issue responses..."
  ISSUE_NUM=""
  COMMENT=""
  IS_FIXED=false

  post_and_close() {
    local num="$1" body="$2" fixed="$3"
    if [[ -z "$num" || -z "$body" ]]; then return; fi
    gh issue comment "$num" --body "$body" 2>/dev/null || \
      log "Failed to post comment on issue #$num"
    if [[ "$fixed" == "true" ]]; then
      gh issue close "$num" 2>/dev/null && \
        log "Closed issue #$num (fixed)" || \
        log "Failed to close issue #$num"
    fi
  }

  while IFS= read -r line; do
    if [[ "$line" =~ ^##\ Issue\ \#([0-9]+) ]]; then
      post_and_close "${ISSUE_NUM:-}" "$COMMENT" "$IS_FIXED"
      ISSUE_NUM="${BASH_REMATCH[1]}"
      COMMENT=""
      IS_FIXED=false
    elif [[ -n "${ISSUE_NUM:-}" ]]; then
      if [[ "$line" =~ ^\*\*Status:\*\*.*fixed ]]; then
        IS_FIXED=true
      fi
      COMMENT="${COMMENT}${line}
"
    fi
  done < ISSUE_RESPONSE.md
  post_and_close "${ISSUE_NUM:-}" "$COMMENT" "$IS_FIXED"
fi

log "Reflection phase complete"

# =============================================================================
# FINALIZE SESSION STATE (SESSION_NUMBER + any uncommitted .md files)
# =============================================================================

log "Finalizing session state..."
increment_session
git add SESSION_NUMBER TODO.md SESSION_PLAN.md JOURNAL.md LEARNINGS.md 2>/dev/null || true
if ! git diff --cached --quiet 2>/dev/null; then
  PENDING_LEFT=$(grep -c '^\*\*Status:\*\* \(pending\|in-progress\)' TODO.md 2>/dev/null || true)
  PENDING_LEFT=${PENDING_LEFT:-0}
  git commit -m "$(cat <<COMMIT_EOF
chore(session$SESSION): finalize session state

Session counter incremented to $(cat SESSION_NUMBER).
Tasks remaining: $PENDING_LEFT (carried to next session in TODO.md).
COMMIT_EOF
  )" 2>/dev/null || true
fi

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

  TAG="session${SESSION}-${TIMESTAMP}"
  git tag "$TAG"

  if [[ "$NO_PUSH" != "1" ]]; then
    git push origin main --tags
    log "Pushed to origin with tag: $TAG"
  else
    log "Local mode: skipping push. Tag: $TAG"
  fi
else
  log "BUILD VERIFICATION FAILED after 3 attempts. Creating BLOCKED.md for human review."

  # Capture the failure details
  BUILD_OUT=$(go build ./... 2>&1 || true)
  TEST_OUT=$(go test ./... 2>&1 | tail -40 || true)
  VET_OUT=$(go vet ./... 2>&1 || true)
  LINT_OUT=$(golangci-lint run 2>&1 | tail -20 || true)

  cat > BLOCKED.md <<BLOCKED_EOF
# BLOCKED — Human Intervention Required

**Session:** $SESSION
**Timestamp:** $TIMESTAMP_PRETTY UTC
**Start SHA:** $START_SHA
**Current SHA:** $(git rev-parse HEAD)

## What Happened

Build verification failed after 3 fix attempts. The agent cannot resolve this automatically.

## Failures

### go build
\`\`\`
$BUILD_OUT
\`\`\`

### go test (last 40 lines)
\`\`\`
$TEST_OUT
\`\`\`

### go vet
\`\`\`
$VET_OUT
\`\`\`

### golangci-lint (last 20 lines)
\`\`\`
$LINT_OUT
\`\`\`

## How to Unblock

1. Fix the failing checks above
2. Run \`make verify\` to confirm everything passes
3. Delete this file: \`rm BLOCKED.md\`
4. Commit and push — the next scheduled session will resume automatically
BLOCKED_EOF

  git add -A
  git commit -m "$(cat <<'COMMIT_EOF'
chore(session): blocked — build verification failed, needs human help

Build verification failed after 3 automated fix attempts on Session $SESSION.
The agent has exhausted its retry budget and cannot resolve the issue.

BLOCKED.md contains the full failure output (build, test, vet, lint).
The agent will skip all future sessions until BLOCKED.md is removed.

A human needs to:
1. Review the failures in BLOCKED.md
2. Fix the underlying issue
3. Delete BLOCKED.md and push
COMMIT_EOF
  )" 2>/dev/null || true

  if [[ "$NO_PUSH" != "1" ]]; then
    git push origin main 2>/dev/null || log "WARNING: push failed"
    log "Pushed BLOCKED.md to origin. Agent will wait for human intervention."
  else
    log "Local mode: BLOCKED.md created. Agent will wait for human intervention."
  fi
fi

# --- Finalize ---
log "=== Session complete ==="
