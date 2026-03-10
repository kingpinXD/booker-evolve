# Skill: Evolve

Rules for safely modifying the booker codebase.

## Plan Before You Build

Before writing any code for a task, always plan first:

1. Read the task description and understand the goal
2. Explore the relevant code (read files, check existing tests, understand interfaces)
3. Write a short plan as a comment in TODO.md under the task entry:
   - What files need to change
   - What approach you will take
   - What tests you will write
4. Only then start implementing

## TODO Tracking

The file `TODO.md` is the source of truth for session work. Format:

```markdown
# TODO

Carried from: Session N (only if resuming previous work)

## Task 1: [title]
**Status:** pending | in-progress | done | skipped
**Plan:** [1-3 sentences: approach, files, tests]
- [ ] step one
- [ ] step two
- [x] completed step
```

Rules:
- At the start of each task, set status to `in-progress` and write your plan
- Check off steps as you complete them
- If you finish, set status to `done`
- If you cannot finish (timeout, 3-strike revert, budget), set status to `skipped` with a reason
- Unfinished tasks carry over to the next session automatically

## Session Limits

Stop working when EITHER limit is reached (whichever comes first):
- **25 minutes** of wall-clock time in the session
- **70% context window usage** — if you notice responses getting compressed or context feels large, wrap up

When hitting a limit: commit your current work, update TODO.md with remaining tasks, and stop.

## TDD Workflow

1. Write or update a test that captures the desired behavior
2. Run `go test ./...` — verify the new test **fails**
3. Implement the minimal code change to make it pass
4. Run the full verification: `go build ./... && go test ./... && go vet ./... && golangci-lint run`
5. Commit the task with a descriptive message (see Commit Message Guidelines)

## Parallel Execution

Use parallel agents and git worktrees to work on multiple independent tasks simultaneously:

1. Identify tasks that don't depend on each other (e.g., different packages, unrelated features)
2. For each parallel task, create a worktree:
   ```bash
   git worktree add .worktrees/task-N -b evolve-task-N
   ```
3. Use the Task tool to spawn agents that work in separate worktrees concurrently
4. Each agent works independently: implement, test, commit in its own branch
5. At the end, rebase all branches into main:
   ```bash
   git checkout main
   git rebase evolve-task-N
   ```
6. Clean up worktrees: `git worktree remove .worktrees/task-N`

Only parallelize truly independent tasks. If tasks touch the same files, run them sequentially.

## Testing Strategy

Choose test type based on change size. When counting lines, only count `.go` files — exclude all `.md` files (JOURNAL, LEARNINGS, TODO, etc.), generated files (`_gen.go`, `_string.go`), mocks, and docs:

- **Under 400 lines of Go code changed**: unit tests are sufficient
- **400–1000 lines of Go code changed**: write at least one integration test exercising the full code path
- Tests must always use mocked/cached data, never live APIs

## Code Modification Rules

- **Modify existing functions** instead of creating new duplicates
- **Prefer early returns** over deeply nested conditionals
- **Prefer switch** over chains of if-else
- **Commit after each task** — each task gets its own commit. Include context usage in the commit message (see Commit Message Guidelines).
- Target under 300 lines of Go code per session. If a task would exceed this, defer remaining work to the next session via TODO.md.
- **Run the full check** after every change:
  ```bash
  go build ./... && go test ./... && go vet ./... && golangci-lint run
  ```

## Failure Protocol

If tests fail after a change:
1. Read the error carefully. Attempt a targeted fix.
2. Run the check again.
3. If still failing, attempt a second fix.
4. If still failing after 3 total attempts, **revert**:
   ```bash
   git checkout -- <modified-files>
   ```
5. Mark the task as `skipped` in TODO.md with the failure reason.
6. Log the failure in JOURNAL.md and move to the next task.

## BLOCKED.md — Escalation to Human

If you encounter a problem that blocks ALL future work (not just one task), create `BLOCKED.md`:

```markdown
# BLOCKED — Human Intervention Required

**Session:** N
**Task:** [what you were working on]

## Problem
[Exact error output]

## What I Tried
[List of fix attempts and their results]

## Why I'm Blocked
[Why this cannot be resolved automatically]

## Suggested Fix
[Your best guess at what a human should do]
```

Only create BLOCKED.md for systemic issues: broken builds not caused by your changes, environment problems, dependency failures, or pre-existing test failures you cannot fix. If just one task is hard, skip it and move on — do not block the entire agent.

## Protected Files — Never Modify

- `IDENTITY.md`, `PERSONALITY.md`, `VISION.md`
- `scripts/*`
- `.github/workflows/*`
- `skills/*`
- `CLAUDE.md`
- `.golangci.yml`

## API Call Budget

- SerpAPI has a strict free-tier limit. **Max 6 SerpAPI calls per session.**
- Always use cached data (`.cache/flights/`) and the cache layer (`provider/cache/`) instead of live requests.
- Write tests against mocked/cached responses, never against live APIs.
- If a task requires live SerpAPI calls and the budget is exhausted, skip the task.

## Commit Message Guidelines

Format:
```
<type>(<scope>): <short summary>

<body — explain WHY the change was made, not just WHAT changed>

- What problem this solves or what behavior it adds
- Why this approach was chosen over alternatives
- Any trade-offs or caveats worth noting
```

- **Types**: `feat`, `fix`, `test`, `refactor`, `docs`, `chore`
- **Scope**: package or area affected (e.g., `serpapi`, `multicity`, `cache`, `config`)
- **Short summary**: imperative mood, under 72 chars (e.g., "add retry logic for transient failures")
- **Body**: required. Explain the reasoning, not just the diff.
- **One commit per task.** Each task gets its own commit.
- **Session progress**: include `Session: task N/M` at the end of every commit message (N = current task number, M = total tasks this session).
- Never commit code that fails `go build` or `golangci-lint run`

Example:
```
fix(cache): return stale entry when provider is unreachable

The cache layer previously returned an error when the underlying provider
failed and no cached entry existed. This caused the combiner to silently
drop stopover cities that had expired cache entries during API outages.

- Fall back to stale cache entry if provider returns a non-auth error
- Log a warning so stale usage is visible in session logs
- Added unit test for stale-fallback path

Session: task 2/4
```

## Journal & Learnings

After completing each task (success or failure), update the governance logs:

1. **JOURNAL.md** — append a brief entry for the task:
   ```
   ### Session N, Task M -- [title]
   [1-2 sentences: what was done, outcome, any surprises]
   ```

2. **LEARNINGS.md** — if the task revealed something generalizable (a pattern, a pitfall, a useful technique), append:
   ```
   ## Lesson: [one-line insight]
   [Context paragraph: what happened, why it matters, when to apply this]
   ```
   Skip if nothing new was learned. Do not log obvious things.
