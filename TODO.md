# TODO

Carried from: Day 46 (all completed)

## Tasks 125-129: Day 46 tasks
**Status:** completed -- Composite ranker wiring, dedup consolidation, score sort, round-trip price fix, Kiwi doc cleanup

---

## Task 130: Return strategy reasoning from Picker.Pick
**Status:** done
**Plan:** Changed Pick to return (Strategy, string, error). LLM path returns parsed reason, fallback returns descriptive string. Both callers (search.go, chat.go) display the reason.
- [x] Write test: Pick returns non-empty reason from LLM path
- [x] Write test: Pick returns non-empty reason from fallback path
- [x] Change Pick signature to return (Strategy, string, error)
- [x] Update pickWithLLM to return reason from pickerResult
- [x] Update fallback to return descriptive reason string
- [x] Update cmd/search.go caller for new signature
- [x] Update cmd/chat.go caller to display reason in search output
- [x] Write chat test verifying reason appears in output
- [x] Verify build + test + vet pass

## Task 131: Test printJSONWithInsights
**Status:** done
**Plan:** Added 5 test cases in display_test.go covering valid/empty PriceInsights and result count.
- [x] Write test: printJSONWithInsights with valid PriceInsights produces price_insights key
- [x] Write test: printJSONWithInsights with empty PriceInsights omits price_insights
- [x] Write test: printJSONWithInsights results array matches input count
- [x] Verify build + test + vet pass

## Task 132: Thread PriceInsights into zero-results chat suggestion
**Status:** done
**Plan:** Added priceInsightHint helper, wired into chatLoop zero-results block after zeroResultsSuggestion. Shows "Typical prices for this route: $X-$Y (price level: Z)" when insights available.
- [x] Write test: zero-results output includes price range when insights have data
- [x] Write test: zero-results output omits price info when no insights
- [x] Add priceInsightHint helper in chat.go
- [x] Thread PriceInsights into chatLoop zero-results block
- [x] Verify build + test + vet pass

## Task 133: Add DEL/BOM to FRA stopover routes
**Status:** done
**Plan:** Added DELToFRAStopovers (6 cities) and BOMToFRAStopovers (5 cities) via Gulf hubs. 4 test cases for bidirectional lookup.
- [x] Write tests for StopoversForRoute DEL-FRA and FRA-DEL
- [x] Write tests for StopoversForRoute BOM-FRA and FRA-BOM
- [x] Add DELToFRAStopovers slice (DOH, AUH, DXB, IST, BAH, KWI)
- [x] Add BOMToFRAStopovers slice (DOH, AUH, DXB, IST, BAH)
- [x] Register both in stopoversMap
- [x] Verify existing consistency test passes
- [x] Verify build + test + vet pass

## Task 134: Test filter edge cases (firstDeparture, flightPassesTimeOfDay)
**Status:** done
**Plan:** Added 5 test cases covering zero-legs, empty-outbound, invalid-time-format paths.
- [x] Write test: firstDeparture with 0 legs returns zero time
- [x] Write test: firstDeparture with empty outbound segments
- [x] Write test: flightPassesTimeOfDay with empty Outbound returns false
- [x] Write test: flightPassesTimeOfDay with invalid HH:MM returns true
- [x] Verify build + test + vet pass
