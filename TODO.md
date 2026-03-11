# TODO

Carried from: Day 37 (all completed)

## Tasks 95-99: Day 37 tasks
**Status:** completed -- refactor stage 4b filtering, total_trip JSON, departure time CLI flags, itinerary deduplication, stopover duration control

---

## Tasks 100-104: Day 38 tasks

## Task 100: Remove KiwiID from StopoverCity and SearchParams
**Status:** done
**Plan:** Remove KiwiID from StopoverCity struct, all KiwiID assignments in stopover entries, OriginKiwiID/DestinationKiwiID from SearchParams, KiwiID refs in fetch goroutines and diagnostic_test.go. Leave types.SearchRequest alone (Kiwi reads it).
- [x] Remove KiwiID field from StopoverCity struct
- [x] Remove all KiwiID assignments from stopover entries (DELToYYZ, BOMToYYZ, DELToYVR, GlobalFallbackHubs)
- [x] Remove OriginKiwiID/DestinationKiwiID from SearchParams
- [x] Remove KiwiID references in fetch goroutines (multicity.go)
- [x] Update diagnostic_test.go to remove KiwiID references
- [x] Verify build + test + vet pass

## Task 101: Simplify fetchWithDualSort to single fetch
**Status:** done
**Plan:** Replace fetchWithDualSort with direct fetchFromAllProviders call in Search(). Remove deduplicateFlights helper + fetchWithDualSort + their tests. Kiwi sort constants kept in config (Kiwi provider still reads them).
- [x] Replace fetchWithDualSort with direct fetchFromAllProviders call
- [x] Remove deduplicateFlights helper (no longer needed)
- [x] Remove fetchWithDualSort function
- [ ] ~~Remove Kiwi sort constants~~ (kept -- Kiwi provider reads them)
- [x] Remove sortCapturingProvider and fetchWithDualSort tests
- [x] Remove deduplicateFlights tests from helpers_test.go
- [x] Verify build + test + vet + lint pass

## Task 102: Add India-US route stopovers (DEL/BOM to JFK)
**Status:** done
**Plan:** Add curated stopover lists for DEL->JFK (8 cities) and BOM->JFK (7 cities), register in stopoversMap, write tests.
- [x] Write tests for StopoversForRoute("DEL","JFK") and StopoversForRoute("BOM","JFK")
- [x] Add DELToJFKStopovers with curated cities
- [x] Add BOMToJFKStopovers with curated cities
- [x] Register routes in stopoversMap
- [x] Verify tests pass

## Task 103: Consolidate stage 3 filter logging
**Status:** done
**Plan:** Extract applyBoth closure to reduce repetitive before/after counting. MaxStops handled separately (different params per leg).
- [x] Extract applyBoth helper closure for filter + drop counting
- [x] Replace verbose before/after counting with consolidated logic
- [x] Handle MaxStops separately (different per-leg params)
- [x] Verify existing multicity tests pass + build + vet + lint

## Task 104: No-results filter suggestion in chat
**Status:** done
**Plan:** Add filterSuggestion(tripParams) that checks active filters and suggests relaxing them. Wire into chatLoop.
- [x] Write tests for filterSuggestion with various active filters
- [x] Write test for filterSuggestion with no filters (empty result)
- [x] Implement filterSuggestion(params tripParams) string
- [x] Wire into chatLoop after no-results message
- [x] Verify build + test + vet pass
