# TODO

Carried from: Day 26 (all completed)

## Tasks 50-58: Days 25-26 tasks
**Status:** completed -- overnight flag, aircraft type, conditional Score/Reason, carbon rounding, legroom, JSON arrival/stops/score, PriceInsights in chat, multi-leg CO2, lint sweep

---

## Task 59: Add aircraft and carbon annotations to ranker prompt
**Status:** done
**Plan:** Added [Aircraft: X] tag per segment and CO2: Xkg line per leg in buildRankingPrompt. 4 new tests.
- [x] Write test: buildRankingPrompt includes [Aircraft: X] tag when segment has aircraft
- [x] Write test: buildRankingPrompt omits aircraft tag when empty
- [x] Add `[Aircraft: X]` tag in buildRankingPrompt per segment (when non-empty)
- [x] Write test: buildRankingPrompt includes CO2 line when CarbonKg > 0
- [x] Write test: buildRankingPrompt omits CO2 line when CarbonKg == 0
- [x] Add `CO2: Xkg` line in buildRankingPrompt per leg (when CarbonKg > 0)
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 60: Parse carbon benchmark data from SerpAPI, surface in JSON and ranker
**Status:** done
**Plan:** Threaded TypicalForThisRoute and DifferencePercent end-to-end. 6 new tests.
- [x] Add TypicalCarbonKg (int) and CarbonDiffPct (int) to types.Flight
- [x] Write test: parser extracts TypicalCarbonKg with rounding
- [x] Write test: parser extracts CarbonDiffPct
- [x] Parse TypicalForThisRoute and DifferencePercent from SerpAPI in parser.go
- [x] Write test: JSON output includes typical_carbon_kg and carbon_diff_percent
- [x] Write test: JSON omits fields when zero
- [x] Add typical_carbon_kg and carbon_diff_percent to jsonLeg struct (omitempty)
- [x] Wire into buildJSONItineraries
- [x] Write test: ranker CO2 line shows benchmark comparison
- [x] Enhance ranker CO2 line to "CO2: Xkg (Y% vs typical)" when benchmark available
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 61: Lint, gofmt sweep, and build gate verification
**Status:** done
**Plan:** Two gofmt fixes (types.go, search.go). All gates clean.
- [x] Run gofmt -l . and fix any violations
- [x] Run go vet ./... and fix any warnings
- [x] Run golangci-lint run and fix any issues
- [x] Run go test ./... and verify all pass
