# TODO

Carried from: Day 47 (all completed)

## Tasks 130-134: Day 47 tasks
**Status:** completed -- Picker reasoning, printJSONWithInsights tests, price insights in zero-results, FRA stopovers, filter edge case tests

---

## Task 135: Include ranking reasoning in chat result summary
**Status:** done
**Plan:** Add Reasoning to top-3 entries in resultSummaryForChat.
- [x] Write test: resultSummaryForChat includes Reasoning when itinerary has non-empty Reasoning
- [x] Write test: resultSummaryForChat omits reasoning text when Reasoning is empty
- [x] Add Reasoning to the fmt.Fprintf in resultSummaryForChat loop
- [x] Verify build + test + vet pass

## Task 136: Auto-infer ranking profile from conversation context
**Status:** done
**Plan:** Add inferProfile function scanning user messages for preference keywords.
- [x] Write test: inferProfile returns "budget" for cheapest/save money keywords
- [x] Write test: inferProfile returns "comfort" for comfortable/hate layovers keywords
- [x] Write test: inferProfile returns "eco" for eco/green/carbon keywords
- [x] Write test: inferProfile returns "" for ambiguous or no signals
- [x] Implement inferProfile function in chat.go
- [x] Wire into chatLoop: use inferred profile when LLM doesn't set one
- [x] Write integration test: chatLoop applies inferred profile
- [x] Verify build + test + vet pass

## Task 137: India to Southeast Asia stopover routes
**Status:** done
**Plan:** Add DEL/BOM to BKK corridor stopovers.
- [x] Write tests for StopoversForRoute DEL-BKK and BKK-DEL
- [x] Write tests for StopoversForRoute BOM-BKK and BKK-BOM
- [x] Add DELToBKKStopovers slice (DOH, AUH, DXB, SIN, KUL, CCU)
- [x] Add BOMToBKKStopovers slice (DOH, AUH, DXB, SIN, KUL)
- [x] Register both in stopoversMap
- [x] Verify existing consistency test passes
- [x] Verify build + test + vet pass

## Task 138: Layover details in chat result summary
**Status:** done
**Plan:** Show layover city and duration for connecting flights in resultSummaryForChat.
- [x] Write test: multi-segment flight shows layover city and duration
- [x] Write test: direct flight shows "nonstop"
- [x] Write test: missing segment data falls back to stop count
- [x] Add layover extraction helper for chat summary
- [x] Update resultSummaryForChat to use layover details
- [x] Verify build + test + vet pass

## Task 139: SetRanker test + picker fallback coverage + stale TODO cleanup
**Status:** done
**Plan:** Fix coverage gaps and remove stale TODO.
- [x] Change TestPicker_BothPassesRankerToComposite to use p.SetRanker(ranker)
- [x] Add test: fallback returns first strategy when no "direct" exists (already covered)
- [x] Remove stale TODO at multicity.go:66 (dedup implemented Day 38)
- [x] Verify build + test + vet pass
