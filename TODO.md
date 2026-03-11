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
**Status:** done
**Plan:** Add curated stopover lists for DEL->SYD and BOM->SYD corridors.
- [x] Write tests for StopoversForRoute("DEL","SYD") and StopoversForRoute("BOM","SYD")
- [x] Add DELToSYDStopovers (6 cities: SIN, BKK, KUL, HKG, NRT, KIX)
- [x] Add BOMToSYDStopovers (5 cities: SIN, BKK, KUL, HKG, NRT)
- [x] Register routes in stopoversMap
- [x] Verify tests pass + consistency test passes

## Task 117: Add eco ranking profile (carbon-weighted)
**Status:** done
**Plan:** Add WeightsEco profile, register in profiles map, wire into chat.
- [x] Write test for WeightsEco profile (weights sum to 100, carbon-aware)
- [x] Add WeightsEco to ranker.go
- [x] Register "eco" in profiles map in cmd/search.go
- [x] Update chat system prompt and refinement hint to mention eco profile
- [x] Add chat test recognizing eco profile
- [x] Verify build + test + vet pass

## Task 118: Add Indian airport clusters (DEL, BOM metro areas)
**Status:** done
**Plan:** Add DEL and BOM clusters to airports.go for NearbySearcher expansion.
- [x] Write tests: NearbyAirports("DEL") returns ["JAI"], NearbyAirports("BOM") returns ["PNQ"]
- [x] Add "Delhi" and "Mumbai" clusters to airportClusters map
- [x] Verify tests pass

## Task 119: Extract display formatting from cmd/search.go into cmd/display.go
**Status:** done
**Plan:** Move all display/formatting functions and JSON types from search.go to display.go.
- [x] Create cmd/display.go with display functions and types
- [x] Remove moved functions/types from cmd/search.go
- [x] Verify build passes (same package, no import changes needed)
- [x] Verify all cmd tests pass unchanged
