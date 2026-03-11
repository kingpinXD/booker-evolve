# TODO

## Task 155: Refactor parsePartialParams to use reflection
**Status:** done
**Plan:** Replace 10-line OR chain with reflection-based anyFieldSet helper. Iterate struct fields via reflect, check IsZero(). Auto-supports new fields.
- [x] Write anyFieldSet helper using reflection
- [x] Replace OR chain in parsePartialParams with anyFieldSet call
- [x] Verify existing tests pass
- [x] Run full verification

## Task 156: Add price insights to chat conversation history
**Status:** pending
**Plan:** Add optional PriceInsights parameter to resultSummaryForChat. When PriceLevel is non-empty, append typical range and level to the summary string. Update chatLoop call site to pass insights.
- [ ] Add priceInsight param to resultSummaryForChat
- [ ] Write test for summary with price insights
- [ ] Update chatLoop call site
- [ ] Run full verification

## Task 157: Add India-Seoul stopover corridor
**Status:** pending
**Plan:** Add DELToICNStopovers and BOMToICNStopovers variables with East/Southeast Asian cities (BKK, SIN, HKG, TPE, KUL). Register in stopoversMap. Verify TestStopoverDataConsistency.
- [ ] Add DELToICNStopovers variable
- [ ] Add BOMToICNStopovers variable
- [ ] Register both in stopoversMap
- [ ] Run tests

## Task 158: Add India-Hong Kong stopover corridor
**Status:** pending
**Plan:** Add DELToHKGStopovers and BOMToHKGStopovers variables with Southeast Asian cities (BKK, SIN, KUL, CCU/CMB, TPE). Register in stopoversMap. Verify TestStopoverDataConsistency.
- [ ] Add DELToHKGStopovers variable
- [ ] Add BOMToHKGStopovers variable
- [ ] Register both in stopoversMap
- [ ] Run tests

## Task 159: Extract displayChatResults helper from chatLoop
**Status:** pending
**Plan:** Extract lines 866-882 of chatLoop (format switch, printTable/printJSON, price insights display) into displayChatResults(out, results, insights, cur). Reduces chatLoop and improves testability.
- [ ] Extract displayChatResults function
- [ ] Update chatLoop to call helper
- [ ] Add unit test for displayChatResults
- [ ] Run full verification
