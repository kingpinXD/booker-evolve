# Skill: Self-Assess

Evaluate the current state of the booker codebase and produce a prioritized list of issues.

## Steps

1. **Build check:**
   ```bash
   go build ./...
   ```
   Capture and list any compilation errors.

2. **Test check:**
   ```bash
   go test ./...
   ```
   Capture and list any test failures.

3. **Vet check:**
   ```bash
   go vet ./...
   ```
   Capture and list any vet warnings.

4. **Coverage analysis:**
   ```bash
   go test -coverprofile=cover.out ./... 2>/dev/null
   go tool cover -func=cover.out
   ```
   Identify:
   - Packages with 0% coverage (no tests at all)
   - Packages below 50% coverage
   - Functions with 0% coverage in otherwise-tested packages

5. **Lint check:**
   ```bash
   golangci-lint run 2>/dev/null || true
   ```
   Capture lint findings if the tool is available.

## Priority Order

1. Build errors (code does not compile)
2. Test failures (existing tests are broken)
3. Vet warnings (potential bugs)
4. Lint errors (code quality)
5. **New features** that make booker more useful (new stopover cities, better ranking, flexible dates, multi-route support, etc.) — with tests for the new code
6. Idiomatic Go refactors (reorganize packages, improve interfaces, simplify code structure)
7. Coverage for new or recently changed code
8. Coverage for existing untested code (lowest priority — only if nothing else to do)

## API Call Awareness

When planning tasks, note that SerpAPI has a **6 call per session limit**. Deprioritize tasks that require live SerpAPI calls. Prefer tasks that work with cached data, tests, or non-API code.

## Output

Write findings to stdout as a numbered list, highest priority first. Each item should include:
- The issue type (build/test/vet/lint/coverage)
- The package or file affected
- A one-line description of the problem
