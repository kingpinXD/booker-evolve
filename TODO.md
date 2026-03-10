# TODO

Carried from: Day 17 (all completed)

## Task 11: Fix errcheck lint in cmd/chat.go
**Status:** completed (Day 17)

## Task 12: Add result summary to chat conversation history
**Status:** completed (Day 17)

## Task 13: Add airport cluster data
**Status:** completed (Day 17)

## Task 14: Flex-date multi-search in direct strategy
**Status:** completed (Day 17)

## Task 15: Surface airport suggestions in chat system prompt
**Status:** completed (Day 17)

---

## Task 16: Nearby-airport search strategy
**Status:** done (Day 18)
**Plan:** Created search/nearby/ package with Searcher that expands origin/dest via airport clusters, fans out delegate calls, merges/deduplicates, sorts by price. 9 tests.
- [x] Write tests: mock delegate strategy, verify fan-out to cluster airports
- [x] Write tests: deduplication of results from multiple airport pairs
- [x] Write tests: MaxResults cap, no-cluster fallback (delegate as-is)
- [x] Implement search/nearby/nearby.go with Strategy that expands clusters
- [x] Verify: `go test ./search/nearby/... -race`

## Task 17: Round-trip support in direct strategy
**Status:** done (Day 18)
**Plan:** Extracted searchFlights helper, added combineRoundTrip for 2-leg itineraries with summed prices. One-way path unchanged. 2 new tests.
- [x] Write tests: round-trip produces 2-leg itinerary with combined price
- [x] Write tests: one-way behavior unchanged when ReturnDate is empty
- [x] Implement return-leg search and itinerary combination in direct.go
- [x] Verify: `go test ./search/direct/... -race`

## Task 18: Extract shared cmd infrastructure
**Status:** done (Day 18)
**Plan:** Created cmd/infra.go with buildPicker(weights, leg2Date) helper. Reduced ~30 duplicated lines across runSearch and runChat.
- [x] Create cmd/infra.go with buildPicker helper
- [x] Refactor runSearch to use buildPicker
- [x] Refactor runChat to use buildPicker
- [x] Verify: `go test ./cmd/... -race` (all existing tests pass)

## Task 19: Structured refinement guidance in chat
**Status:** done (Day 18)
**Plan:** Added refinementHint() returning available levers; appended as system message to history after results. 2 new tests.
- [x] Write test: refinement hint with specific levers appears in history after results
- [x] Add refinementHint function returning available levers
- [x] Append hint to conversation history after result summary
- [x] Verify: `go test ./cmd/... -race`

## Task 20: Lint and gofmt sweep
**Status:** done (Day 18)
**Plan:** Fixed 1 gofmt violation in direct.go (worktree agent output). Zero lint issues.
- [x] Run gofmt -l . and fix any violations
- [x] Run golangci-lint run and fix any issues
- [x] Verify: zero issues reported
