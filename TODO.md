# TODO

## Task 175: Remove Book column from table output
**Status:** done
**Plan:** Remove "Book" header and legBookingURL data from printTable for both single-leg and multi-leg layouts. Keep BookingURL in JSON output.
- [x] Write test asserting table output does not contain "BOOK" column
- [x] Remove Book column from single-leg header and row in printTable
- [x] Remove Book column from multi-leg header and row in printTable
- [x] Verify legBookingURL is still used by formatOptionDetail (keep helper)
- [x] Run go build && go test ./... && go vet ./...

## Task 176: Bullet-point display for chat results
**Status:** done
**Plan:** Add printBulletResults in display.go for concise per-itinerary bullets. Change displayChatResults default from table to bullet format. Keep --format table/json as overrides.
- [x] Write tests for printBulletResults (single-leg, multi-leg, scored, empty)
- [x] Implement printBulletResults in display.go
- [x] Update displayChatResults to use bullet format by default
- [x] Add "bullet" format option, keep table/json as explicit choices
- [x] Run go build && go test ./... && go vet ./...

## Task 177: Truncate Reason column in table
**Status:** pending
**Plan:** Add truncateText helper, apply to Reason field in printTable row construction. Cap at 50 chars with "..." suffix.
- [ ] Write test for truncateText at boundary cases (49, 50, 51 chars)
- [ ] Implement truncateText helper in display.go
- [ ] Apply truncateText to Reason in both single-leg and multi-leg rows
- [ ] Run go build && go test ./... && go vet ./...

## Task 178: Shorten multi-leg table headers
**Status:** pending
**Plan:** Replace verbose "Leg 1 X" / "Leg 2 X" headers with "L1 X" / "L2 X" variants. Single-leg headers unchanged.
- [ ] Write test asserting multi-leg table has "L1" / "L2" prefixed headers
- [ ] Update multi-leg header row in printTable
- [ ] Update any tests that check for old header text
- [ ] Run go build && go test ./... && go vet ./...

## Task 179: North America to Europe stopover corridors
**Status:** pending
**Plan:** Add JFK→LHR, JFK→CDG, LAX→LHR transatlantic corridors using KEF, DUB as waypoints.
- [ ] Add JFK→LHR stopover corridor (KEF, DUB, YHZ)
- [ ] Add JFK→CDG stopover corridor (KEF, DUB, LHR)
- [ ] Add LAX→LHR stopover corridor (YVR, YYZ, KEF)
- [ ] Add specific lookup tests for new corridors
- [ ] Run go build && go test ./... && go vet ./...
