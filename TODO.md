# TODO

## Task 1: Surface PriceInsights from SerpAPI
**Status:** completed (Day 12)
**Notes:** PriceInsights type added to search/search.go, parsed in serpapi, stored on provider with getter. Tests written.

## Task 2: Expand stopover route corridors
**Status:** completed (Day 12)
**Notes:** Added BOM->YYZ (7 cities), DEL->YVR (6 cities). Fixed fallback to return nil for unknown routes.

## Task 3: Show booking URL in JSON output
**Status:** completed (Day 12)
**Notes:** BookingURL field added to jsonLeg, wired from types.Flight.BookingURL.

## Task 4: Add price summary footer to table output
**Status:** completed (Day 12)
**Notes:** priceSummary helper shows result count and price range after table.

## Task 5: cmd printTable stdout capture tests
**Status:** completed (Day 12)
**Notes:** TestPrintTable_SingleLeg and TestPrintTable_MultiLeg with stdout capture. cmd coverage 54.5% to 68.2%.

---

## Task 6: Fix gofmt issue in search/search.go
**Status:** completed (Day 16)
**Notes:** PriceInsights struct alignment fixed with gofmt -w.

## Task 7: Add CompositeStrategy for running multiple strategies in parallel
**Status:** completed (Day 16)
**Notes:** search/composite.go runs child strategies concurrently via sync.WaitGroup, merges/deduplicates by route+price, optional Ranker. 10 tests with race detector.

## Task 8: Extend Picker to support composite strategy selection
**Status:** completed (Day 16)
**Notes:** Picker recognizes "both" from LLM, wraps strategies in CompositeStrategy. System prompt updated. 3 new tests.

## Task 9: Add chat command with conversational LLM loop
**Status:** completed (Day 16)
**Notes:** cmd/chat.go with cobra subcommand. Multi-turn LLM dialogue gathers trip params, emits JSON, parses into search.Request. 10 tests.

## Task 10: Wire chat command to execute searches
**Status:** completed (Day 16)
**Notes:** Merged with Task 9. chatLoop runs Picker + strategy + printTable/printJSON after param extraction. Refinement prompt shown after results.
