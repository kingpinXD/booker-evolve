# TODO

Carried from: Day 48 (all completed)

## Tasks 135-139: Day 48 tasks
**Status:** completed -- Ranking reasoning in summary, auto-infer profile, BKK stopovers, layover details, picker test cleanup

---

## Task 140: Request-aware Picker fallback
**Status:** done
**Plan:** Make fallback inspect request shape instead of always returning "direct".
- [x] Write test: fallback with Leg2Date set expects "multicity" strategy and accurate reason
- [x] Write test: fallback with ReturnDate set expects "direct" strategy (handles round-trips)
- [x] Write test: fallback with plain single-leg expects "direct" strategy
- [x] Refactor fallback() to accept Request and return (Strategy, string)
- [x] Update Pick() to pass req to fallback
- [x] Verify build + test + vet pass

## Task 141: Segments array in JSON output
**Status:** done
**Plan:** Add per-segment detail array to jsonLeg in JSON output.
- [x] Define jsonSegment struct with airline, flight_number, origin, destination, departure, arrival, duration, aircraft, legroom, layover_duration, overnight
- [x] Add Segments field to jsonLeg (omitempty)
- [x] Populate segments in buildJSONItineraries loop
- [x] Write test: multi-segment leg produces correct segments array
- [x] Write test: single-segment leg produces 1-element array
- [x] Write test: segments omitted when empty
- [x] Verify existing JSON tests still pass
- [x] Verify build + test + vet pass

## Task 142: Flex-date departure date in chat result summary
**Status:** done
**Plan:** Show departure date in top-3 entries when FlexDays > 0.
- [x] Write test: resultSummaryForChat with FlexDays > 0 includes departure date
- [x] Write test: resultSummaryForChat with FlexDays = 0 omits departure date
- [x] Write test: resultSummaryForChat with FlexDays > 0 and empty segments (no crash)
- [x] Modify resultSummaryForChat to include date from segment DepartureTime when FlexDays > 0
- [x] Format departure date as "Jan 2" short form in each entry
- [x] Verify build + test + vet pass

## Task 143: India-Tokyo stopover routes
**Status:** done
**Plan:** Add DEL/BOM to NRT stopover corridors.
- [x] Write tests for StopoversForRoute DEL-NRT and NRT-DEL
- [x] Write tests for StopoversForRoute BOM-NRT and NRT-BOM
- [x] Add DELToNRTStopovers slice (BKK, SIN, HKG, TPE, ICN, KUL)
- [x] Add BOMToNRTStopovers slice (BKK, SIN, HKG, TPE, ICN)
- [x] Register both in stopoversMap
- [x] Verify existing consistency test passes
- [x] Verify build + test + vet pass

## Task 144: Stale worktree cleanup
**Status:** done
**Plan:** Remove abandoned worktree directories and git branches.
- [x] Remove .claude/worktrees/agent-ade875dc and agent-a69bbfb5
- [x] Delete git branches worktree-agent-ade875dc and worktree-agent-a69bbfb5
- [x] Verify git worktree list and git branch are clean
