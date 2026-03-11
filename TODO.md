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
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Replace fetchWithDualSort with direct fetchFromAllProviders call
- [ ] Remove deduplicateFlights helper (no longer needed)
- [ ] Remove KiwiSortByQuality, KiwiSortByPrice, KiwiSortAscending constants from config/routes.go
- [ ] Remove KiwiParamSortBy, KiwiParamSortOrder from config/routes.go
- [ ] Update sortCapturingProvider test helper in helpers_test.go
- [ ] Remove deduplicateFlights tests from helpers_test.go
- [ ] Verify build + test + vet pass

## Task 102: Add India-US route stopovers (DEL/BOM to JFK)
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Write tests for StopoversForRoute("DEL","JFK") and StopoversForRoute("BOM","JFK")
- [ ] Add DELToJFKStopovers with curated cities
- [ ] Add BOMToJFKStopovers with curated cities
- [ ] Register routes in stopoversMap
- [ ] Verify tests pass

## Task 103: Consolidate stage 3 filter logging
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Extract applyFilter helper or use compact loop pattern
- [ ] Replace verbose before/after counting with consolidated logic
- [ ] Verify existing multicity tests pass
- [ ] Verify build + vet pass

## Task 104: No-results filter suggestion in chat
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Write tests for filterSuggestion with various active filters
- [ ] Write test for filterSuggestion with no filters (empty result)
- [ ] Implement filterSuggestion(params tripParams) string
- [ ] Wire into chatLoop after no-results message
- [ ] Verify build + test + vet pass
