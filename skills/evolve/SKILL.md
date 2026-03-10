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

Carried from: Day N (only if resuming previous work)

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

## TDD Workflow

1. Write or update a test that captures the desired behavior
2. Run `go test ./...` — verify the new test **fails**
3. Implement the minimal code change to make it pass
4. Run the full verification: `go build ./... && go test ./... && go vet ./... && golangci-lint run`
5. Commit with a descriptive message

## Testing Strategy

Choose test type based on change size (exclude generated files: docs, mocks, `_gen.go`, `_string.go`):

- **Under 400 lines changed**: unit tests are sufficient
- **400–1000 lines changed**: write at least one integration test exercising the full code path
- Tests must always use mocked/cached data, never live APIs

## Code Modification Rules

- **Modify existing functions** instead of creating new duplicates
- **Prefer early returns** over deeply nested conditionals
- **Prefer switch** over chains of if-else
- **Keep each commit small** — target under 300 lines of non-generated code. If a task is larger, split it into multiple commits.
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
5. Log the failure and move to the next task.

## Protected Files — Never Modify

- `IDENTITY.md`, `PERSONALITY.md`
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
- **Body**: required for all commits. Explain the reasoning, not just the diff.
- One logical change per commit
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
```

## Journal & Learnings

After completing each task (success or failure), update the governance logs:

1. **JOURNAL.md** — append a brief entry for the task:
   ```
   ### Day N, Task M -- [title]
   [1-2 sentences: what was done, outcome, any surprises]
   ```

2. **LEARNINGS.md** — if the task revealed something generalizable (a pattern, a pitfall, a useful technique), append:
   ```
   ## Lesson: [one-line insight]
   [Context paragraph: what happened, why it matters, when to apply this]
   ```
   Skip if nothing new was learned. Do not log obvious things.
