# TODO

Carried from: Day 36 (all completed)

## Tasks 90-94: Day 36 tasks
**Status:** completed -- per-leg cabin columns, flight number JSON, fallback hub stopovers, combiner red-eye filter, multi-leg trip summary

---

## Tasks 95-99: Day 37 tasks
**Status:** completed -- refactor stage 4b filtering, total_trip JSON, departure time CLI flags, itinerary deduplication, stopover duration control

## Task 95: Refactor stage 4b multi-city filtering
**Status:** done
**Plan:** Extract passesAllFilters(f types.Flight, params SearchParams) bool helper to replace verbose single-element slice filtering in stage 4b. Files: search/multicity/multicity.go, search/multicity/helpers_test.go.
- [x] Write test for passesAllFilters helper (14 table-driven cases)
- [x] Extract helper function
- [x] Replace stage 4b code with helper calls (-56 lines)
- [x] Verify existing tests pass

## Task 96: Add total_trip to JSON output
**Status:** done
**Plan:** Add TotalTrip string field to jsonItinerary, populate from itin.TotalTrip using formatTripDuration. Files: cmd/search.go, cmd/search_test.go.
- [x] Write test verifying total_trip in JSON
- [x] Write test verifying total_trip omitted when zero
- [x] Add field to jsonItinerary struct
- [x] Populate in buildJSONItineraries
- [x] Verify existing tests pass

## Task 97: Wire departure time CLI flags
**Status:** done
**Plan:** Add --departure-after and --departure-before flags. Fields already on search.Request and wired in chat. Files: cmd/search.go.
- [x] Add key constants and flags
- [x] Wire into Request in runSearch
- [x] Verify build passes

## Task 98: Itinerary deduplication in multicity
**Status:** done
**Plan:** Add deduplicateItineraries helper after COMBINE+sort stage. Key by first segment flight number + departure time per leg. Keep cheapest. Files: search/multicity/multicity.go, search/multicity/helpers_test.go.
- [x] Write tests for deduplication (3 cases)
- [x] Implement deduplicateItineraries and itineraryKey
- [x] Wire into Search() after price sort
- [x] Verify existing tests pass

## Task 99: Stopover duration control via CLI and chat
**Status:** done
**Plan:** Add MinStopover/MaxStopover (time.Duration) to search.Request and multicity.SearchParams. Thread to CombineParams. Add CLI flags. Wire into chat tripParams. Files: search/strategy.go, search/multicity/strategy.go, search/multicity/multicity.go, cmd/search.go, cmd/chat.go.
- [x] Add fields to search.Request
- [x] Add fields to multicity.SearchParams and thread to CombineParams
- [x] Add CLI flags and wire to Request
- [x] Add chat params (parse, merge, build, prompt, hint)
- [x] Write tests: combiner override, chat parse/merge/build (7 test cases)
- [x] Verify all tests pass
