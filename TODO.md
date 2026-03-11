# TODO

Carried from: Day 44 (all completed)

## Tasks 120-124: Day 44 tasks
**Status:** completed -- Dynamic profile switching, StripCodeFences dedup, NearbySearcher SortBy fix, user context to ranker, red-eye departure time override

---

## Task 125: Wire ranker to composite strategy
**Status:** done
**Plan:**
- [x] Write test: picker with "both" response creates composite that ranks results
- [x] Add Ranker field to Picker + SetRanker method
- [x] Pass ranker to NewCompositeStrategy in pickWithLLM "both" branch
- [x] Update buildPicker in cmd/infra.go to call SetRanker
- [x] Verify existing picker tests still pass
- [x] Verify build + test + vet + lint pass

## Task 126: Consolidate itinRoute and deduplicate to search package
**Status:** done
**Plan:**
- [x] Export ItinRoute and DeduplicateItineraries in search package (composite.go)
- [x] Update composite.go to use exported functions
- [x] Update nearby/nearby.go to use search.DeduplicateItineraries
- [x] Remove duplicate unexported functions from nearby.go
- [x] Verify all existing tests pass unchanged
- [x] Verify build + test + vet + lint pass

## Task 127: Add "score" sort mode to SortResults
**Status:** done
**Plan:**
- [x] Write test: SortResults with "score" sorts by Score descending
- [x] Write test: SortResults with "score" and all-zero scores is stable
- [x] Add "score" case to SortResults switch in filter.go
- [x] Update chat system prompt sort_by description to include "score"
- [x] Update refinementHint to mention sort_by "score"
- [x] Update --sort-by flag description in cmd/search.go
- [x] Verify build + test + vet + lint pass

## Task 128: Fix round-trip max_price to check total itinerary price
**Status:** done
**Plan:**
- [x] Write test: round-trip with max_price, verify total price filtering
- [x] Add total-price filter after combineRoundTrip in direct.go Search method
- [x] Verify existing round-trip tests still pass
- [x] Verify build + test + vet pass

## Task 129: Clean stale Kiwi references in search and filter docs
**Status:** done
**Plan:**
- [x] Update search/search.go line 34 "currently Kiwi only" to SerpAPI
- [x] Update filter.go line 138-139 Kiwi reference to generic provider description
- [x] Verify build + test + vet + lint clean
