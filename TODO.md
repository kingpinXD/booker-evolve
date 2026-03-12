# TODO

## Task 160: Extract chat helpers into cmd/chathelpers.go
**Status:** done
**Plan:** Move ~20 helper functions from chat.go to new chathelpers.go. Pure file split, no logic changes.
- [x] Identify functions to extract (helpers, param functions, utilities)
- [x] Create cmd/chathelpers.go with extracted functions
- [x] Remove extracted functions from cmd/chat.go
- [x] Verify go build && go test ./cmd/...

## Task 161: filterSuggestion reflection refactor
**Status:** done
**Plan:** Replace per-field if-blocks with reflection loop. Define filterLabels map from json tag to human-readable name. Iterate tripParams fields, check non-zero + in label map.
- [x] Define filterLabels map
- [x] Rewrite filterSuggestion using reflection
- [x] Verify existing tests pass
- [x] Run full verification

## Task 162: Multi-leg info in formatComparison
**Status:** pending
**Plan:** Show per-leg details for multi-city itineraries. Extract shared leg-extraction logic into legSummary helper used by both formatComparison and formatOptionDetail.
- [ ] Extract legSummary helper function
- [ ] Update formatComparison to show all legs
- [ ] Update formatOptionDetail to use legSummary
- [ ] Add test for 2-leg comparison
- [ ] Run full verification

## Task 163: India-Los Angeles stopover corridors
**Status:** done (parallel worktree)
**Plan:** Add DELToLAXStopovers and BOMToLAXStopovers via East Asia Pacific corridor (BKK, SIN, KUL, HKG, NRT, TPE). Register in stopoversMap.
- [x] Add DELToLAXStopovers variable
- [x] Add BOMToLAXStopovers variable
- [x] Register both in stopoversMap
- [x] Add route-specific tests + reverse lookup tests
- [x] Run full verification

## Task 164: India-Chicago stopover corridors
**Status:** done (parallel worktree)
**Plan:** Add DELToORDStopovers and BOMToORDStopovers via East Asia Pacific (BKK, SIN, HKG, NRT, ICN) + European (IST) corridor. Register in stopoversMap.
- [x] Add DELToORDStopovers variable
- [x] Add BOMToORDStopovers variable
- [x] Register both in stopoversMap
- [x] Add route-specific tests + reverse lookup tests
- [x] Run full verification
