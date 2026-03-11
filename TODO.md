# TODO

Carried from: Day 27 (all completed)

## Tasks 59-61: Day 27 tasks
**Status:** completed -- aircraft+carbon in ranker, carbon benchmarks end-to-end, lint sweep

---

## Task 62: Expand airport clusters with 8 new metro areas
**Status:** done
**Plan:** Added 8 multi-airport metro clusters (Bangkok, Istanbul, Beijing, Osaka, Rome, Taipei, Miami, Sao Paulo) to airportClusters map. Updated init capacity. 17 new test cases.
- [x] Write tests for new clusters (NearbyAirports returns siblings, bidirectional lookup)
- [x] Add 8 new metro clusters to airportClusters map
- [x] Update init map capacity
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 63: Add alliance preference filter for search and chat
**Status:** done
**Plan:** Added PreferredAlliance to search.Request, FilterByAlliance to filter.go, preferred_alliance to chat tripParams. Wired into direct pipeline, buildRequestFromParams, mergeParams, parsePartialParams, system prompt, refinement hint. 10 new tests.
- [x] Add PreferredAlliance field to search.Request
- [x] Write test: FilterByAlliance keeps matching flights, removes non-matching
- [x] Implement FilterByAlliance in filter.go
- [x] Wire FilterByAlliance into direct search pipeline
- [x] Add preferred_alliance to tripParams
- [x] Update buildRequestFromParams, mergeParams, parsePartialParams
- [x] Update chatSystemPrompt and refinementHint
- [x] Write chat test: preferred_alliance extraction and merge
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 64: Lint, gofmt sweep, and build gate verification
**Status:** done
**Plan:** Post-merge verification. All gates clean.
- [x] Run gofmt -l . and fix any violations
- [x] Run go vet ./... and fix any warnings
- [x] Run golangci-lint run and fix any issues
- [x] Run go test ./... and verify all pass
