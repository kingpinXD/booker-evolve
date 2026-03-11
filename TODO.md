# TODO

Carried from: Day 49 (all completed)

## Tasks 140-144: Day 49 tasks
**Status:** completed -- Request-aware picker fallback, JSON segments, flex-date summary, NRT stopovers, stale worktree cleanup

---

## Task 145: Result caching and comparison in chat
**Status:** done
**Plan:** Store lastResults in chatLoop. Add keyword detection (looksLikeComparison/looksLikeDetail) and parseOptionIndices to intercept user input before LLM call. formatComparison and formatOptionDetail generate structured text from cached results. Intercept in chatLoop main loop before LLM call. Files: cmd/chat.go, cmd/chat_test.go.
- [ ] Add lastResults []search.Itinerary variable in chatLoop
- [ ] Write formatComparison(results, indices) helper
- [ ] Write formatOptionDetail(results, index) helper
- [ ] Write looksLikeComparison/looksLikeDetail keyword detection
- [ ] Write parseOptionIndices helper to extract numbers from user input
- [ ] Write unit tests for all helper functions
- [ ] Intercept user input in chatLoop before LLM call for compare/detail
- [ ] Add comparison/detail text to conversation history
- [ ] Write integration test for chatLoop comparison flow
- [ ] Verify build, test, vet pass

## Task 146: Proactive stopover suggestion in chat
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Write stopoverSuggestion(origin, dest, leg2Date) helper
- [ ] Write unit tests for stopoverSuggestion
- [ ] Wire into chatLoop after results display (single-leg only)
- [ ] Write integration test for chatLoop showing stopover tip
- [ ] Verify build, test, vet pass

## Task 147: India-Melbourne and India-Paris stopover routes
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Add DELToMELStopovers and BOMToMELStopovers
- [ ] Add DELToCDGStopovers and BOMToCDGStopovers
- [ ] Register all 4 in stopoversMap
- [ ] Write route-specific tests for all 4 corridors
- [ ] Verify TestStopoverDataConsistency passes
- [ ] Verify build, test, vet pass

## Task 148: Carbon diff annotation in display
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Enhance legCarbon to show diff percentage when available
- [ ] Add carbon_diff_pct to jsonLeg struct
- [ ] Populate carbon_diff_pct in buildJSONItineraries
- [ ] Write unit tests for legCarbon with diff values
- [ ] Write unit test for JSON output with carbon_diff_pct
- [ ] Verify build, test, vet pass

## Task 149: Chat filter reset via clear_fields
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Add ClearFields []string to tripParams
- [ ] Update mergeParams to skip propagation for cleared fields
- [ ] Update parsePartialParams to recognize clear_fields
- [ ] Update system prompt and refinement hint to document clear_fields
- [ ] Write unit tests for mergeParams with clear_fields
- [ ] Write integration test for chatLoop filter reset flow
- [ ] Verify build, test, vet pass
