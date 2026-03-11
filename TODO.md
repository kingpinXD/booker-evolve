# TODO

Carried from: Day 20 (all completed)

## Tasks 26-30: Day 20 tasks
**Status:** completed (Day 20) -- chat param merging, ranking profile, airline alliance data, richer result context, lint sweep

---

## Task 31: Alliance-aware ranking in multicity
**Status:** done
**Plan:** Wire search.Alliance() into buildRankingPrompt to add alliance tags next to airline names. Import search package in ranker.go. For each segment, call search.Alliance(seg.Airline) and append "[Star Alliance]" etc. Resolves TODO at combiner.go:33.
- [x] Write test: buildRankingPrompt output contains alliance tags for known members
- [x] Write test: unknown airlines show no alliance tag
- [x] Import search package in ranker.go (already imported)
- [x] Modify buildRankingPrompt to append alliance info per segment
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 32: Stopover city notes in ranker prompt
**Status:** done
**Plan:** Add Notes field to search.Stopover. Pass Notes from StopoverCity when building itinerary in combiner.go. Display Notes in buildRankingPrompt under stopover section.
- [x] Write test: buildRankingPrompt includes stopover notes when Notes is set
- [x] Write test: no extra output when Notes is empty
- [x] Add Notes string field to search.Stopover
- [x] Pass StopoverCity.Notes into Stopover in combiner.go buildItinerary
- [x] Display Notes in buildRankingPrompt stopover section
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 33: Direct-only preference in chat
**Status:** done
**Plan:** Add DirectOnly bool to tripParams. Wire to MaxStops=0 in buildRequestFromParams. Add to system prompt, refinement hint, and parsePartialParams recognition.
- [x] Write test: buildRequestFromParams with DirectOnly=true yields MaxStops=0
- [x] Write test: parsePartialParams recognizes {"direct_only":true}
- [x] Write test: mergeParams preserves DirectOnly
- [x] Add DirectOnly field to tripParams
- [x] Update buildRequestFromParams to set MaxStops=0 when DirectOnly
- [x] Update chatSystemPrompt and refinementHint
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 34: Surface ranking reasoning in output
**Status:** done
**Plan:** Add Reasoning to table and JSON output. Add "Reason" column to table for scored itineraries. Add "reasoning" field to jsonItinerary struct.
- [x] Write test: table output contains reasoning for scored itineraries
- [x] Write test: JSON output includes reasoning field
- [x] Add "reasoning" field to jsonItinerary
- [x] Add "Reason" column to table in printTable
- [x] Update buildJSONItineraries to include reasoning
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 35: Lint, gofmt sweep, and build gate verification
**Status:** done
**Plan:** Run all linting and testing tools. Fix any violations.
- [x] Run gofmt -l . and fix violations -- clean
- [x] Run go vet ./... and fix warnings -- clean
- [x] Run golangci-lint run and fix findings -- 0 issues
- [x] Run go test ./... and verify all pass -- all 15 packages pass
