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
**Status:** done
**Plan:** Add FareTrend struct + ComputeFareTrend in search/search.go as pure function on []Itinerary. Add formatFareTrend in cmd/display.go. Include in chat summary and display in chatLoop when FlexDays>0.
- [x] Define FareTrend struct and ComputeFareTrend in search/search.go
- [x] Write tests for ComputeFareTrend (multi-date, empty, single-date)
- [x] Add formatFareTrend in cmd/display.go + 3 tests
- [x] Include fare trend in chat result summary (chathelpers.go)
- [x] Display fare trend in chatLoop after results (chat.go)
- [x] Run go build && go test ./... && go vet ./...

## Task 172: Ranker cache max size with eviction
**Status:** done (completed in worktree)
- [x] Write tests for cache eviction at maxCacheSize
- [x] Add maxCacheSize const and keys slice to Ranker
- [x] Implement eviction in Rank method when inserting new entry
- [x] Verify CacheStats still accurate after eviction
- [x] Run go build && go test ./... && go vet ./...

## Task 173: Context-aware ranking weight adjustment from conversation
**Status:** done
**Plan:** Add contextWeights function in chathelpers.go that scans user messages for preference phrases and returns additive RankingWeights deltas. Wire into chatLoop to apply deltas before search via weightsUpdater. Tests in chat_test.go.
- [x] Write 7 tests for contextWeights (layover, carbon, schedule, duration, no-signal, ignore-assistant, multiple)
- [x] Implement contextWeights + addWeights in cmd/chathelpers.go
- [x] Wire into chatLoop to apply combined (base + context delta) weights before each search
- [x] Update existing TestChatLoop_ProfileSwitchUpdatesWeights for new behavior
- [x] Run go build && go test ./... && go vet ./...

## Task 174: Europe-to-Asia stopover corridors
**Status:** done (completed in worktree)
- [x] Add LHR→BKK stopover corridor
- [x] Add LHR→NRT stopover corridor
- [x] Add CDG→BKK stopover corridor
- [x] Add CDG→NRT stopover corridor
- [x] Add specific lookup tests for new corridors
- [x] Run go build && go test ./... && go vet ./...
