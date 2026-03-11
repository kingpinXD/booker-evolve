# TODO

Carried from: Day 31 (all completed)

## Tasks 70-74: Day 31 tasks
**Status:** completed -- multicity departure time filter, SerpAPI stops+return_date params, ranker city name cleanup, build gate clean

---

## Tasks 75-79: Day 33 tasks
**Status:** completed -- post-fetch sorting, connection risk tags, sort_by in chat, JSON enrichment, avoid airline filter

## Task 75: Post-fetch sorting in direct strategy + CLI flag
**Status:** done
- [x] SortResults in filter.go (price/duration/departure) + 5 tests
- [x] --sort-by CLI flag, wired into direct.go

## Task 76: Connection risk tags in ranker prompt
**Status:** done
- [x] [Risky connection: Xm] / [Tight connection: Xm] tags in buildRankingPrompt + 3 tests

## Task 77: Wire sort_by into chat conversation
**Status:** done
- [x] SortBy in tripParams, parsePartialParams, mergeParams, buildRequestFromParams, system prompt, refinement hint + 5 tests

## Task 78: Enrich JSON output with airline codes and city names
**Status:** done
- [x] AirlineCode, OriginCity, DestinationCity, OriginName, DestinationName in jsonLeg + 2 tests

## Task 79: Avoid airline filter
**Status:** done
- [x] FilterByAvoidAirlines in filter.go + 5 tests
- [x] Wired into direct, multicity, --avoid-airlines CLI flag
