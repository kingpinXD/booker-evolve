# TODO

Carried from: Day 19 (all completed)

## Task 21-25: Day 19 tasks
**Status:** completed (Day 19) — return-date flag, nearby wired, max-price filter, PriceInsights display, multicity test

---

## Task 26: Chat param merging for follow-up searches
**Status:** done
**Plan:** Add mergeParams(prev, partial) that fills zero-value fields from prev. Add parsePartialParams that accepts JSON with at least one recognized field. Store lastParams in chatLoop; on follow-up, merge partial into lastParams. Update refinementHint to instruct LLM to re-emit JSON with changed fields only. Files: cmd/chat.go, cmd/chat_test.go.
- [x] Write mergeParams tests (TDD: full merge, partial cabin, partial date, zero-value)
- [x] Implement mergeParams(prev, partial tripParams) tripParams
- [x] Add parsePartialParams that accepts partial JSON (no required fields)
- [x] Store lastParams in chatLoop, call mergeParams on follow-up
- [x] Update refinementHint to instruct LLM to emit JSON with changes only
- [x] Write chatLoop test: mock LLM emits partial JSON on second turn, verify re-search
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 27: Dynamic ranking profile from conversation
**Status:** done
**Plan:** Reuse existing profiles map from search.go. Add Profile field to tripParams + system prompt. Add --profile flag to chatCmd. chatLoop accepts initial profile; when profile changes in tripParams, log it. Files: cmd/chat.go, cmd/chat_test.go.
- [x] Write profileWeights tests (TDD: budget, comfort, balanced, unknown)
- [x] Implement profileWeights(name string) multicity.RankingWeights
- [x] Add Profile field to tripParams + system prompt mention
- [x] Add --profile flag to chatCmd
- [x] Write profileWeights test: all 3 profiles + unknown defaults
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 28: Airline alliance reference data
**Status:** done
**Plan:** Created search/airlines.go with alliance member data. Implemented Alliance and SameAlliance functions.
- [x] Write Alliance and SameAlliance tests (TDD: same alliance, different, unknown)
- [x] Implement alliance map and lookup functions
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 29: Richer result context in chat history
**Status:** done
**Plan:** Modify resultSummaryForChat to accept tripParams and include top result's route, airline, duration, plus the search params (origin, dest, date, cabin, max_price). Files: cmd/chat.go, cmd/chat_test.go.
- [x] Write enhanced resultSummaryForChat tests (TDD: route, airline, duration, params)
- [x] Implement enhanced summary with top-result details and search param context
- [x] Update callers
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 30: Lint, gofmt sweep, and build gate verification
**Status:** pending
**Plan:** Run all linting and testing tools. Fix any violations.
- [ ] Run gofmt -l . and fix violations
- [ ] Run go vet ./... and fix warnings
- [ ] Run golangci-lint run and fix findings
- [ ] Run go test ./... and verify all pass
