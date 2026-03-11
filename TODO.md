# TODO

Carried from: Day 29 (all completed)

## Tasks 65-69: Day 29 tasks
**Status:** completed -- multicity alliance+maxprice consistency, departure time filter, seats display, build gate clean

---

## Task 70: Wire FilterByDepartureTime into multicity pipeline
**Status:** done
**Plan:** Add DepartureAfter/DepartureBefore to SearchParams, map in toSearchParams, apply FilterByDepartureTime in FILTER stage + stage 4b.
- [x] Write test: multicity search with DepartureAfter/DepartureBefore filters out flights outside time window
- [x] Add DepartureAfter and DepartureBefore string to multicity.SearchParams
- [x] Map DepartureAfter/DepartureBefore from search.Request in toSearchParams
- [x] Apply FilterByDepartureTime in FILTER stage for both leg1 and leg2
- [x] Apply FilterByDepartureTime in stage 4b for mcItineraries
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 71: Pass stops=0 to SerpAPI when direct-only
**Status:** done
**Plan:** When req.MaxStops==0, add stops=0 to the SerpAPI params map.
- [x] Write test: Search with MaxStops=0 includes stops=0 in request params
- [x] Write test: Search with MaxStops!=0 does not include stops param
- [x] Add stops=0 to params map when req.MaxStops == 0 in serpapi.Search()
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 72: Send return_date to SerpAPI for round-trip requests
**Status:** done
**Plan:** Add SerpAPIParamReturnDate constant, include return_date in params when IsRoundTrip().
- [x] Add SerpAPIParamReturnDate = "return_date" to config/routes.go
- [x] Update serpapi.Search() to include return_date when req.IsRoundTrip()
- [x] Update existing round-trip test to verify return_date param is sent
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 73: Fix empty city names in ranker prompt
**Status:** done
**Plan:** Conditionally include city parenthetical only when OriginCity or DestinationCity is non-empty.
- [x] Write test: buildRankingPrompt with empty OriginCity/DestinationCity produces clean output (no empty parens)
- [x] Modify ranker.go line ~221 to conditionally include city parenthetical
- [x] Verify existing ranker tests still pass
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 74: Build gate verification and session finalize
**Status:** done
**Plan:** Full verification, fix gofmt alignment, update governance files, increment session number.
- [x] Run gofmt -l . and fix any violations
- [x] Run go vet ./... -- must be clean
- [x] Run golangci-lint run -- must be 0 issues
- [x] Run go test ./... -- all must pass
- [x] Update JOURNAL.md with session summary
- [x] Increment SESSION_NUMBER
