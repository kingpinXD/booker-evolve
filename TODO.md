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
**Status:** pending
**Plan:**
- [ ] Write test: printJSONWithInsights with valid PriceInsights produces price_insights key
- [ ] Write test: printJSONWithInsights with empty PriceInsights omits price_insights
- [ ] Write test: printJSONWithInsights results array matches input count
- [ ] Verify build + test + vet pass

## Task 132: Thread PriceInsights into zero-results chat suggestion
**Status:** pending
**Plan:**
- [ ] Write test: zero-results output includes price range when insights have data
- [ ] Write test: zero-results output omits price info when no insights
- [ ] Add priceInsightHint helper in chat.go
- [ ] Thread PriceInsights into chatLoop zero-results block
- [ ] Verify build + test + vet pass

## Task 133: Add DEL/BOM to FRA stopover routes
**Status:** pending
**Plan:**
- [ ] Write tests for StopoversForRoute DEL-FRA and FRA-DEL
- [ ] Write tests for StopoversForRoute BOM-FRA and FRA-BOM
- [ ] Add DELToFRAStopovers slice (DOH, AUH, DXB, IST, BAH, KWI)
- [ ] Add BOMToFRAStopovers slice (DOH, AUH, DXB, IST, BAH)
- [ ] Register both in stopoversMap
- [ ] Verify existing consistency test passes
- [ ] Verify build + test + vet pass

## Task 134: Test filter edge cases (firstDeparture, flightPassesTimeOfDay)
**Status:** pending
**Plan:**
- [ ] Write test: firstDeparture with 0 legs returns zero time
- [ ] Write test: firstDeparture with empty outbound segments
- [ ] Write test: flightPassesTimeOfDay with empty Outbound returns false
- [ ] Write test: flightPassesTimeOfDay with invalid HH:MM returns true
- [ ] Verify build + test + vet pass
