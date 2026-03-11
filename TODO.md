# TODO

Carried from: Day 41 (all completed)

## Tasks 110-114: Day 41 tasks
**Status:** completed -- refactor ranker sort + applyScores, bidirectional route lookup, India-US West Coast stopovers, zero-results chat suggestions, stopover data consistency test

---

## Task 115: Update stale Kiwi references in multicity pipeline docs
**Status:** done
**Plan:** Update package-level comments in multicity.go to reference SerpAPI instead of Kiwi.
- [x] Replace all "Kiwi" references in doc comments with SerpAPI equivalents
- [x] Update TODOs that reference Kiwi-specific limitations
- [x] Verify build + vet pass

## Task 116: Add India-Australia stopover routes (DEL/BOM to SYD)
**Status:** pending
**Plan:** Add curated stopover lists for DEL->SYD and BOM->SYD corridors.
- [ ] Write tests for StopoversForRoute("DEL","SYD") and StopoversForRoute("BOM","SYD")
- [ ] Add DELToSYDStopovers (~6 cities: SIN, BKK, KUL, HKG, NRT, KIX)
- [ ] Add BOMToSYDStopovers (~5 cities: SIN, BKK, KUL, HKG, NRT)
- [ ] Register routes in stopoversMap
- [ ] Verify tests pass + consistency test passes

## Task 117: Add eco ranking profile (carbon-weighted)
**Status:** pending
**Plan:** Add WeightsEco profile, register in profiles map, wire into chat.
- [ ] Write test for WeightsEco profile (weights sum to 100, carbon-aware)
- [ ] Add WeightsEco to ranker.go
- [ ] Register "eco" in profiles map in cmd/search.go
- [ ] Update chat system prompt and refinement hint to mention eco profile
- [ ] Add chat test recognizing eco profile
- [ ] Verify build + test + vet pass

## Task 118: Add Indian airport clusters (DEL, BOM metro areas)
**Status:** pending
**Plan:** Add DEL and BOM clusters to airports.go for NearbySearcher expansion.
- [ ] Write tests: NearbyAirports("DEL") returns ["JAI"], NearbyAirports("BOM") returns ["PNQ"]
- [ ] Add "Delhi" and "Mumbai" clusters to airportClusters map
- [ ] Verify tests pass

## Task 119: Extract display formatting from cmd/search.go into cmd/display.go
**Status:** pending
**Plan:** Move all display/formatting functions and JSON types from search.go to display.go.
- [ ] Create cmd/display.go with display functions and types
- [ ] Remove moved functions/types from cmd/search.go
- [ ] Verify build passes (same package, no import changes needed)
- [ ] Verify all cmd tests pass unchanged
