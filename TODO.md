# TODO

<<<<<<< Updated upstream
Carried from: Day 22 (all completed)

## Tasks 36-40: Day 22 tasks
**Status:** completed (Day 22) -- io.Writer output routing, booking URLs, history truncation, stops column, lint sweep

---

## Task 41: Flex-days support in chat
**Status:** done
**Plan:** Add flex_days field to tripParams, wire through buildRequestFromParams (default to 3), mergeParams, parsePartialParams, system prompt, refinement hint. TDD with 5 new tests.
- [x] Write test: parsePartialParams recognizes flex_days field
- [x] Write test: mergeParams preserves flex_days from prev when partial is zero
- [x] Write test: buildRequestFromParams uses flex_days when set, defaults to defaultFlexDays when zero
- [x] Add flex_days field to tripParams struct
- [x] Update chatSystemPrompt to document flex_days option
- [x] Update buildRequestFromParams to use p.FlexDays instead of hardcoded default
- [x] Update mergeParams to handle FlexDays
- [x] Update parsePartialParams to recognize FlexDays
- [x] Update refinementHint to mention flex-days
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 42: Add today's date to chat system prompt
**Status:** done
**Plan:** Change chatSystemPrompt() to accept time.Time, prepend "Today's date is YYYY-MM-DD" to prompt, update chatLoop caller to pass time.Now(). Update existing tests to pass known dates.
- [x] Write test: chatSystemPrompt output contains a YYYY-MM-DD date string
- [x] Write test: verify format with known date
- [x] Make chatSystemPrompt accept a date parameter
- [x] Inject current date into system prompt text
- [x] Update chatLoop to pass current date
- [x] Update existing chat tests if signature changes
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 43: Show layover durations in table output
**Status:** done
**Plan:** Add formatStops(itin) string that sums stops and layover durations. When stops > 0 and layover data exists, format as "N (Xh Ym)". Replace itineraryStops int in printTable with formatStops string. TDD with 3 new tests.
- [x] Write test: formatStops returns "0" for direct flights
- [x] Write test: formatStops returns "1 (2h 30m)" for connecting flights with layover
- [x] Write test: formatStops returns "1" for connecting flights without layover data
- [x] Add formatStops helper that includes layover time
- [x] Replace raw itineraryStops int with formatted string in printTable
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 44: Show arrival time in table output
**Status:** done
**Plan:** Add legArrival(itin, legIdx) following legDeparture pattern. Add "Arrival" column to single-leg table and "Leg N Arrival" columns to multi-leg table. TDD with 2 new tests + 2 updated tests.
- [x] Write test: legArrival returns correct time from last segment
- [x] Write test: legArrival returns empty for out-of-bounds leg
- [x] Write test: table output contains arrival time strings
- [x] Add legArrival helper following legDeparture pattern
- [x] Add "Arrival" column to both single-leg and multi-leg table layouts
- [x] Verify: `go build && go test ./... && go vet ./...`

## Task 45: Lint, gofmt sweep, and build gate verification
**Status:** done
**Plan:** Run gofmt, go vet, golangci-lint, go test on full codebase after rebasing all worktree branches.
- [x] Run gofmt -l . and fix any violations -- clean
- [x] Run go vet ./... and fix any warnings -- clean
- [x] Run golangci-lint run and fix any issues -- 0 issues
- [x] Run go test ./... and verify all pass -- 15 packages pass
=======
## Task 1: Add booking URL to table and JSON output
**Status:** pending
**Plan:**
- [ ] Add legBookingURL helper to extract URL from itinerary leg
- [ ] Write tests for legBookingURL and updated JSON output
- [ ] Add booking_url to jsonLeg and jsonItinerary structs
- [ ] Add Book column to table output (both single-leg and multi-leg)
- [ ] Run go build && go test ./cmd/... && go vet ./...

## Task 2: Airline alliance data and same-alliance scoring
**Status:** pending
**Plan:**
- [ ] Write tests for Alliance lookup and SameAlliance function
- [ ] Add airline alliance map (Star Alliance, OneWorld, SkyTeam members)
- [ ] Add Alliance(iataCode string) string function
- [ ] Add SameAlliance(a, b Flight) bool function using alliances
- [ ] Run go build && go test ./search/multicity/... && go vet ./...

## Task 3: Expand stopover routes beyond DEL-YYZ
**Status:** pending
**Plan:**
- [ ] Write tests for new routes (BOM→YYZ, DEL→YVR, reverse directions)
- [ ] Add BOMToYYZStopovers list
- [ ] Add DELToYVRStopovers list
- [ ] Update StopoversForRoute with new route lookups and reverse support
- [ ] Run go build && go test ./search/multicity/... && go vet ./...

## Task 4: Refactor printTable to accept io.Writer and add tests
**Status:** pending
**Plan:**
- [ ] Change printTable signature to accept io.Writer
- [ ] Update runSearch caller to pass os.Stdout
- [ ] Write tests for single-leg and multi-leg table output
- [ ] Run go build && go test ./cmd/... && go vet ./...

## Task 5: Add seats-left indicator to output
**Status:** pending
**Plan:**
- [ ] Write tests for seatsString helper
- [ ] Add seatsString(itin, legIdx) helper
- [ ] Add seats_left to jsonLeg struct
- [ ] Add Seats column to table output (both layouts)
- [ ] Run go build && go test ./cmd/... && go vet ./...
>>>>>>> Stashed changes
