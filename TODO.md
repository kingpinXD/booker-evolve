# TODO

Carried from: Day 34 (all completed)

## Tasks 80-84: Day 34 tasks
**Status:** completed -- avoid-airlines in chat, multicity in chat, arrival time filter, max duration filter, cmd helper coverage

---

## Tasks 85-89: Day 35 tasks
**Status:** pending

## Task 85: Next-day arrival indicator
**Status:** done
**Plan:** Add isNextDay(dep, arr) helper comparing dates. Modify legArrival to append " (+N)" when arrival date > departure date. Add arrival_next_day bool to jsonLeg. Test with same-day and next-day flights. Files: cmd/search.go, cmd/search_test.go.
- [x] Write test for next-day detection helper
- [x] Implement helper to detect arrival date > departure date
- [x] Modify legArrival to append (+N) marker
- [x] Add arrival_next_day boolean to jsonLeg
- [x] Write tests for table and JSON output with next-day arrivals
- [x] Verify existing tests still pass

## Task 86: Operating carrier display (codeshare indicator)
**Status:** done
**Plan:** Modify legAirlines to append "(op. XX)" when OperatingCarrier differs from Airline. Add operating_carrier to jsonLeg. Test codeshare and non-codeshare segments. Files: cmd/search.go, cmd/search_test.go.
- [x] Write test for codeshare display format "AC (op. UA)"
- [x] Modify legAirlines to show operating carrier when different
- [x] Add operating_carrier to jsonLeg struct
- [x] Populate operating_carrier in buildJSONItineraries
- [x] Write tests for non-codeshare case (no change)
- [x] Verify existing tests still pass

## Task 87: Richer result summary in chat history
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Write test with 5+ results expecting top 3 in summary
- [ ] Write test with 1-2 results for graceful degradation
- [ ] Modify resultSummaryForChat to include top 3 results
- [ ] Include price, airline, duration, stops per result
- [ ] Verify 0-result case unchanged
- [ ] Verify existing chat tests still pass

## Task 88: Preferred airlines filter (positive filter)
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Write FilterByPreferredAirlines tests (empty keeps all, single code, multiple codes, operating carrier match)
- [ ] Implement FilterByPreferredAirlines in filter.go
- [ ] Add PreferredAirlines to search.Request
- [ ] Wire into direct pipeline
- [ ] Wire into multicity stages (FILTER + 4b)
- [ ] Add --preferred-airlines CLI flag
- [ ] Wire into chat tripParams (parse/merge/build/prompt/hint)
- [ ] Write chat tests

## Task 89: Ranker LLM response caching
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Write test: identical itineraries + weights -> cache hit (mock LLM called once)
- [ ] Write test: different itineraries -> cache miss
- [ ] Write test: different weights -> cache miss
- [ ] Implement cache key generation (hash sorted candidates + weights)
- [ ] Add in-memory cache map to Ranker struct
- [ ] Short-circuit Rank() on cache hit
- [ ] Remove TODO comment from ranker.go
