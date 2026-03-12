# TODO

## Task 170: Result diversification for search output
**Status:** done
**Plan:** Add DiversifyResults in search/filter.go that selects: 1 cheapest, 1 fastest, 1 best-scored (if scored), then fills remaining slots maximizing airline diversity. Wire into direct.go and composite.go before MaxResults cap. Tests in filter_test.go.
- [x] Write tests for DiversifyResults with near-identical itineraries
- [x] Implement DiversifyResults in search/filter.go
- [x] Wire into direct.go Search after ranking, before MaxResults cap
- [x] Wire into composite.go Search after ranking, before MaxResults cap
- [x] Run go build && go test ./... && go vet ./...

## Task 171: Fare trend summary for flex-date searches
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Define FareTrend struct in search/types.go
- [ ] Implement ComputeFareTrend from itineraries
- [ ] Wire into direct.go to compute trend when FlexDays > 0
- [ ] Add formatFareTrend display helper in cmd/display.go
- [ ] Include fare trend in chat result summary (chathelpers.go)
- [ ] Run go build && go test ./... && go vet ./...

## Task 172: Ranker cache max size with eviction
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Write tests for cache eviction at maxCacheSize
- [ ] Add maxCacheSize const and keys slice to Ranker
- [ ] Implement eviction in Rank method when inserting new entry
- [ ] Verify CacheStats still accurate after eviction
- [ ] Run go build && go test ./... && go vet ./...

## Task 173: Context-aware ranking weight adjustment from conversation
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Write tests for contextWeights with various user phrases
- [ ] Implement contextWeights in cmd/chathelpers.go
- [ ] Wire into chatLoop to apply weight deltas before each search
- [ ] Test integration with chatLoop mock
- [ ] Run go build && go test ./... && go vet ./...

## Task 174: Europe-to-Asia stopover corridors
**Status:** pending
**Plan:** to be filled during implementation
- [ ] Add LHR→BKK stopover corridor
- [ ] Add LHR→NRT stopover corridor
- [ ] Add CDG→BKK stopover corridor
- [ ] Add CDG→NRT stopover corridor
- [ ] Add specific lookup tests for new corridors
- [ ] Run go build && go test ./... && go vet ./...
