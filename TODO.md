# TODO

Carried from: Day 21 (all completed)

## Tasks 31-35: Day 21 tasks
**Status:** completed (Day 21) -- alliance tags in ranker, stopover notes in ranker, direct-only chat, reasoning in output, lint sweep

---

## Task 36: Fix chat output routing to use io.Writer
**Status:** done
**Plan:** Add `io.Writer` parameter to `printTable`, `printJSON`, and `printJSONWithInsights`. Replace `os.Stdout` and bare `fmt.Print*` calls inside those functions with writes to the provided writer. Update `chatLoop` to pass `out`, `runSearch` to pass `os.Stdout`. Update tests to use writer directly (remove os.Pipe hacks). Write new test verifying chatLoop buffer captures table result data.
- [x] Write test: chatLoop output buffer contains result data (price, route) after search
- [x] Refactor printTable to accept io.Writer instead of hardcoding os.Stdout
- [x] Refactor printJSON and printJSONWithInsights to accept io.Writer
- [x] Update chatLoop to pass `out` writer to print functions
- [x] Update runSearch to pass os.Stdout explicitly
- [x] Update existing tests to use writer parameter directly
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 37: Show booking URLs in table output
**Status:** pending
**Plan:**
- [ ] Write test: table output for flight with BookingURL contains the URL
- [ ] Write test: flights without BookingURL show no extra content
- [ ] Add booking URL display to printTable (sub-row or column)
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 38: Conversation history truncation
**Status:** pending
**Plan:**
- [ ] Write test: history truncated after threshold while system prompt preserved
- [ ] Write test: recent messages remain after truncation
- [ ] Add truncateHistory function with configurable max messages
- [ ] Wire into chatLoop before each LLM call
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 39: Add stops count to table output
**Status:** pending
**Plan:**
- [ ] Write test: table output contains "STOPS" header
- [ ] Write test: correct stop counts for direct (0) and connecting (1+) flights
- [ ] Add stops helper function for itineraries
- [ ] Add "Stops" column to both single-leg and multi-leg table layouts
- [ ] Verify: `go build && go test ./... && go vet ./...`

## Task 40: Lint, gofmt sweep, and build gate verification
**Status:** pending
**Plan:**
- [ ] Run gofmt -l . and fix violations
- [ ] Run go vet ./... and fix warnings
- [ ] Run golangci-lint run and fix findings
- [ ] Run go test ./... and verify all pass
