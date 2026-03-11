# TODO

Carried from: Day 24 (all completed)

## Tasks 46-49: Day 24 tasks
**Status:** completed (Day 24) -- cabin class display, carbon emissions, red-eye detection, lint sweep

---

## Task 50: Parse overnight flag and annotate in ranker prompt
**Status:** done
**Plan:** Parse Overnight bool from SerpAPI FlightSegment into types.Segment.Overnight. Add [Overnight] tag in buildRankingPrompt for overnight connections (similar to [Red-eye] tag). Improves ranking quality for itineraries with overnight layovers.
- [x] Add Overnight bool to types.Segment
- [x] Parse Overnight from SerpAPI FlightSegment in parser.go
- [x] Write test: parser extracts Overnight flag
- [x] Write test: buildRankingPrompt includes [Overnight] tag
- [x] Add [Overnight] annotation in buildRankingPrompt
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 51: Parse aircraft type from SerpAPI and display in JSON output
**Status:** done
**Plan:** Parse Airplane field from SerpAPI FlightSegment into types.Segment.Aircraft. Display in JSON output as aircraft field. Helps users compare aircraft comfort on long-haul flights.
- [x] Add Aircraft string to types.Segment
- [x] Parse Airplane from SerpAPI in parser.go
- [x] Write test: parser extracts aircraft type
- [x] Add aircraft field to jsonLeg struct in cmd/search.go
- [x] Write test: JSON output includes aircraft field
- [x] Write test: JSON omits aircraft when empty
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 52: Conditional Score/Reason columns in table output
**Status:** done
**Plan:** When all scores are 0 (no ranker), hide Score and Reason columns to reduce noise in direct search output. Also fix carbon emissions integer division bug (grams/1000 truncates small values).
- [x] Write test: table output hides Score/Reason when all scores 0
- [x] Write test: table output shows Score/Reason when scores exist
- [x] Add hasScores helper function
- [x] Conditionally include Score/Reason in table headers and rows
- [x] Fix carbon emissions integer division: use rounding
- [x] Write test: carbon emissions correctly converts small gram values
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 53: Lint, gofmt sweep, and build gate verification
**Status:** done
**Plan:** Final validation pass.
- [x] Run gofmt -l . and fix any violations -- clean
- [x] Run go vet ./... and fix any warnings -- clean
- [x] Run golangci-lint run and fix any issues -- 0 issues
- [x] Run go test ./... and verify all pass -- 15 packages pass
