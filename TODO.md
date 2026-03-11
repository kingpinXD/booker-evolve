# TODO

Carried from: Day 31 (all completed)

## Tasks 70-74: Day 31 tasks
**Status:** completed -- multicity departure time filter, SerpAPI stops+return_date params, ranker city name cleanup, build gate clean

---

## Task 75: Post-fetch sorting in direct strategy + CLI flag
**Status:** done
**Plan:** Add SortBy to search.Request. Implement SortResults in filter.go (price/duration/departure). Wire into direct.go after ranker. Add --sort-by CLI flag. Files: search/strategy.go, search/filter.go, search/filter_test.go, search/direct/direct.go, cmd/search.go.
- [ ] Write test: SortResults sorts by price ascending (default)
- [ ] Write test: SortResults sorts by duration ascending
- [ ] Write test: SortResults sorts by departure time ascending
- [ ] Write test: empty/unknown sort mode returns original order
- [ ] Add SortBy string to search.Request
- [ ] Implement SortResults(itineraries []Itinerary, sortBy string) in filter.go
- [ ] Wire SortResults after filters in direct.go Search()
- [ ] Add --sort-by CLI flag to search command
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 76: Connection risk tags in ranker prompt
**Status:** pending
**Plan:** [to be filled during implementation]
- [ ] Write test: buildRankingPrompt with 45min layover shows "[Risky connection: 45m]"
- [ ] Write test: buildRankingPrompt with 75min layover shows "[Tight connection: 75m]"
- [ ] Write test: buildRankingPrompt with 120min layover shows no connection tag
- [ ] Add connection risk tagging after layover line in buildRankingPrompt
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 77: Wire sort_by into chat conversation
**Status:** pending
**Plan:** [to be filled during implementation]
- [ ] Add SortBy string to tripParams
- [ ] Wire sort_by into parsePartialParams
- [ ] Wire sort_by into mergeParams
- [ ] Wire sort_by into buildRequestFromParams
- [ ] Add sort_by to system prompt and refinement hint
- [ ] Write test: parsePartialParams with sort_by field
- [ ] Write test: mergeParams preserves and overrides sort_by
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 78: Enrich JSON output with airline codes and city names
**Status:** pending
**Plan:** [to be filled during implementation]
- [ ] Add AirlineCode, OriginCity, DestinationCity, OriginName, DestinationName to jsonLeg (omitempty)
- [ ] Populate new fields in buildJSONItineraries from first segment
- [ ] Write test: JSON output includes new fields when segment data is available
- [ ] Write test: new fields omitted when segment data is empty
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 79: Avoid airline filter
**Status:** in-progress
**Plan:** Add AvoidAirlines string to search.Request. Implement FilterByAvoidAirlines in filter.go (splits comma-separated codes, checks Airline and OperatingCarrier). Wire into direct.go filter pipeline and multicity.go. Add --avoid-airlines CLI flag. Files: search/strategy.go, search/filter.go, search/filter_test.go, search/direct/direct.go, cmd/search.go, search/multicity/multicity.go.
- [ ] Write test: FilterByAvoidAirlines removes flight matching airline code
- [ ] Write test: FilterByAvoidAirlines removes flight matching operating carrier
- [ ] Write test: FilterByAvoidAirlines passes through non-matching flights
- [ ] Add AvoidAirlines string to search.Request
- [ ] Implement FilterByAvoidAirlines in filter.go
- [ ] Wire into direct pipeline in direct.go
- [ ] Wire into multicity SearchParams and multicity.go
- [ ] Add --avoid-airlines CLI flag
- [ ] Verify: `go build && go test ./... && go vet ./...`
