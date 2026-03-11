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
**Status:** pending
**Plan:** Add profile field to tripParams. Create profileWeights mapping function. Rebuild picker when profile changes. Add --profile flag to chat.
- [ ] Write profileWeights tests (TDD: budget, comfort, balanced, unknown)
- [ ] Implement profileWeights(name string) multicity.RankingWeights
- [ ] Add Profile field to tripParams
- [ ] Modify chatLoop to rebuild picker when profile changes
- [ ] Add --profile flag to chatCmd
- [ ] Write chatLoop test: profile in params triggers correct weights
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 28: Airline alliance reference data
**Status:** pending
**Plan:** Create search/airlines.go with alliance member data. Implement Alliance and SameAlliance functions.
- [ ] Write Alliance and SameAlliance tests (TDD: same alliance, different, unknown)
- [ ] Implement alliance map and lookup functions
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 29: Richer result context in chat history
**Status:** pending
**Plan:** Enhance resultSummaryForChat to include top result details and search parameters.
- [ ] Write enhanced resultSummaryForChat tests (TDD: route, airline, duration, params)
- [ ] Implement enhanced summary with top-result details and search param context
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 30: Lint, gofmt sweep, and build gate verification
**Status:** pending
**Plan:** Run all linting and testing tools. Fix any violations.
- [ ] Run gofmt -l . and fix violations
- [ ] Run go vet ./... and fix warnings
- [ ] Run golangci-lint run and fix findings
- [ ] Run go test ./... and verify all pass
