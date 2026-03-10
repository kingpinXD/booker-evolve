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

## Suggested for next session

## Task 6: Add unit tests for provider/kiwi/parser.go
**Status:** pending
**Plan:**
- [ ] Read parser.go to understand parsing logic
- [ ] Write parser_test.go with table-driven tests using sample JSON
- [ ] Target 80%+ coverage

## Task 7: Add unit tests for search/multicity/combiner.go
**Status:** pending
**Plan:**
- [ ] Read combiner.go (CombineLegs, hasLongLayover, etc.)
- [ ] Write combiner_test.go with edge cases for layover validation
- [ ] Target 80%+ coverage

## Task 8: Implement direct search Strategy (issue #3)
**Status:** pending

## Task 9: Wire Strategy picker into cmd/search.go (issue #5)
**Status:** pending
