# TODO

## Task 1: Gate live API tests behind integration build tag
**Status:** pending
**Plan:**
- [ ] Add `//go:build integration` tag to `provider/serpapi/seed_cache_test.go`
- [ ] Add `//go:build integration` tag to `search/multicity/diagnostic_test.go`
- [ ] Add `//go:build integration` tag to `search/multicity/multicity_test.go`
- [ ] Verify `go test ./...` passes without API keys
- [ ] Verify `go build ./... && go vet ./...` still clean

## Task 2: Add unit tests for search/filter.go
**Status:** pending
**Plan:**
- [ ] Read `search/filter.go` to understand function signatures and logic
- [ ] Write `search/filter_test.go` with tests for all exported functions
- [ ] Run tests and verify they pass
- [ ] Check coverage meets 80%+ target

## Task 3: Add unit tests for types/types.go
**Status:** pending
**Plan:**
- [ ] Read `types/types.go` to understand type definitions and methods
- [ ] Write `types/types_test.go` with tests for `IsRoundTrip`, `Stops`, `Error`
- [ ] Run tests and verify they pass
- [ ] Check coverage meets 100% target

## Task 4: Add Strategy interface and common Request type
**Status:** pending
**Plan:**
- [ ] Read issue #2 details and existing `search/` package
- [ ] Write `search/strategy_test.go` first (TDD)
- [ ] Create `search/strategy.go` with `Strategy`, `Request`, `Ranker` types
- [ ] Run tests and verify they pass
- [ ] Run full build gate: `go build ./... && go test ./... && go vet ./...`

## Task 5: Add multicity Strategy adapter
**Status:** pending
**Plan:**
- [ ] Read issue #4 details and existing `search/multicity/` code
- [ ] Write `search/multicity/strategy_test.go` first (TDD)
- [ ] Create `search/multicity/strategy.go` adapter
- [ ] Run tests and verify they pass
- [ ] Run full build gate: `go build ./... && go test ./... && go vet ./...`
