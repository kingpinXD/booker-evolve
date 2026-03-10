# TODO

## Task 1: Surface PriceInsights from SerpAPI
**Status:** pending
**Plan:**
- [ ] Add PriceInsights type to search/search.go on Itinerary
- [ ] Modify serpapi ParseResponse to return PriceInsights alongside flights
- [ ] Propagate PriceInsights through direct.Search to itineraries
- [ ] Display price insights footer in table output
- [ ] Add price_insights fields to JSON output
- [ ] Write tests first (TDD), verify all pass

## Task 2: Expand stopover route corridors
**Status:** pending
**Plan:**
- [ ] Write tests for BOM→YYZ, DEL→YVR routes and nil-fallback
- [ ] Add BOM→YYZ stopover cities to stopovers.go
- [ ] Add DEL→YVR stopover cities to stopovers.go
- [ ] Update StopoversForRoute to handle new routes
- [ ] Fix fallback to return nil for unknown routes
- [ ] Verify existing stopovers tests still pass

## Task 3: Show booking URL in JSON output
**Status:** pending
**Plan:**
- [ ] Write test for booking_url in JSON output
- [ ] Add BookingURL field to jsonLeg struct
- [ ] Wire BookingURL from types.Flight into jsonLeg
- [ ] Verify test passes

## Task 4: Add price summary footer to table output
**Status:** pending
**Plan:**
- [ ] Write test for priceSummary helper function
- [ ] Implement priceSummary(itineraries, currency) string
- [ ] Call priceSummary after table render in printTable
- [ ] Integrate PriceInsights level if available (from Task 1)
- [ ] Verify test passes

## Task 5: cmd printTable stdout capture tests
**Status:** pending
**Plan:**
- [ ] Write TestPrintTable_SingleLeg with stdout capture
- [ ] Write TestPrintTable_MultiLeg with stdout capture
- [ ] Verify header columns match expected for each case
- [ ] Check coverage increase with go test -cover ./cmd
