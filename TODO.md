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
**Status:** done
**Plan:** Add "Book" column to both single-leg and multi-leg table layouts. Add legBookingURL helper.
- [x] Write test: table output for flight with BookingURL contains the URL
- [x] Write test: flights without BookingURL show no extra content
- [x] Write test: multi-leg table with BookingURL
- [x] Add legBookingURL helper + "Book" column to printTable
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 38: Conversation history truncation
**Status:** done
**Plan:** Add truncateHistory sliding window (system prompt + last 20 messages). Wire into chatLoop before ChatCompletion call.
- [x] Write test: history truncated after threshold while system prompt preserved
- [x] Write test: short history unchanged
- [x] Write test: chatLoop integration with history truncation
- [x] Add truncateHistory function + maxHistoryMessages const
- [x] Wire into chatLoop before each LLM call
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 39: Add stops count to table output
**Status:** done
**Plan:** Add `itineraryStops` helper that sums Flight.Stops() across all legs. Add "Stops" column to both single-leg and multi-leg layouts in printTable. TDD with tests for 0-stop direct and 1-stop connecting flights.
- [x] Write test: table output contains "STOPS" header and correct counts
- [x] Add itineraryStops helper + "Stops" column to printTable
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 40: Lint, gofmt sweep, and build gate verification
**Status:** done
**Plan:** Run all gates, fix any violations from worktree merges or new code.
- [x] Run gofmt -l . -- clean (0 violations)
- [x] Run go vet ./... -- clean
- [x] Run golangci-lint run -- 0 issues
- [x] Run go test ./... -- all pass
