# Skill: Release

Gates and conventions for tagging a session release.

## Release Gates

ALL of the following must pass before tagging:

```bash
go build ./...         # Clean compilation
go test ./...          # All tests pass
go vet ./...           # No vet warnings
gofmt -l .             # No unformatted files (empty output = pass)
golangci-lint run      # No lint errors
```

If any gate fails, do NOT tag. Fix or defer to next session.

## Tagging Convention

Session tags use the format:
```
day{N}-{HHMMSS}
```

Example: `day5-143022` means Day 5 session tagged at 14:30:22 UTC.

## Rules

- Never force-push tags
- Never delete existing tags
- Verify all gates pass BEFORE creating the tag
- Semver tags (v1.0.0, etc.) are reserved for human decisions — the agent does not create semver tags
- After tagging, push both the branch and tags:
  ```bash
  git push origin main --tags
  ```
