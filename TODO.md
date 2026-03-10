# TODO

Carried from: Day 16 (all completed)

## Task 1: Surface PriceInsights from SerpAPI
**Status:** completed (Day 12)

## Task 2: Expand stopover route corridors
**Status:** completed (Day 12)

## Task 3: Show booking URL in JSON output
**Status:** completed (Day 12)

## Task 4: Add price summary footer to table output
**Status:** completed (Day 12)

## Task 5: cmd printTable stdout capture tests
**Status:** completed (Day 12)

## Task 6: Fix gofmt issue in search/search.go
**Status:** completed (Day 16)

## Task 7: Add CompositeStrategy for running multiple strategies in parallel
**Status:** completed (Day 16)

## Task 8: Extend Picker to support composite strategy selection
**Status:** completed (Day 16)

## Task 9: Add chat command with conversational LLM loop
**Status:** completed (Day 16)

## Task 10: Wire chat command to execute searches
**Status:** completed (Day 16)

---

## Task 11: Fix errcheck lint in cmd/chat.go
**Status:** completed (Day 17)
**Notes:** Discarded return values on 7 fmt.Fprint* calls. golangci-lint now reports 0 issues.

## Task 12: Add result summary to chat conversation history
**Status:** completed (Day 17)
**Notes:** Added resultSummaryForChat helper. After displaying results, summary (count + price range) is appended to conversation history as assistant message. 2 new tests.

## Task 13: Add airport cluster data
**Status:** completed (Day 17)
**Notes:** search/airports.go with 14 metro-area clusters, NearbyAirports function with O(1) lookup via reverse index. 4 tests.

## Task 14: Flex-date multi-search in direct strategy
**Status:** completed (Day 17)
**Notes:** direct.Search now loops over [dep-flex, dep+flex] dates when FlexDays > 0, making 2*flex+1 provider calls and merging results. 2 new tests with dateTrackingProvider mock.

## Task 15: Surface airport suggestions in chat system prompt
**Status:** completed (Day 17)
**Notes:** Chat system prompt now mentions nearby airports. nearbyAirportHint shows tips after param extraction. 2 new tests.
