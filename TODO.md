# TODO

## Task 1: Gate live API tests behind integration build tag
**Status:** done
**Commit:** 975623d

## Task 2: Add unit tests for search/filter.go
**Status:** done (100% coverage)
**Commit:** 17834ef

## Task 3: Add unit tests for types/types.go
**Status:** done (100% coverage)
**Commit:** a8a3ddd

## Task 4: Add Strategy interface and common Request type
**Status:** done
**Commit:** 88916bf

## Task 5: Add multicity Strategy adapter
**Status:** done
**Commit:** 5b7da67

---

## Day 3 Tasks

## Task 6: Add unit tests for provider/kiwi/parser.go
**Status:** pending
**Plan:**
- [ ] Write parser_test.go with table-driven tests using constructed Response structs
- [ ] Cover all exported and unexported functions (ParseResponse, parseItinerary, parseSector, parseSegment, extractBookingURL, buildCarrierLookup, mapCabinClass)
- [ ] Target 80%+ coverage

## Task 7: Add unit tests for search/multicity/combiner.go
**Status:** pending
**Plan:**
- [ ] Write combiner_test.go with table-driven tests
- [ ] Cover CombineLegs (valid/invalid pairs, gap constraints, layover rejection), hasLongLayover, PrimaryAirline, SameAirline
- [ ] Target 80%+ coverage

## Task 8: Implement direct search Strategy (issue #3)
**Status:** pending
**Note:** Deferred unless Tasks 6-7 complete quickly.

## Task 9: Wire Strategy picker into cmd/search.go (issue #5)
**Status:** pending
**Note:** Depends on Task 8.
