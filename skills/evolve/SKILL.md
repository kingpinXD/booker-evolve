# Skill: Evolve

Rules for safely modifying the booker codebase.

## TDD Workflow

1. Write or update a test that captures the desired behavior
2. Run `go test ./...` — verify the new test **fails**
3. Implement the minimal code change to make it pass
4. Run `go build ./... && go test ./... && go vet ./...` — all must pass
5. Commit with a descriptive message

## Code Modification Rules

- **Modify existing functions** instead of creating new duplicates
- **Prefer early returns** over deeply nested conditionals
- **Prefer switch** over chains of if-else
- **Keep changes small** — one logical change per commit
- **Run the full check** after every change:
  ```bash
  go build ./... && go test ./... && go vet ./...
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

## Commit Guidelines

- Message format: `<type>: <description>` (e.g., `test: add coverage for currency package`)
- Types: `feat`, `fix`, `test`, `refactor`, `docs`, `chore`
- One logical change per commit
- Never commit code that fails `go build`
