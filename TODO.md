# TODO

Carried from: Day 40 (all completed)

## Tasks 95-99: Day 37 tasks
**Status:** completed -- refactor stage 4b filtering, total_trip JSON, departure time CLI flags, itinerary deduplication, stopover duration control

## Tasks 100-104: Day 38 tasks
**Status:** completed -- remove KiwiID from StopoverCity/SearchParams, simplify fetchWithDualSort, add India-US stopovers, consolidate stage 3 filter logging, no-results filter suggestion in chat

## Tasks 105-109: Day 40 tasks
**Status:** completed -- consolidate time-of-day filters, add single-flight predicates, India-UK stopovers, ranker cache stats, chat agent personality

---

## Task 110: Refactor ranker sort and extract applyScores helper
**Status:** pending
**Plan:**
- [ ] Write test for sort stability with equal scores
- [ ] Replace bubble sort in applySortByScore with sort.Slice
- [ ] Extract duplicate score-application loop into applyScores helper
- [ ] Verify build + test + vet pass

## Task 111: Bidirectional route lookup in StopoversForRoute
**Status:** done
**Plan:** Add reverse lookup to StopoversForRoute: check dest->origin when forward not found, filter origin/dest.
- [x] Write test: StopoversForRoute("YYZ","DEL") returns route-specific, not fallback
- [x] Modify StopoversForRoute to check reverse key when forward not found
- [x] Filter origin/dest from reverse results
- [x] Verify all existing stopover tests still pass

## Task 112: Add India-US West Coast stopovers (DEL/BOM to SFO)
**Status:** done
**Plan:** Add curated stopover lists for DEL->SFO (6 cities) and BOM->SFO (5 cities).
- [x] Write tests for StopoversForRoute("DEL","SFO") and StopoversForRoute("BOM","SFO")
- [x] Add DELToSFOStopovers and BOMToSFOStopovers
- [x] Register routes in stopoversMap
- [x] Verify tests pass

## Task 113: Zero-results proactive suggestions in chat
**Status:** pending
**Plan:**
- [ ] Write test: zero-results output includes nearby airports when available
- [ ] Write test: zero-results mentions flex dates
- [ ] Implement zeroResultsSuggestion helper
- [ ] Wire into chatLoop zero-results block
- [ ] Verify build + test + vet pass

## Task 114: Stopover data consistency validation test
**Status:** pending
**Plan:**
- [ ] Write test: all route stopovers exclude origin/dest airports
- [ ] Write test: all IATA codes are 3 uppercase letters
- [ ] Write test: MinStay < MaxStay, City/Notes non-empty
- [ ] Verify tests pass
