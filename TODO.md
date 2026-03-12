# TODO

## Task 180: Refactor stopovers.go with newStopover helper
**Status:** pending
**Plan:** Add newStopover helper that applies default MinStay/MaxStay, replace all 209 entries with single-line calls.
- [ ] Add newStopover(city, airport, region, notes) helper function
- [ ] Replace all StopoverCity literal entries with newStopover calls
- [ ] Verify file compiles and all tests pass (TestStopoverDataConsistency)
- [ ] Run go build ./... && go test ./... && go vet ./...

## Task 181: Auto-retry with relaxed filters on zero results
**Status:** done
**Plan:** Add relaxFilters(tripParams) (tripParams, string) to chathelpers.go. Relaxation priority: direct_only -> preferred_alliance -> preferred_airlines -> max_price (50% increase) -> departure/arrival time -> max_duration. Wire into chatLoop zero-results block: if filters active, relax and retry chatSearch once. Files: chathelpers.go, chat.go, chat_test.go.
- [x] Write tests for relaxFilters helper (direct_only, alliance, max_price, time constraints)
- [x] Implement relaxFilters(tripParams) (tripParams, string) in chathelpers.go
- [x] Wire auto-retry into chatLoop zero-results block
- [x] Write integration test for chatLoop auto-retry behavior
- [x] Run go build ./... && go test ./... && go vet ./...
- [ ] Write tests for relaxFilters helper (direct_only, alliance, max_price, time constraints)
- [ ] Implement relaxFilters(tripParams) (tripParams, string) in chathelpers.go
- [ ] Wire auto-retry into chatLoop zero-results block
- [ ] Write integration test for chatLoop auto-retry behavior
- [ ] Run go build ./... && go test ./... && go vet ./...

## Task 182: Enrich bullet output with departure date and per-leg price
**Status:** pending
**Plan:** Add departure date, per-leg price, and cabin class to bullet output format.
- [ ] Write tests for bullet output with departure date, multi-leg per-leg prices, cabin class
- [ ] Add departure date to single-leg and multi-leg bullet lines
- [ ] Add per-leg price to multi-leg sub-bullets
- [ ] Show cabin class when non-economy
- [ ] Run go build ./... && go test ./... && go vet ./...

## Task 183: Search parameter echo in chat
**Status:** done
**Plan:** Add formatSearchParams(tripParams) string to chathelpers.go. Format: "Searching DEL -> YYZ on 2025-06-15 (economy, flex +/-3 days, max $1200)". Wire into chatSearch to replace the current simple "Searching X -> Y on Z..." line. Files: chathelpers.go, chat_test.go.
- [x] Write tests for formatSearchParams with various param combinations
- [x] Implement formatSearchParams(tripParams) string in chathelpers.go
- [x] Wire into chatSearch to replace simple output line
- [x] Run go build ./... && go test ./... && go vet ./...

## Task 184: Context-aware refinement prompt
**Status:** done
**Plan:** Add refinementSuggestion(results, params, pi) string to chathelpers.go. Generates context-aware suggestions: "direct_only is limiting results", "cheapest date is Jun 12", "prices look high, try flex_days". Replace static "Want to refine?" in chatLoop. Files: chathelpers.go, chat.go, chat_test.go.
- [x] Write tests for refinementSuggestion with various scenarios
- [x] Implement refinementSuggestion(results, params, PriceInsights) string in chathelpers.go
- [x] Replace static prompt in chatLoop with refinementSuggestion output
- [x] Run go build ./... && go test ./... && go vet ./...
