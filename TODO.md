# TODO

Carried from: Day 33 (all completed)

## Tasks 75-79: Day 33 tasks
**Status:** completed -- post-fetch sorting, connection risk tags, sort_by in chat, JSON enrichment, avoid airline filter

---

## Tasks 80-84: Day 34 tasks
**Status:** completed -- avoid-airlines in chat, multicity in chat, arrival time filter, max duration filter, cmd helper coverage

## Pre-task: Commit stale gofmt fix
**Status:** done
- [x] Commit ranker_test.go gofmt alignment fix from Session 33

## Task 80: Wire AvoidAirlines into chat conversation
**Status:** done
**Plan:** Add avoid_airlines to tripParams, wire through all 5 chat pipeline stages (parse/merge/build/prompt/hint), add tests.
- [x] Add AvoidAirlines string to tripParams struct
- [x] Wire into parsePartialParams
- [x] Wire into mergeParams
- [x] Wire into buildRequestFromParams
- [x] Add to system prompt
- [x] Add to refinement hint
- [x] Write tests for parse/merge/build/prompt

## Task 81: Wire multicity/leg2_date into chat conversation
**Status:** done
**Plan:** Add Leg2Date to tripParams and search.Request. Wire through chat pipeline. Update multicity Strategy to prefer req.Leg2Date over constructor default. Also wire AvoidAirlines through to multicity Strategy (was missing).
- [x] Add Leg2Date string to tripParams
- [x] Add Leg2Date to search.Request
- [x] Wire into parsePartialParams, mergeParams, buildRequestFromParams
- [x] Update system prompt with leg2_date guidance
- [x] Update refinement hint
- [x] Modify multicity.Strategy to prefer req.Leg2Date over default
- [x] Wire AvoidAirlines through toSearchParams (was missing)
- [x] Write tests (chat + strategy override/fallback)

## Task 82: Arrival time filter (ArrivalBefore/ArrivalAfter)
**Status:** done
**Plan:** Add FilterByArrivalTime in filter.go using parseHHMM. Add ArrivalAfter/ArrivalBefore to search.Request. Wire through direct, multicity, CLI, and chat.
- [x] Add ArrivalBefore/ArrivalAfter to search.Request
- [x] Implement FilterByArrivalTime in filter.go
- [x] Write unit tests for FilterByArrivalTime
- [x] Wire into direct pipeline
- [x] Wire into multicity stages (FILTER + 4b)
- [x] Add --arrival-after/--arrival-before CLI flags
- [x] Wire into chat tripParams (parse/merge/build/prompt/hint)

## Task 83: Max duration filter
**Status:** done
**Plan:** Add FilterByMaxDuration in filter.go. Add MaxDuration to search.Request. Wire through direct, multicity, CLI (--max-duration), and chat (max_duration_hours).
- [x] Add MaxDuration to search.Request
- [x] Implement FilterByMaxDuration in filter.go
- [x] Write unit tests for FilterByMaxDuration
- [x] Wire into direct pipeline
- [x] Wire into multicity stages (FILTER + 4b)
- [x] Add --max-duration CLI flag
- [x] Wire into chat tripParams (max_duration_hours)

## Task 84: Coverage for low-coverage cmd helpers
**Status:** done
**Plan:** Add edge-case tests for empty segments and out-of-bounds paths in cmd helpers.
- [x] Add edge-case tests for legAircraft, legLegroom, legBookingURL (empty segs + OOB)
- [x] Add edge-case tests for legCabin, legArrival, legDeparture (empty segs)
- [x] Add edge-case tests for legSeatsLeft (OOB)
- [x] Add edge-case tests for formatPriceInsights (PriceLevel-only, zero range)
