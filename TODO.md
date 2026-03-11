# TODO

Carried from: Day 38 (all completed)

## Tasks 95-99: Day 37 tasks
**Status:** completed -- refactor stage 4b filtering, total_trip JSON, departure time CLI flags, itinerary deduplication, stopover duration control

## Tasks 100-104: Day 38 tasks
**Status:** completed -- remove KiwiID from StopoverCity/SearchParams, simplify fetchWithDualSort, add India-US stopovers, consolidate stage 3 filter logging, no-results filter suggestion in chat

## Tasks 105-109: Day 40 tasks
**Status:** completed -- consolidate time-of-day filters, add single-flight predicates, India-UK stopovers, ranker cache stats, chat agent personality

---

## Task 105: Consolidate time-of-day filter functions
**Status:** done
**Plan:** Extract shared filterByTimeOfDay helper from FilterByDepartureTime and FilterByArrivalTime.
- [x] Extract filterByTimeOfDay(flights, after, before, extractTime) helper
- [x] Rewrite FilterByDepartureTime as thin wrapper
- [x] Rewrite FilterByArrivalTime as thin wrapper
- [x] Verify build + test + vet pass

## Task 106: Single-flight filter predicates for passesAllFilters
**Status:** done
**Plan:** Add single-flight predicate functions to eliminate []Flight{f} wrapping in passesAllFilters.
- [x] Write predicate tests (FlightPassesBlocked, FlightPassesAlliance, etc.)
- [x] Implement predicates in filter.go + parseAirlineCodes helper
- [x] Rewrite passesAllFilters in multicity.go to use predicates
- [x] Verify existing multicity tests + build + vet pass

## Task 107: Add India-UK route stopovers (DEL/BOM to LHR)
**Status:** done
**Plan:** Add curated stopover lists for DEL->LHR (6 cities) and BOM->LHR (6 cities).
- [x] Write tests for StopoversForRoute("DEL","LHR") and StopoversForRoute("BOM","LHR")
- [x] Add DELToLHRStopovers and BOMToLHRStopovers
- [x] Register routes in stopoversMap
- [x] Verify tests pass

## Task 108: Ranker cache hit/miss counters
**Status:** done
**Plan:** Add hits/misses counters to Ranker, CacheStats() method.
- [x] Write tests: hit+miss, all misses, empty
- [x] Add hits/misses fields, CacheStats() method
- [x] Increment counters in Rank()
- [x] Verify build + test + vet pass

## Task 109: Chat system prompt agent personality
**Status:** done
**Plan:** Enhance chatSystemPrompt to position booker as proactive travel agent per VISION.md.
- [x] Add agent personality + stopover/tradeoff/flexibility guidance
- [x] Verify existing chat tests pass + build + vet
