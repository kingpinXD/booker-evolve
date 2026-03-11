# TODO

Carried from: Day 18 (all completed)

## Task 16-20: Day 18 tasks
**Status:** completed (Day 18) — nearby strategy, round-trip, infra refactor, refinement hints, lint sweep

---

## Task 21: Add --return-date flag to search command
**Status:** done
**Plan:** Add keyReturnDate const, register --return-date flag in init(), wire to req.ReturnDate in runSearch.
- [x] Add --return-date string flag to searchCmd in init()
- [x] Wire viper.GetString("return-date") into req.ReturnDate in runSearch
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 22: Wire NearbySearcher into buildPicker
**Status:** done
**Plan:** Import nearby package, wrap directStrategy with nearby.NewSearcher, pass as 3rd strategy to NewPicker.
- [x] Import search/nearby in cmd/infra.go
- [x] Create NearbySearcher wrapping directStrategy in buildPicker
- [x] Register nearbyStrategy as third picker strategy
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 23: Add --max-price budget filter
**Status:** done
**Plan:** Add MaxPrice to Request, FilterByMaxPrice filter, wire into direct pipeline, CLI flag, and chat tripParams. TDD: test filter first, then implement.
- [x] Add MaxPrice float64 to search.Request
- [x] Write FilterByMaxPrice test (TDD: red then green)
- [x] Implement FilterByMaxPrice
- [x] Wire into direct.searchFlights
- [x] Add --max-price CLI flag + chat integration (incl. system prompt)
- [x] Fix gofmt alignment
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 24: Surface PriceInsights in output
**Status:** done
**Plan:** Return raw *serpapi.Provider from buildPicker, write formatPriceInsights TDD, display after search in table + JSON.
- [x] Modify buildPicker to also return *serpapi.Provider
- [x] Write formatPriceInsights test (TDD) + implement
- [x] Add printJSONWithInsights with price_insights field
- [x] Refactor printJSON to use shared buildJSONItineraries
- [x] Wire into runSearch (table + JSON output)
- [x] Fix gofmt alignment
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 25: Test multicity.Strategy.Search
**Status:** done
**Plan:** Used existing test helpers (newTestSearcher, validLeg1/2, llmRankingHandler) to test Strategy.Search happy path and error propagation.
- [x] Write TestStrategy_Search with mock Searcher (happy path)
- [x] Write TestStrategy_Search_Error (invalid date propagation)
- [x] Verify: `go test ./search/multicity/... -race`
