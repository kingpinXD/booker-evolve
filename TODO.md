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
**Status:** pending
**Plan:** Add inferProfile function scanning user messages for preference keywords.
- [ ] Write test: inferProfile returns "budget" for cheapest/save money keywords
- [ ] Write test: inferProfile returns "comfort" for comfortable/hate layovers keywords
- [ ] Write test: inferProfile returns "eco" for eco/green/carbon keywords
- [ ] Write test: inferProfile returns "" for ambiguous or no signals
- [ ] Implement inferProfile function in chat.go
- [ ] Wire into chatLoop: use inferred profile when LLM doesn't set one
- [ ] Write integration test: chatLoop applies inferred profile
- [ ] Verify build + test + vet pass

## Task 137: India to Southeast Asia stopover routes
**Status:** pending
**Plan:** Add DEL/BOM to BKK corridor stopovers.
- [ ] Write tests for StopoversForRoute DEL-BKK and BKK-DEL
- [ ] Write tests for StopoversForRoute BOM-BKK and BKK-BOM
- [ ] Add DELToBKKStopovers slice (DOH, AUH, DXB, SIN, KUL, CCU)
- [ ] Add BOMToBKKStopovers slice (DOH, AUH, DXB, SIN, KUL)
- [ ] Register both in stopoversMap
- [ ] Verify existing consistency test passes
- [ ] Verify build + test + vet pass

## Task 138: Layover details in chat result summary
**Status:** pending
**Plan:** Show layover city and duration for connecting flights in resultSummaryForChat.
- [ ] Write test: multi-segment flight shows layover city and duration
- [ ] Write test: direct flight shows "nonstop"
- [ ] Write test: missing segment data falls back to stop count
- [ ] Add layover extraction helper for chat summary
- [ ] Update resultSummaryForChat to use layover details
- [ ] Verify build + test + vet pass

## Task 139: SetRanker test + picker fallback coverage + stale TODO cleanup
**Status:** pending
**Plan:** Fix coverage gaps and remove stale TODO.
- [ ] Change TestPicker_BothPassesRankerToComposite to use p.SetRanker(ranker)
- [ ] Add test: fallback returns first strategy when no "direct" exists
- [ ] Remove stale TODO at multicity.go:66 (dedup implemented Day 38)
- [ ] Verify build + test + vet pass
