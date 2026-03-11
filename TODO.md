# TODO

Carried from: Day 43 (all completed)

## Tasks 115-119: Day 43 tasks
**Status:** completed -- Kiwi doc cleanup, India-Australia stopovers, eco ranking profile, Indian airport clusters, display extraction

---

## Task 120: Fix chat profile switching (dynamic ranker per search)
**Status:** done
**Plan:** Add SetWeights to Ranker, share one Ranker between direct+multicity strategies, define weightsUpdater interface for chatLoop, update weights before each search when profile changes.
- [x] Write test: changing profile mid-chat changes ranking weights used
- [x] Add SetWeights method to Ranker
- [x] Share single Ranker between direct and multicity (changed NewSearcher signature)
- [x] Wire profileWeights(params.Profile) into chatLoop before each search via weightsUpdater
- [x] Integration test: profile switch mid-chat triggers correct weight updates
- [x] Verify build + test + vet + lint pass

## Task 121: Extract StripCodeFences helper to deduplicate 4 call sites
**Status:** done
**Plan:** Add StripCodeFences to llm/client.go, write 6-case table test, replace 4 call sites with single function calls.
- [x] Write test for llm.StripCodeFences (json fences, plain fences, no fences, nested, empty, fence-only)
- [x] Add StripCodeFences to llm/client.go
- [x] Replace 4 call sites: chat.go (x2), picker.go, ranker.go
- [x] Verify all existing tests pass

## Task 122: Fix NearbySearcher ignoring SortBy
**Status:** done
**Plan:** Replace hardcoded sort.Slice by price with search.SortResults(merged, req.SortBy). Add test with SortBy="duration".
- [x] Write test: NearbySearcher with SortBy="duration" returns duration-sorted results
- [x] Replace hardcoded price sort with search.SortResults(merged, req.SortBy)
- [x] Verify existing nearby tests pass

## Task 123: Thread user Context to multicity ranker
**Status:** done
**Plan:** Add Context to SearchParams, thread from toSearchParams, add UserContext field to Ranker, append USER PREFERENCES to buildRankingPrompt.
- [x] Write test: buildRankingPrompt includes context when non-empty
- [x] Write test: buildRankingPrompt unchanged when context is empty
- [x] Add Context field to SearchParams
- [x] Map req.Context to SearchParams.Context in toSearchParams
- [x] Add UserContext field to Ranker, pass to buildRankingPrompt
- [x] Verify build + test + vet pass

## Task 124: Respect departure time preferences in CombineLegs red-eye filter
**Status:** done
**Plan:** Add DepartureAfter/DepartureBefore to CombineParams, skip red-eye check when either is set, thread from SearchParams in multicity.go.
- [x] Write test: CombineLegs with DepartureAfter="01:00" allows 02:00 leg2 departures
- [x] Write test: CombineLegs without explicit times still rejects red-eye
- [x] Add DepartureAfter/DepartureBefore to CombineParams
- [x] Skip isRedEye check when user has explicit departure time constraints
- [x] Thread DepartureAfter/DepartureBefore from SearchParams to CombineParams in multicity.go
- [x] Verify build + test + vet + lint pass
