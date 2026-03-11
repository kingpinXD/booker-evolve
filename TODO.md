# TODO

Carried from: Day 25 (all completed)

## Tasks 50-53: Day 25 tasks
**Status:** completed (Day 25) -- overnight flag, aircraft type, conditional Score/Reason, carbon rounding, lint sweep

---

## Task 54: Parse legroom from SerpAPI and annotate in ranker
**Status:** done
**Plan:**
- [ ] Add Legroom string to types.Segment
- [ ] Parse Legroom from SerpAPI FlightSegment.Legroom in parser.go
- [ ] Write test: parser extracts legroom string
- [ ] Add legroom field to jsonLeg struct in cmd/search.go
- [ ] Wire legroom into buildJSONItineraries via new legLegroom helper
- [ ] Write test: JSON output includes legroom field
- [ ] Write test: JSON omits legroom when empty
- [ ] Add [Legroom: Xin] annotation in buildRankingPrompt
- [ ] Write test: buildRankingPrompt includes legroom tag
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 55: Enrich JSON output with arrival time, stops, and omit zero score
**Status:** done
**Plan:** Added Arrival (RFC3339) and Stops fields to jsonLeg, added omitempty to Score in jsonItinerary, wired into buildJSONItineraries.
- [x] Add Arrival string field to jsonLeg struct
- [x] Wire arrival time into buildJSONItineraries (last segment ArrivalTime, RFC3339)
- [x] Add Stops int field to jsonLeg struct
- [x] Wire stops into buildJSONItineraries using Flight.Stops()
- [x] Add omitempty to Score field in jsonItinerary
- [x] Write test: JSON output includes arrival and stops fields
- [x] Write test: JSON Score omitted when 0
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 56: Wire PriceInsights into chat output
**Status:** done
**Plan:**
- [ ] Update chatLoop signature to accept a PriceInsightsProvider interface or callback
- [ ] Thread rawProvider from runChat into chatLoop
- [ ] After printing results in chatLoop, call formatPriceInsights and display if non-empty
- [ ] Write test: chat output includes price insights after results
- [ ] Update existing chat tests for new chatLoop signature
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 57: Fix multi-leg CO2 display
**Status:** done
**Plan:** Replaced single "CO2" column with "Leg 1 CO2" and "Leg 2 CO2" in multi-leg layout header and rows.
- [x] Replace single "CO2" column with "Leg 1 CO2" and "Leg 2 CO2" in multi-leg table header
- [x] Wire legCarbon(itin, 0) and legCarbon(itin, 1) into multi-leg table rows
- [x] Write test: multi-leg table output shows CO2 for both legs
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 58: Lint, gofmt sweep, and build gate verification
**Status:** pending
**Plan:**
- [ ] Run gofmt -l . and fix any violations
- [ ] Run go vet ./... and fix any warnings
- [ ] Run golangci-lint run and fix any issues
- [ ] Run go test ./... and verify all pass
