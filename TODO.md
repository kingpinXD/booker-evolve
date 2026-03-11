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
**Status:** pending
**Plan:**
- [ ] Add MaxPrice float64 field to search.Request
- [ ] Write FilterByMaxPrice test in search/filter_test.go (TDD: verify fail)
- [ ] Implement FilterByMaxPrice in search/filter.go
- [ ] Call FilterByMaxPrice in direct.searchFlights pipeline
- [ ] Add --max-price flag to searchCmd, wire to req.MaxPrice
- [ ] Add max_price to tripParams in chat.go, wire to buildRequestFromParams
- [ ] Update chatSystemPrompt to mention max_price/budget option
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 24: Surface PriceInsights in output
**Status:** pending
**Plan:**
- [ ] Modify buildPicker to also return *serpapi.Provider reference
- [ ] Write formatPriceInsights helper test (TDD)
- [ ] Implement formatPriceInsights(insights search.PriceInsights) string
- [ ] Call LastPriceInsights() in runSearch after strategy.Search, display below price summary
- [ ] Add price_insights to JSON output
- [ ] Wire into chatLoop for chat output
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 25: Test multicity.Strategy.Search
**Status:** pending
**Plan:**
- [ ] Write test with mock Searcher that records calls
- [ ] Verify Search delegates to Searcher.Search with correct params from toSearchParams
- [ ] Verify: `go test ./search/multicity/... -race`
