# TODO

Carried from: Day 23 (all completed)

## Tasks 41-45: Day 23 tasks
**Status:** completed (Day 23) -- flex-days in chat, date in system prompt, layover durations, arrival time column, lint sweep

---

## Task 46: Display cabin class in table and JSON output
**Status:** done
**Plan:** Segment.CabinClass is populated by SerpAPI but never shown. Add a "Cabin" column to table output and a cabin_class field to JSON output. Helps users verify they're seeing results for the correct cabin.
- [x] Write test: legCabin returns correct cabin string
- [x] Write test: table output contains cabin class string
- [x] Write test: JSON output contains cabin_class field
- [x] Add legCabin helper following legAirlines pattern
- [x] Add "Cabin" column to single-leg and multi-leg table layouts
- [x] Add cabin_class to jsonLeg struct and buildJSONItineraries
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 47: Parse and display carbon emissions from SerpAPI
**Status:** done
**Plan:** SerpAPI returns carbon_emissions per flight group (this_flight grams, typical_for_this_route, difference_percent). Add CarbonEmissions to response.go FlightGroup, parse in parser.go into types.Flight.CarbonKg, display as "CO2" column in table and carbon_emissions in JSON.
- [x] Add CarbonEmissions struct to response.go
- [x] Add CarbonKg field to types.Flight
- [x] Write test: parser extracts carbon emissions from response
- [x] Write test: legCarbon returns formatted kg string
- [x] Update parseFlightGroup to extract carbon emissions
- [x] Add legCarbon helper to cmd/search.go
- [x] Add "CO2" column to table layouts and carbon_kg to jsonLeg
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 48: Add red-eye detection to ranker prompt
**Status:** done
**Plan:** The ranker mentions "reasonable departure times" but doesn't flag red-eye flights. Add isRedEye helper detecting departures 00:00-05:00, annotate with [Red-eye] in buildRankingPrompt.
- [x] Write test: isRedEye returns true for 2:30 AM departure
- [x] Write test: isRedEye returns false for 10:00 AM departure
- [x] Write test: buildRankingPrompt includes [Red-eye] tag for late-night flights
- [x] Add isRedEye helper in search/multicity/ranker.go
- [x] Annotate red-eye flights in buildRankingPrompt
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 49: Lint, gofmt sweep, and build gate verification
**Status:** done
**Plan:** Final validation pass.
- [x] Run gofmt -l . and fix any violations -- one fix in response.go
- [x] Run go vet ./... and fix any warnings -- clean
- [x] Run golangci-lint run and fix any issues -- 0 issues
- [x] Run go test ./... and verify all pass -- 15 packages pass
