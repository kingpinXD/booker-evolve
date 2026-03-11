# TODO

Carried from: Day 35 (all completed)

## Tasks 85-89: Day 35 tasks
**Status:** completed -- next-day arrival, codeshare display, richer chat summary, preferred airlines filter, ranker LLM caching

---

## Tasks 90-94: Day 36 tasks
**Status:** completed -- per-leg cabin columns, flight number JSON, fallback hub stopovers, combiner red-eye filter, multi-leg trip summary

## Task 90: Multi-leg per-leg cabin class columns
**Status:** done
**Plan:** Fix multi-leg table to show "Leg 1 Cabin" / "Leg 2 Cabin" instead of single "Cabin" showing only leg 0. Same pattern as CO2 per-leg fix. Files: cmd/search.go, cmd/search_test.go.
- [x] Write test verifying multi-leg table has two cabin columns
- [x] Change header from "Cabin" to "Leg 1 Cabin", "Leg 2 Cabin"
- [x] Change row from legCabin(itin, 0) to legCabin(itin, 0), legCabin(itin, 1)
- [x] Verify single-leg table still has single "Cabin" column
- [x] Verify existing tests still pass

## Task 91: Flight number in JSON output
**Status:** done
**Plan:** Add flight_number to jsonLeg (first segment's FlightNumber). Original plan was bags display but BagsIncluded only populated by inactive Kiwi provider. Files: cmd/search.go, cmd/search_test.go.
- [x] Write test for flight_number in JSON output
- [x] Write test for flight_number omitempty when empty
- [x] Add FlightNumber to jsonLeg struct
- [x] Populate in buildJSONItineraries
- [x] Verify existing tests still pass

## Task 92: Fallback global hub stopovers
**Status:** done
**Plan:** Add GlobalFallbackHubs slice (8 well-connected hubs). Modify StopoversForRoute to return filtered fallback hubs when route-specific stopovers are nil. Filter out origin/destination airports. Files: search/multicity/stopovers.go, search/multicity/stopovers_test.go, search/multicity/search_test.go.
- [x] Write test: known routes return specific stopovers (unchanged)
- [x] Write test: unknown route returns fallback hubs
- [x] Write test: fallback hubs exclude origin and destination airports
- [x] Define GlobalFallbackHubs slice
- [x] Modify StopoversForRoute to return filtered fallbacks
- [x] Update search_test.go for new behavior (unknown routes now use fallbacks)
- [x] Verify existing tests still pass

## Task 93: Combiner red-eye leg filtering
**Status:** done
**Plan:** Reuse existing isRedEye from ranker.go (same package). Skip combinations where leg2 departure is 00:00-04:59. Files: search/multicity/combiner.go, search/multicity/combiner_test.go.
- [x] Write test: red-eye leg2 departure (02:00) rejected
- [x] Write test: normal leg2 departure (10:00) passes
- [x] Write test: edge cases (05:00 OK, 04:59 rejected, 00:00 rejected)
- [x] Add red-eye check in CombineLegs loop (reuses ranker's isRedEye)
- [x] Update TODO comment in combiner.go
- [x] Verify existing combiner tests still pass

## Task 94: Multi-leg trip summary footer
**Status:** done
**Plan:** Enhance priceSummary to show total trip duration range for multi-leg itineraries. Files: cmd/search.go, cmd/search_test.go.
- [x] Write test: multi-leg priceSummary includes duration range
- [x] Write test: single-leg priceSummary unchanged
- [x] Write test: single result shows duration without range
- [x] Add formatTripDuration helper (Xd Yh format)
- [x] Extend priceSummary with TotalTrip range for multi-leg
- [x] Verify existing tests still pass
