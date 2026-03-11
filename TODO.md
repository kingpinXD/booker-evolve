# TODO

Carried from: Day 28 (all completed)

## Tasks 62-64: Day 28 tasks
**Status:** completed -- airport clusters expanded to 22 metros, alliance preference filter end-to-end, build gate clean

---

## Task 65: Wire PreferredAlliance into multicity strategy
**Status:** done
**Plan:** Added PreferredAlliance to multicity.SearchParams, mapped from search.Request in toSearchParams, applied FilterByAlliance in filter stage for both legs. 2 new tests.
- [x] Write test: multicity search with PreferredAlliance filters out non-matching alliance flights
- [x] Add PreferredAlliance string to multicity.SearchParams
- [x] Map PreferredAlliance from search.Request in toSearchParams
- [x] Apply FilterByAlliance in filter stage (both legs + multi-city results)
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 66: Wire MaxPrice into multicity strategy
**Status:** done
**Plan:** Added MaxPrice to multicity.SearchParams, mapped from search.Request, filter on total combined itinerary price after COMBINE stage. 2 new tests.
- [x] Write test: multicity search with MaxPrice filters out itineraries exceeding budget
- [x] Add MaxPrice float64 to multicity.SearchParams
- [x] Map MaxPrice from search.Request in toSearchParams
- [x] Apply max price filtering on combined itineraries (total price, not per-flight)
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 67: Add departure time preference filter
**Status:** done
**Plan:** Added DepartureAfter/DepartureBefore to search.Request, FilterByDepartureTime in filter.go, wired into direct pipeline and chat tripParams. 9 new tests.
- [x] Add DepartureAfter and DepartureBefore string fields to search.Request
- [x] Write test: FilterByDepartureTime keeps flights within time range, removes outside
- [x] Implement FilterByDepartureTime in search/filter.go
- [x] Wire into direct search pipeline in direct.go
- [x] Add departure_after, departure_before to chat tripParams
- [x] Update buildRequestFromParams, mergeParams, parsePartialParams
- [x] Update chatSystemPrompt and refinementHint
- [x] Write chat tests for departure time param extraction and merge
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 68: Surface SeatsLeft in JSON output and ranker prompt
**Status:** done
**Plan:** Added seats_left to jsonLeg (omitempty), legSeatsLeft helper (min across segments), [Seats: N left] tag in ranker prompt. 6 new tests.
- [x] Write test: JSON output includes seats_left when SeatsLeft > 0
- [x] Add seats_left to jsonLeg (omitempty)
- [x] Add legSeatsLeft helper
- [x] Write test: ranker prompt includes [Seats: N left] tag when available
- [x] Add seats annotation in buildRankingPrompt
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 69: Build gate verification and session finalize
**Status:** done
**Plan:** Post-merge verification. All gates clean after gofmt fix in strategy.go.
- [x] Run gofmt -l . and fix violations (alignment in strategy.go)
- [x] Run go vet ./... -- clean
- [x] Run golangci-lint run -- 0 issues
- [x] Run go test ./... -- all pass
- [x] Update JOURNAL.md with session summary
- [x] Update LEARNINGS.md with new insights
- [x] Increment SESSION_NUMBER to 30
