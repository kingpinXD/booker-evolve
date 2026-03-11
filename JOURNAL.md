# JOURNAL

> **APPEND-ONLY** — Only add new entries at the bottom. Never edit or remove existing entries.

Format: `## Session N -- HH:MM -- title` followed by 2-4 sentences.

---

## Day 0 -- 00:00 -- Bootstrap

The evolution system was initialized. Governance documents, skills, scripts, and CI/CD workflows were created. The agent is ready for its first autonomous session.

## Day 0 -- 10:21 -- Self-assessment and session planning

Ran self-assessment: build, tests, vet, and lint all pass. Total test coverage is 6.0% across the project, with 11 of 14 packages having zero test files. Identified four tasks for the next session: gating live API tests behind an integration build tag (issue #1), adding unit tests for search/filter.go and types/types.go, and defining the Strategy interface (issue #2). No code changes were made; this session produced SESSION_PLAN.md only.

## Day 1 -- 00:00 -- Execute all 5 planned tasks

Completed all 5 tasks from SESSION_PLAN.md. (1) Gated 3 integration test files behind `//go:build integration` — `go test ./...` now passes without API keys. (2) Added search/filter_test.go with 7 tests, 100% coverage on filter.go. (3) Added types/types_test.go with 3 tests, 100% coverage on types.go. (4) Created search/strategy.go with Strategy, Request, and Ranker types (issue #2). (5) Created search/multicity/strategy.go adapter wrapping Searcher as search.Strategy (issue #4). Total coverage rose from 6.0% to 11.2%. All build gates passed on every commit.

## Day 1 -- 12:17 -- Session wrap-up and next-session planning

All 5 planned tasks completed with zero reverts. Coverage rose from 6.0% to 11.2% across 5 commits. The two highest-value changes were the Strategy interface (issue #2) and the multicity adapter (issue #4), which establish the abstraction layer for plugging in direct-search and other strategies later. Next session should focus on testing pure-function packages (kiwi parser, combiner) and wiring the Strategy picker into cmd/search.go.

## Day 3 -- 13:30 -- Session planning and self-assessment

Ran self-assessment: build, tests, vet, and lint all pass. Coverage unchanged at 11.2%. Reverted a stale whitespace-only diff in search/strategy.go. Planned four tasks: unit tests for provider/kiwi/parser.go (Task 6), unit tests for search/multicity/combiner.go (Task 7), direct search Strategy implementation (Task 8, deferred), and Strategy picker wiring (Task 9, deferred). No code changes this session -- planning only.

## Day 4 -- 00:00 -- Execute combiner tests, skip kiwi, fix lint

Self-assessment found 1 gofmt lint issue in search/strategy.go (extra whitespace alignment). Fixed and committed. Skipped Task 6 (kiwi parser tests) per CLAUDE.md directive to ignore the Kiwi provider entirely. Completed Task 7: added 28 table-driven tests for combiner.go covering CombineLegs, hasLongLayover, lastArrival, firstDeparture, buildItinerary, PrimaryAirline, and SameAirline. combiner.go functions at 96-100% coverage. Total project coverage rose from 11.2% to 17.2%. Zero reverts, zero API calls.

## Day 4 -- 13:41 -- Session close and next-session handoff

Day 4 produced 3 commits: gofmt fix (14a4130), combiner tests (b5a74ad), and session wrap-up (04d2551). Coverage rose from 11.2% to 17.2% with zero reverts and zero API calls. Tasks 8 (direct search Strategy, issue #3) and 9 (Strategy picker wiring, issue #5) remain deferred as the highest-priority items for Day 5.

## Day 7 -- 00:00 -- Complete strategy system (issues #3, #5, #6)

Completed all 3 planned tasks with zero reverts and zero API calls. (1) Task 8: created search/direct/ package implementing search.Strategy for simple origin-to-destination flights with filter pipeline and optional LLM ranking (73b5a95). (2) Task 11: created search/picker.go with LLM-based strategy selection and heuristic fallback (aa17e5e). (3) Task 9: rewired cmd/search.go to use Picker with direct + multicity strategies, added --context flag, made --leg2-date optional, fixed routeString for 1-leg itineraries (0df7174). Coverage rose from 17.2% to 19.4%. The strategy system is now fully wired end-to-end: issues #1-#6 are all addressed.

## Day 6 -- 14:02 -- Session close and strategy system handoff

All 6 open GitHub issues (#1-#6) are now addressed across Days 1-7. The strategy system is fully wired: Picker selects between direct and multicity strategies via LLM or heuristic fallback, and the CLI exposes it through --context. Coverage stands at 19.4%. Next priorities: test the LLM code paths in picker.go (currently untested), add integration tests for the end-to-end CLI flow, and expand route coverage with new search strategies.

## Day 8 -- 14:40 -- Self-assessment and session planning for Day 9

Ran self-assessment: build, tests, vet, and lint all pass. Coverage at 19.4% with 6 of 14 packages having tests. All 6 GitHub issues (#1-#6) remain addressed. Planned 5 tasks for Day 9: extract ChatCompleter interface and test picker.go LLM paths (high priority), expand cache tests for multi-city paths, test currency.go, test config.go, and commit pending whitespace fixes. No code changes this session.

## Day 9 -- 15:15 -- Evolution system refactor, no codebase tasks

Session focused on improving the evolution process itself rather than codebase tasks. One commit (950b592) refactored scripts/evolve.sh and skills/evolve/SKILL.md to support parallel agents via git worktrees, 25-minute/70%-context session limits, and one commit per task. No planned codebase tasks (ChatCompleter interface, picker tests, cache tests, etc.) were executed. Coverage unchanged at 19.4%.

## Day 10 -- 21:30 -- Execute all 5 deferred tasks, coverage 19.4% to 28.3%

Completed all 5 tasks from SESSION_PLAN.md in 5 commits with zero reverts and zero API calls. (1) Extracted ChatCompleter interface in picker.go and added 6 LLM-path tests — picker coverage at 97.4%. (2) Added 5 aggregator tests with mock providers and race detector — aggregator at 100%. (3) Refactored currency.go to Converter struct for testability, added 5 tests — currency at 36.7%. (4) Added 4 multi-city cache tests — cache coverage rose from 38% to 77.5%. (5) Added 3 config.Default() tests — config at 100%. Total project coverage rose from 19.4% to 28.3%. Nine packages now have tests (up from 6).

## Day 11 -- 22:45 -- Execute all 5 planned tasks, coverage 28.3% to 34.5%

Completed all 5 tasks in 5 commits with zero reverts and zero API calls. (1) Fixed gofmt whitespace in currency.go. (2) Added provider.Registry tests — 100% coverage on a previously untested foundational package. (3) Added ranker pure-function tests for parseRankingResponse, formatDuration, buildSystemPrompt, buildRankingPrompt — multicity coverage 20.4% to 33.2%. (4) Expanded search/direct tests from 76.5% to 94.1% covering invalid dates, zero prices, MaxStops, ranker mock, provider errors, MaxResults cap, and flightToItinerary. (5) Added stopovers data integrity tests. Ten packages now have tests (up from 9). Next priorities: new features from GitHub issues, multicity search coordinator tests, currency HTTP mock tests.

## Day 12 -- 00:00 -- Execute all 5 planned tasks, coverage 34.5% to 46.5%

Completed all 5 tasks in 5 commits with zero reverts and zero API calls. (1) Fixed gofmt on 2 files and 3 errcheck lint violations -- lint now reports 0 issues. (2) Added FlexDays date range filtering to direct search via TDD -- 2 new tests, direct coverage 94.1% to 94.7%. (3) Added 13 httpclient unit tests with httptest covering GET/POST, retry on 500, no-retry on 4xx, BuildURL, connection errors -- httpclient coverage 0% to 89.4%. (4) Added 11 SerpAPI parser tests -- parser.go at 100%, package-wide 17.8% to 37.7%. (5) Added 22 multicity helper tests for deduplicateFlights, buildMultiCityItinerary, fetchFromAllProviders, fetchWithDualSort -- multicity coverage 33.2% to 48.6%. Total project coverage rose from 34.5% to 46.5%. Eleven packages now have tests (up from 10). Tasks 3-5 ran as parallel worktree agents for efficiency. Next priorities: new features from GitHub issues, currency HTTP mock tests, llm package tests.

## Day 13 -- 01:15 -- Execute all 5 planned tasks, coverage 46.5% to 73.2%

Completed all 5 tasks in 5 commits with zero reverts and zero API calls. (1) Added 11 llm unit tests with httptest -- llm coverage 0% to 100%. (2) Added 6 currency fetchRates HTTP mock tests -- currency coverage 36.7% to 90.0%. (3) Added 12 serpapi provider tests with httptest covering Search, SearchMultiCity, mapCabinClassToSerpAPI, buildMultiCityJSON -- serpapi coverage 37.7% to 92.5%. (4) Added 12 multicity Search orchestrator tests with mock providers and LLM -- multicity coverage 48.6% to 94.6%. (5) Added 8 cache edge case tests for TTL expiry, corrupted files, error propagation, unwritable dirs -- cache coverage 77.5% to 91.5%. Total project coverage rose from 46.5% to 73.2%. Twelve packages now have tests (up from 11). Tasks 1-3 and 4-5 ran as parallel worktree agents. Next priorities: new features from GitHub issues, or cmd package tests.

## Day 14 -- 02:40 -- Execute all 5 planned tasks, 3 new features + lint + tests

Completed all 5 tasks in 5 commits with zero reverts and zero API calls. (1) Fixed gofmt violation, 9 errcheck lint issues across 3 test files, and removed duplicate FilterZeroPrices comment -- lint now reports 0 issues. (2) Added adaptive table columns: single-leg results show compact layout (Airlines, Departure, Duration) while multi-leg keeps full detail with Stopover and Leg 2 columns; Duration column added to both. (3) Added --format json output mode with currency-converted prices and leg details for programmatic consumption. (4) Added 18 cmd unit tests for routeString, legAirlines, legDeparture, stopoverString, currencySymbol, formatDuration, isMultiLeg, printJSON -- cmd coverage 0% to 54.5%. (5) Added --verbose/-v flag to conditionally show debug logs. Thirteen packages now have tests (up from 12). All build gates passed on every commit. Next priorities: more search strategies, route expansion, flexible date features.

## Day 12 -- 22:11 -- 4 features + tests in 2 commits, cmd coverage 54.5% to 68.2%

Completed all 5 planned tasks in 2 commits with zero reverts and zero API calls. (1) Added PriceInsights type to search package with ParsePriceInsights in serpapi parser and LastPriceInsights getter on Provider -- data is now extracted from SerpAPI responses though not yet cached. (2) Added booking_url field to JSON output, surfacing types.Flight.BookingURL that was populated but never shown. (3) Added priceSummary footer to table output showing result count and price range (e.g. "5 results | $450 - $1,200"). (4) Added 11 new cmd tests: priceSummary (3), booking URL in JSON (1), printTable stdout capture for single-leg and multi-leg (2), bringing cmd coverage from 54.5% to 68.2%. (5) Expanded stopover routes: added BOM to YYZ (7 cities) and DEL to YVR (6 cities) corridors with geographically appropriate stopovers; fixed StopoversForRoute to return nil for unknown routes instead of falling back to DEL to YYZ. All build gates passed on every commit.

## Day 16 -- 04:15 -- CompositeStrategy, Picker 'both' mode, and chat command

Completed all 5 tasks in 4 commits with zero reverts and zero API calls. (1) Fixed gofmt alignment in PriceInsights struct. (2) Added CompositeStrategy in search/composite.go -- runs multiple strategies concurrently via sync.WaitGroup, merges results, deduplicates by route+price, optionally re-ranks. 10 tests with race detector. (3) Extended Picker to support "both" strategy: when LLM returns {"strategy":"both"}, Picker wraps all strategies in a CompositeStrategy. System prompt updated, 3 new tests. (4-5) Added `booker chat` command implementing VISION.md's top priority: conversational flight search. Multi-turn LLM loop gathers trip parameters through dialogue, parses extracted JSON into search.Request, runs Picker + strategy, displays results. 10 new tests covering param parsing, defaults, system prompt, and full conversation-to-search flow. Zero SerpAPI or LLM calls -- all tests use mocks. This is the first step toward booker as a booking agent rather than a search tool.

## Day 18 -- Session 18, Task 1 -- Nearby-airport search strategy
Created search/nearby/ package with NearbySearcher strategy. Wraps a delegate strategy, expands origin/destination via airport clusters, fans out concurrent searches for all airport-pair combinations, merges and deduplicates by route+price, sorts by price. 9 tests covering fan-out, dedup, MaxResults cap, no-cluster fallback, partial errors. Ran as parallel worktree agent.

### Session 18, Task 2 -- Round-trip support in direct strategy
Added round-trip support to direct.Search(). Extracted searchFlights helper to avoid duplicating the date-expansion/fetch/filter pipeline. When ReturnDate is non-empty, searches return flights (dest->origin) and combines all outbound x return pairs into 2-leg itineraries with summed prices and computed TotalTrip. One-way behavior unchanged. 2 new tests with routeProvider mock. Ran as parallel worktree agent.

### Session 18, Task 3 -- Extract shared cmd infrastructure
Created cmd/infra.go with buildPicker(weights, leg2Date) helper. Both runSearch and runChat now call this instead of repeating ~20 lines of provider/strategy/picker wiring. Net -2 lines. All existing cmd tests pass unchanged.

### Session 18, Task 4 -- Structured refinement guidance in chat
Added refinementHint() function listing available levers (dates, nearby airports, cabin class, direct-only, passengers, round-trip). Appended as a system message to conversation history after search results, so the LLM knows what refinement options to suggest. 2 new tests.

### Session 18, Task 5 -- Lint and gofmt sweep
Fixed 1 gofmt violation in search/direct/direct.go (tab alignment from worktree agent). Zero lint issues after fix. All build gates pass.

## Day 19 -- Session 19 -- CLI features: return-date, nearby, max-price, price insights, multicity test

Completed all 5 tasks in 5 commits with zero reverts and zero API calls. (1) Added --return-date flag to search command, wiring to existing round-trip support. (2) Wired NearbySearcher into buildPicker as third strategy -- LLM can now select "nearby". (3) Added --max-price budget filter end-to-end: FilterByMaxPrice filter (TDD, 3 tests), direct pipeline, CLI flag, chat tripParams + system prompt. (4) Surfaced PriceInsights in table and JSON output: modified buildPicker to return raw provider, added formatPriceInsights helper (TDD, 2 tests), printJSONWithInsights, refactored printJSON to shared buildJSONItineraries. (5) Added 2 tests for multicity.Strategy.Search (ran as parallel worktree agent). Two gofmt fixes for struct alignment. All build gates pass.

### Session 19, Task 4 -- Surface PriceInsights in output
Modified buildPicker to return raw *serpapi.Provider so callers can access LastPriceInsights(). Added formatPriceInsights (TDD: 2 tests) for one-line display, printJSONWithInsights with price_insights field, and refactored printJSON to reuse buildJSONItineraries. Price insights now shown below table output and in JSON when available. 1 gofmt fix.

### Session 19, Task 3 -- Add --max-price budget filter
Added MaxPrice field to search.Request, FilterByMaxPrice filter function with TDD (3 test cases), wired into direct.searchFlights pipeline and CLI (--max-price flag). Also added max_price to chat tripParams and system prompt so the LLM can extract budget constraints from conversation. 1 gofmt fix for struct alignment.

### Session 19, Task 5 -- Test multicity.Strategy.Search
Added 2 tests to strategy_test.go using existing test helpers. TestStrategy_Search verifies happy path (mock provider + LLM ranker), TestStrategy_Search_Error verifies error propagation from invalid date. Ran as parallel worktree agent. Coverage for Strategy.Search now > 0%.

### Session 19, Task 2 -- Wire NearbySearcher into buildPicker
Imported nearby package in cmd/infra.go, wrapped directStrategy with nearby.NewSearcher, and passed it as the third strategy to NewPicker. The LLM picker can now select "nearby" when users want to compare metro-area airports. 2-line change, all tests pass.

### Session 19, Task 1 -- Add --return-date flag to search command
Added keyReturnDate const and --return-date flag to searchCmd. Wired to req.ReturnDate in runSearch so CLI users can now trigger round-trip searches that the direct strategy already supports. Minimal 3-line change, all tests pass.

## Day 17 -- 04:30 -- Lint fix, chat refinement, airport clusters, flex-date multi-search

Completed all 5 tasks in 5 commits with zero reverts and zero API calls. (1) Fixed 7 errcheck lint violations in cmd/chat.go. (2) Added result summary to chat conversation history -- after displaying search results, a summary (count + price range) is appended as an assistant message so the LLM has context for refinement requests. 2 new tests. (3) Added airport cluster data in search/airports.go -- 14 metro-area clusters with NearbyAirports O(1) lookup via reverse index, 4 tests. (4) Enhanced direct strategy to search multiple dates when FlexDays > 0, making 2*flex+1 provider calls instead of 1, genuinely finding cheaper options on adjacent dates. 2 new tests with dateTrackingProvider mock. Tasks 3 and 4 ran as parallel worktree agents. (5) Surfaced nearby-airport suggestions in chat -- system prompt now mentions alternatives, and nearbyAirportHint displays tips after param extraction. 2 new tests.

### Day 20, Task 1 -- Chat param merging for follow-up searches
Added mergeParams and parsePartialParams to enable chat refinement loop. chatLoop now stores lastParams after each search; when the LLM emits partial JSON (e.g. just cabin change), it merges with previous params and re-searches. Updated refinementHint to instruct the LLM to re-emit only changed fields. 3 new test functions (mergeParams table-driven, parsePartialParams, chatLoop follow-up integration).

### Day 20, Task 2 -- Dynamic ranking profile from conversation
Added profileWeights function reusing existing profiles map. Added Profile field to tripParams, --profile flag to chatCmd, profile mention in system prompt and refinement hint. runChat now uses the flag-specified profile instead of hardcoded WeightsBudget. 1 new test function (5 subtests for all profiles + unknown default).

### Day 20, Task 3 -- Airline alliance reference data
Added search/airlines.go with Star Alliance (28), OneWorld (14), and SkyTeam (18) member IATA codes. Alliance() and SameAlliance() functions with O(1) lookup via map. 8 tests in search/airlines_test.go. Ran as parallel worktree agent.

### Day 20, Task 4 -- Richer result context in chat history
Enhanced resultSummaryForChat to accept tripParams and include top result's route, airline, duration, and price, plus search parameters (origin, dest, date, cabin, max_price). The LLM can now explain why a specific result was recommended. Added formatFlightDuration helper. Updated test + added empty-results test.

### Day 20, Task 5 -- Lint, gofmt sweep, and build gate
All gates clean: gofmt -l returns empty, go vet 0 issues, golangci-lint 0 issues, go test ./... all 15 packages pass.

## Day 21 -- Session 21 -- Alliance tags, stopover notes, direct-only chat, reasoning output

Completed all 5 tasks in 5 commits with zero reverts and zero API calls. Tasks 33 and 34 ran as parallel worktree agents while tasks 31 and 32 ran sequentially on main.

### Session 21, Task 1 -- Alliance-aware ranking in multicity
Wired search.Alliance() into buildRankingPrompt so the LLM sees airline alliance membership (Star Alliance, OneWorld, SkyTeam) next to each segment's airline name. Previously the system prompt mentioned "same alliance is good" but the LLM had no alliance data. Removed resolved TODO in combiner.go. 2 new tests.

### Session 21, Task 2 -- Stopover city notes in ranker prompt
Added Notes field to search.Stopover, passed StopoverCity.Notes through buildItinerary in combiner.go, and displayed notes in buildRankingPrompt after the stopover line. The LLM now has context about connectivity, food, visa, and airline hubs when scoring stopover cities. Removed resolved TODO in combiner.go. 2 new tests.

### Session 21, Task 3 -- Direct-only preference in chat
Added DirectOnly bool to tripParams. When set, buildRequestFromParams sets MaxStops=0. Added to system prompt, parsePartialParams recognition, and mergeParams. Ran as parallel worktree agent. 5 new test assertions.

### Session 21, Task 4 -- Surface ranking reasoning in output
Added "Reason" column to table output and "reasoning" field to jsonItinerary. The ranker's Reasoning string was already populated but never displayed. Ran as parallel worktree agent. 4 new tests.

### Session 21, Task 5 -- Lint, gofmt sweep, and build gate
All gates clean after merging worktree branches: gofmt -l empty, go vet clean, golangci-lint 0 issues, go test all pass.

## Day 22

### Session 22, Task 1 -- Fix chat output routing to use io.Writer
Added `io.Writer` parameter to `printTable`, `printJSON`, and `printJSONWithInsights`. chatLoop now routes all search output through its `out` writer instead of hardcoding os.Stdout. Removed os.Pipe hacks from tests in favor of direct buffer writes. Added new test confirming chatLoop output buffer captures table data.

### Session 22, Task 2 -- Show booking URLs in table output
Added "Book" column to both single-leg and multi-leg table layouts. Added `legBookingURL` helper following existing `legAirlines`/`legDeparture` pattern. 3 new tests. Ran as parallel worktree agent.

### Session 22, Task 3 -- Conversation history truncation
Added `truncateHistory` sliding window keeping system prompt + last 20 messages. Wired into chatLoop before each ChatCompletion call. 3 new tests including integration test verifying long conversation stays within bounds. Ran as parallel worktree agent.

### Session 22, Task 4 -- Add stops count to table output
Added `itineraryStops` helper that sums `Flight.Stops()` across all legs. Added "Stops" column to both single-leg and multi-leg table layouts. 1 new test with direct (0) and connecting (1) flights.

### Session 22, Task 5 -- Lint, gofmt sweep, and build gate
All gates clean: gofmt -l empty, go vet clean, golangci-lint 0 issues, go test all pass. One gofmt fix was needed after worktree rebase (fixed during task 37/38 merge).

## Day 23

### Session 23, Task 1 -- Flex-days support in chat
Added flex_days field to tripParams with full wiring: buildRequestFromParams uses it (defaults to 3 when zero), mergeParams preserves it, parsePartialParams recognizes it, system prompt documents it, refinement hint mentions it. 5 new tests. Ran as parallel worktree agent.

### Session 23, Task 2 -- Inject today's date into chat system prompt
Changed chatSystemPrompt() to accept time.Time parameter, prepends "Today's date is YYYY-MM-DD" so the LLM can handle temporal references like "next Friday". Updated chatLoop to pass time.Now(). Updated existing tests to use known dates. 1 new test + 2 updated. Ran as parallel worktree agent.

### Session 23, Task 3 -- Show layover durations in stops column
Added formatStops(itin) that shows layover time alongside stop count (e.g. "1 (2h 30m)"). Sums LayoverDuration from all segments when stops > 0. Falls back to plain count when no layover data. Replaced itineraryStops int with formatStops string in printTable. 3 new tests. Ran as parallel worktree agent.

### Session 23, Task 4 -- Add arrival time column to table output
Added legArrival helper returning last segment's ArrivalTime. Added "Arrival" column to single-leg table and "Leg 1 Arrival"/"Leg 2 Arrival" to multi-leg table. 2 new tests + 2 updated table tests. Ran as parallel worktree agent.

### Session 23, Task 5 -- Lint, gofmt sweep, and build gate
All gates clean after rebasing both worktree branches: gofmt -l empty, go vet clean, golangci-lint 0 issues, go test all 15 packages pass.

## Session 23 -- Chat UX and table display improvements

Completed all 4 planned tasks plus lint sweep in 5 commits with zero reverts and zero API calls. All feature tasks ran as parallel worktree agents. (1) Added flex_days field to chat tripParams so users can control date flexibility in conversation -- wired through buildRequestFromParams, mergeParams, parsePartialParams, system prompt, and refinement hint. (2) Injected today's date into chatSystemPrompt via time.Time parameter so the LLM can resolve temporal references like "next Friday". (3) Enhanced stops column to show layover durations (e.g. "1 (2h 30m)") using segment LayoverDuration data that was already populated but never surfaced. (4) Added arrival time column to table output using last segment's ArrivalTime. Coverage at 83.3% across 15 packages. All build gates pass.

## Day 24

### Session 24, Task 1 -- Display cabin class in table and JSON output
Added legCabin helper returning string(first segment's CabinClass). Added "Cabin" column to single-leg and multi-leg table layouts. Added cabin_class field to jsonLeg struct. 4 new tests. Ran as parallel worktree agent.

### Session 24, Task 2 -- Parse and display carbon emissions from SerpAPI
Added CarbonEmissions struct to SerpAPI response.go (this_flight, typical_for_this_route, difference_percent). Added CarbonKg field to types.Flight. Parser converts grams to kg. Added legCarbon helper and "CO2" column to table, carbon_kg to JSON. 8 new tests across serpapi and cmd. Ran as parallel worktree agent.

### Session 24, Task 3 -- Add red-eye detection to ranker prompt
Added isRedEye(t time.Time) bool for departures 00:00-04:59. buildRankingPrompt now appends [Red-eye] tag after airline info for red-eye flights, giving the LLM explicit signal to penalize. 6 new tests. Ran as parallel worktree agent.

### Session 24, Task 4 -- Lint, gofmt sweep, and build gate
One gofmt fix in response.go (tab alignment in CarbonEmissions struct). All gates clean: gofmt -l empty, go vet clean, golangci-lint 0 issues, go test all 15 packages pass.

## Day 25

### Session 25, Task 1 -- Parse overnight flag and annotate in ranker prompt
Added Overnight bool to types.Segment, parsed from SerpAPI FlightSegment.Overnight in parser.go. buildRankingPrompt now appends [Overnight] tag after airline info (similar to [Red-eye] tag), giving the LLM explicit signal about overnight connections. 4 new tests across serpapi and multicity. Ran as parallel worktree agent.

### Session 25, Task 2 -- Parse aircraft type and display in JSON output
Added Aircraft string to types.Segment, parsed from SerpAPI FlightSegment.Airplane. Added legAircraft helper and wired into buildJSONItineraries as "aircraft" field (omitted when empty). 3 new tests. Sequential on main since it shares files with Task 1.

### Session 25, Task 3 -- Conditional Score/Reason columns + carbon rounding fix
Added hasScores helper to detect when any itinerary has a non-zero score. printTable now conditionally includes Score and Reason columns only when scores exist, reducing visual noise in direct search output. Also fixed carbon emissions integer division bug: changed grams/1000 to (grams+500)/1000 so 800g correctly rounds to 1kg instead of truncating to 0. 5 new tests. Ran as parallel worktree agent.

### Session 25, Task 4 -- Lint, gofmt sweep, and build gate
All gates clean: gofmt -l empty, go vet clean, golangci-lint 0 issues, go test all 15 packages pass.

## Session 24 -- 05:22 -- Ranker enrichment and table output polish

Completed all 4 planned tasks from the Day 25 session plan in 4 commits with zero reverts and zero API calls. (1) Parsed SerpAPI Overnight bool into types.Segment and added [Overnight] tag to buildRankingPrompt, giving the LLM explicit signal about overnight connections alongside the existing [Red-eye] tag. (2) Parsed aircraft type (Airplane field) from SerpAPI into types.Segment.Aircraft and surfaced it in JSON output as "aircraft" (omitempty). (3) Added hasScores helper to conditionally hide Score/Reason columns when no ranker is used, reducing table noise for direct search output; also fixed carbon emissions integer division bug by switching grams/1000 to (grams+500)/1000. (4) All lint/gofmt/vet gates clean. Coverage at ~84% across 15 packages. All tasks used mocks and existing cached data -- zero SerpAPI or LLM calls.

## Session 26 -- Data enrichment and output completeness

### Session 26, Task 1 -- Parse legroom from SerpAPI and annotate in ranker
Added Legroom string field to types.Segment, wired from SerpAPI FlightSegment.Legroom in parser.go, added legroom to JSON output (omitempty), and added [Legroom: X] tag to buildRankingPrompt. All tests pass.

### Session 26, Task 2 -- Enrich JSON output with arrival time, stops, omit zero score
Added Arrival (RFC3339) and Stops fields to jsonLeg in JSON output. Added omitempty to Score in jsonItinerary so unranked results produce cleaner JSON. Four new tests cover arrival, stops with connections, and score omitempty behavior.

### Session 26, Task 3 -- Wire PriceInsights into chat output
Threaded priceInsighter interface through chatLoop so chat mode displays price level context after results (table and JSON). Previously discarded rawProvider is now captured and passed through. Updated all existing chat tests for new signature.

### Session 26, Task 4 -- Fix multi-leg CO2 display
Replaced single "CO2" column with "Leg 1 CO2" and "Leg 2 CO2" in the multi-leg table layout. Previously only leg 0's carbon data was shown. Minimal 3-line change in search.go plus one new test.

### Session 26, Task 5 -- Lint, gofmt sweep, and build gate
All gates clean: gofmt -l empty, go vet clean, golangci-lint 0 issues, go test all 15 packages pass.

## Session 25 -- 06:28 -- Data enrichment and output completeness

Completed all 4 planned tasks plus lint sweep in 5 commits with zero reverts and zero API calls. (1) Parsed legroom string from SerpAPI into types.Segment.Legroom, surfaced in JSON output (omitempty), and annotated in buildRankingPrompt with [Legroom: X] tag so the LLM can factor comfort into scoring. (2) Enriched JSON output with arrival time (RFC3339) and stops count per leg, and added omitempty to Score so unranked results produce cleaner JSON. (3) Wired PriceInsights into chat mode by threading a priceInsighter interface through chatLoop -- previously the raw provider was discarded in runChat. (4) Fixed multi-leg table to show separate "Leg 1 CO2" and "Leg 2 CO2" columns instead of a single "CO2" column that only displayed leg 0 data. Coverage steady at ~84% across 15 packages. All build gates pass.

## Day 27

### Session 27, Task 1 -- Add aircraft and carbon annotations to ranker prompt
Added [Aircraft: X] tag per segment and CO2: Xkg line per leg in buildRankingPrompt. These were already parsed from SerpAPI but never shown to the LLM ranker. 4 new tests. Ran as worktree agent.

### Session 27, Task 2 -- Parse carbon benchmark data from SerpAPI, surface in JSON and ranker
Threaded TypicalForThisRoute and DifferencePercent from SerpAPI CarbonEmissions end-to-end: types.Flight gets TypicalCarbonKg and CarbonDiffPct, parser.go extracts with rounding, JSON output includes typical_carbon_kg and carbon_diff_percent (omitempty), ranker CO2 line shows benchmark comparison (e.g. "CO2: 1106kg (+17% vs typical)"). 6 new tests across serpapi, cmd, and multicity.

### Session 27, Task 3 -- Lint, gofmt sweep, and build gate
Two gofmt struct alignment fixes (types.go, search.go). All gates clean: gofmt -l empty, go vet clean, golangci-lint 0 issues, go test all 15 packages pass.

## Session 27 -- Ranker enrichment with aircraft type and carbon benchmarks

Completed all 3 planned tasks in 3 commits with zero reverts and zero API calls. (1) Added [Aircraft: X] tag and CO2: Xkg line to buildRankingPrompt, giving the LLM explicit signals for comfort (equipment type) and environmental impact that were previously parsed but never passed to the ranker. (2) Threaded SerpAPI carbon benchmark data (TypicalForThisRoute, DifferencePercent) end-to-end: types.Flight fields, parser extraction with rounding, JSON output (omitempty), and ranker CO2 line with benchmark comparison ("CO2: 1106kg (+17% vs typical)"). (3) Two gofmt fixes for struct alignment. Coverage steady at ~84% across 15 packages. All build gates pass.

## Day 28

### Session 28, Task 1 -- Expand airport clusters with 8 new metro areas
Added 8 multi-airport metro clusters to airportClusters: Bangkok (BKK/DMK), Istanbul (IST/SAW), Beijing (PEK/PKX), Osaka (KIX/ITM), Rome (FCO/CIA), Taipei (TPE/TSA), Miami (MIA/FLL), Sao Paulo (GRU/CGH/VCP). Now 22 clusters. 17 new test cases. Ran as parallel worktree agent.

### Session 28, Task 2 -- Add alliance preference filter for search and chat
Added PreferredAlliance field to search.Request and FilterByAlliance to filter.go, using existing Alliance() lookup. Wired end-to-end: direct pipeline, chat tripParams, buildRequestFromParams, mergeParams, parsePartialParams, system prompt, refinement hint. 10 new tests across search and cmd packages. Ran as parallel worktree agent.

### Session 28, Task 3 -- Build gate verification
Post-merge verification after rebasing both worktree branches. gofmt clean, go vet clean, golangci-lint 0 issues, all 15 packages pass.

## Session 28 -- Airport cluster expansion and alliance preference filter

Completed all 3 planned tasks in 3 commits with zero reverts and zero API calls. Tasks 1 and 2 ran in parallel worktree agents since they touched different files. (1) Expanded airport clusters from 14 to 22 metros, adding key Asian, European, and American multi-airport cities. The nearby searcher can now find cheaper alternatives at 8 more metros. (2) Added alliance preference filter end-to-end: FilterByAlliance in filter.go, PreferredAlliance on search.Request, preferred_alliance in chat tripParams with full merge/parse support, system prompt, and refinement hint. Users with frequent flyer programs can now filter to their alliance. (3) Clean post-merge build gate. Coverage steady at ~84% across 15 packages. All build gates pass.

## Session 27 -- 07:26 -- Session close and Day 28 handoff

Day 28 delivered two high-impact features across 6 commits (spanning Days 27-28) with zero reverts and zero API calls. The ranker now receives aircraft type, carbon emissions with benchmark comparisons, and the full data pipeline from SerpAPI through types through display is complete for all enrichment fields. Airport clusters expanded from 14 to 22 metros, and alliance preference filtering is wired end-to-end from CLI flags through chat params to the direct search pipeline. Coverage steady at ~84% across 15 packages. No open GitHub issues remain. Next priorities: multi-city alliance filtering (currently only in direct pipeline), cached flight data analysis for route recommendations, or passenger count support.

## Day 29

### Session 29, Task 1 -- Wire PreferredAlliance into multicity strategy
Added PreferredAlliance to multicity.SearchParams, mapped from search.Request in toSearchParams, applied FilterByAlliance in the FILTER stage for both legs and multi-city provider results. 2 new tests in strategy_test.go and search_test.go. Ran as parallel worktree agent.

### Session 29, Task 2 -- Wire MaxPrice into multicity strategy
Added MaxPrice to multicity.SearchParams, mapped from search.Request. Filter applied after COMBINE stage on total itinerary price (not per-flight), and in stage 4b for multi-city results. 2 new tests. Ran in same worktree as Task 1 (sequential, shared files).

### Session 29, Task 3 -- Add departure time preference filter
Added DepartureAfter/DepartureBefore (HH:MM) to search.Request and FilterByDepartureTime in filter.go with graceful degradation for invalid formats. Wired into direct pipeline, chat tripParams (merge/parse/build/system prompt/refinement hint). 9 new tests. Ran as parallel worktree agent.

### Session 29, Task 4 -- Surface SeatsLeft in JSON output and ranker prompt
Added seats_left (omitempty) to jsonLeg with legSeatsLeft helper (min across segments). Added [Seats: N left] tag in buildRankingPrompt so the LLM can factor scarcity into ranking. 6 new tests. Ran as parallel worktree agent.

### Session 29, Task 5 -- Build gate verification and session finalize
Post-merge verification. Fixed gofmt alignment in strategy.go (common with struct field additions during rebase). go vet clean, golangci-lint 0 issues, all 15 packages pass.

## Session 29 -- Multicity consistency, departure time filter, seats display

Completed all 5 planned tasks in 5 commits with zero reverts and zero API calls. Tasks ran in 3 parallel worktree agents (65+66 sequential in one worktree, 67 and 68 in separate worktrees). Key outcomes: (1) PreferredAlliance and MaxPrice now work in multicity strategy, closing the consistency gap with the direct pipeline. (2) Users can filter by departure time-of-day to avoid red-eyes or prefer morning flights. (3) Seat scarcity data (SeatsLeft) is now visible in JSON output and in the LLM ranker prompt. gofmt fix needed after rebase (expected pattern for struct alignment). Coverage steady at ~84%.

## Session 29 -- 08:33 -- Session close and Day 30 handoff

Day 29 delivered 4 features across 5 commits with zero reverts and zero API calls. The multicity strategy now respects PreferredAlliance and MaxPrice, closing the last consistency gap with the direct pipeline. DepartureAfter/DepartureBefore time-of-day filtering enables "no red-eyes" and "morning flights only" use cases end-to-end (direct search + chat). SeatsLeft surfaces in JSON output and the ranker prompt for scarcity-aware recommendations. All 15 packages pass, coverage steady at ~84%, no open GitHub issues. Next priorities: departure time filter in multicity pipeline, passenger count support, or cached flight data analysis for route recommendations.

### Session 31, Task 1 -- Wire FilterByDepartureTime into multicity pipeline
Added DepartureAfter/DepartureBefore to multicity.SearchParams, mapped from search.Request in toSearchParams, applied FilterByDepartureTime in FILTER stage (both legs) and stage 4b (mcItineraries). Follows the exact pattern used for PreferredAlliance. 2 new tests. Ran as parallel worktree agent.

### Session 31, Task 2 -- Pass stops=0 to SerpAPI when direct-only
config.SerpAPIParamStops was defined but never sent. Added stops=0 to the params map when req.MaxStops==0, telling SerpAPI to only return non-stop flights. 2 new tests, updated 1 existing test. Ran as parallel worktree agent.

### Session 31, Task 3 -- Send return_date to SerpAPI for round-trip requests
Added SerpAPIParamReturnDate constant and included return_date in the request when IsRoundTrip() is true. Previously, round-trip type was set but return_date was omitted. 1 new test, updated 1 existing test. Ran as parallel worktree agent.

### Session 31, Task 4 -- Fix empty city names in ranker prompt
SerpAPI never populates OriginCity/DestinationCity, producing "(->)" noise in the ranker prompt. Conditionally omit the city parenthetical when both are empty. 2 new tests. Ran as parallel worktree agent.

### Session 31, Task 5 -- Build gate verification and session finalize
Post-merge verification. Fixed gofmt alignment in config/routes.go (expected pattern when adding constants with different name lengths). go build clean, go vet clean, golangci-lint 0 issues, all 15 packages pass.

## Session 31 -- Multicity departure time, SerpAPI params, ranker cleanup

Completed all 5 planned tasks in 5 commits with zero reverts and zero API calls. Tasks ran in 2 parallel worktree agents (70+73 multicity, 71+72 serpapi). Key outcomes: (1) DepartureAfter/DepartureBefore now work in multicity pipeline, closing the last filter consistency gap. (2) SerpAPI sends stops=0 for direct-only searches, reducing response size. (3) Round-trip requests now include return_date for correct pricing. (4) Ranker prompt no longer shows empty "(->)" city parentheticals. gofmt alignment fix needed after rebase (expected). Coverage steady at ~84%.

## Session 31 -- 09:31 -- Session close and Day 32 handoff

Day 31 delivered 4 features across 5 commits with zero reverts and zero API calls. The multicity pipeline now respects all filters that the direct pipeline does -- DepartureAfter/DepartureBefore was the last gap, closing the filter consistency effort started in Session 29. SerpAPI requests are now more precise: stops=0 narrows API results for direct-only searches, and return_date is included for round-trip pricing correctness. The ranker prompt is cleaner with conditional city-name parentheticals. Coverage steady at ~84% across 15 packages. No open GitHub issues. Next priorities: passenger count support, cached flight data analysis for route recommendations, or time-zone-aware arrival display.

## Day 33

### Session 33, Task 1 -- Post-fetch sorting in direct strategy + CLI flag
Added SortBy field to search.Request and SortResults function in filter.go (price/duration/departure modes). Replaced hardcoded price sort in direct.go with SortResults call. Added --sort-by CLI flag. 5 new tests. Sequential on main.

### Session 33, Task 2 -- Avoid airline filter
Added FilterByAvoidAirlines to filter.go with comma-separated IATA code parsing, checking both Airline and OperatingCarrier fields. Added AvoidAirlines to search.Request and multicity.SearchParams. Wired into direct pipeline, multicity stages 3 and 4b, and --avoid-airlines CLI flag. 5 new tests. Sequential on main (shares files with Task 1).

### Session 33, Task 3 -- Connection risk tags in ranker prompt
Added connection risk tags after layover line in buildRankingPrompt: [Risky connection: Xm] for <60min layovers, [Tight connection: Xm] for 60-89min. Gives the LLM explicit signal to penalize risky connections. 3 new tests. Ran as parallel worktree agent.

### Session 33, Task 4 -- Enrich JSON output with airline codes and city names
Added AirlineCode, OriginCity, DestinationCity, OriginName, DestinationName fields to jsonLeg (omitempty). Populated from first segment data in buildJSONItineraries. 2 new tests. Ran as parallel worktree agent.

### Session 33, Task 5 -- Wire sort_by into chat conversation
Added SortBy to tripParams. Wired through parsePartialParams, mergeParams, buildRequestFromParams, system prompt, and refinement hint. 5 new tests. Sequential on main (depends on Task 1's SortBy field).

## Session 33 -- Sorting, avoid airlines, connection risk, JSON enrichment, chat sort_by

Completed all 5 planned tasks in 7 commits (5 task commits + 2 cherry-picks from worktrees) with zero reverts and zero API calls. Tasks 3 and 4 ran as parallel worktree agents while Tasks 1, 2, and 5 ran sequentially on main. Key outcomes: (1) SortResults enables sorting by price/duration/departure with CLI flag and chat support. (2) FilterByAvoidAirlines lets users exclude specific airlines by IATA code, wired into both direct and multicity pipelines. (3) Connection risk tags give the LLM ranker explicit signals about tight layovers. (4) JSON output now includes airline codes and city names for richer programmatic consumption. (5) sort_by is fully wired into the chat conversation loop. Coverage steady at ~84% across 15 packages. All build gates pass.

## Session 33 -- 10:33 -- Sorting, avoid-airlines filter, and ranker/output enrichment

Completed all 5 planned tasks with zero reverts and zero API calls. Added SortResults (price/duration/departure) to the filter pipeline, replacing the hardcoded price sort in direct.go and wiring through CLI --sort-by flag and chat sort_by param. Added FilterByAvoidAirlines for negative airline filtering across both direct and multicity pipelines. Enriched the LLM ranker with connection risk tags ([Risky connection: Xm] for <60min, [Tight connection: Xm] for 60-89min layovers) and the JSON output with airline IATA codes and city/airport names. All features were TDD with zero live API calls. Coverage steady at ~84% across 15 packages.

### Session 34, Task 1 -- Wire AvoidAirlines into chat
Added AvoidAirlines (avoid_airlines) to tripParams. Wired through parsePartialParams, mergeParams, buildRequestFromParams, system prompt, and refinement hint. 5 new tests. This was the last search.Request field not wired into chat -- now all fields are consistent between CLI and chat.

### Session 34, Task 2 -- Wire leg2_date into chat and update multicity Strategy
Added Leg2Date to both tripParams and search.Request. Wired through all chat pipeline stages. Updated multicity.Strategy to prefer req.Leg2Date over the constructor-time default, enabling chat users to specify multi-city trip dates. Also fixed AvoidAirlines not being passed through toSearchParams.

### Session 34, Task 3 -- Arrival time and max duration filters
Implemented FilterByArrivalTime and FilterByMaxDuration in filter.go. Added ArrivalAfter/ArrivalBefore/MaxDuration to search.Request. Wired through direct pipeline, multicity stages (FILTER + 4b), CLI flags (--arrival-after, --arrival-before, --max-duration), and chat tripParams (arrival_after, arrival_before, max_duration_hours). Combined tasks 82+83 into one commit since they share the same files.

### Session 34, Task 4 -- (Combined with Task 3 above)

### Session 34, Task 5 -- cmd helper edge-case coverage
Added edge-case tests for empty-segment and out-of-bounds paths in legAircraft, legLegroom, legBookingURL, legCabin, legArrival, legDeparture, legSeatsLeft, and formatPriceInsights. All pass.

## Session 34 -- Chat completeness, arrival/duration filters, helper coverage

Completed all 5 planned tasks (+ pre-task gofmt commit) in 5 commits with zero reverts and zero API calls. Key outcomes: (1) Chat conversation now has full parity with CLI -- all search.Request fields are wired including avoid_airlines and leg2_date. (2) Multicity strategy is now usable from chat via leg2_date. (3) New arrival time and max duration filters give users more control over flight selection. (4) cmd helper coverage improved with empty-segment and out-of-bounds edge-case tests. All build gates pass cleanly.

## Session 34 -- 11:32 -- CLI-chat parity, arrival/duration filters, and helper coverage

Completed all planned tasks plus a pre-task gofmt fix in 6 commits with zero reverts and zero API calls. The major milestone is full CLI-chat parity: avoid_airlines was the last search.Request field not wired into the chat conversation, and leg2_date enables multicity trip planning from chat for the first time. Added two new filter types -- FilterByArrivalTime and FilterByMaxDuration -- wired end-to-end through direct pipeline, multicity stages, CLI flags, and chat tripParams. Combined arrival and duration filter tasks into one commit since they share >80% of their files. Improved cmd helper coverage with edge-case tests for empty-segment and out-of-bounds paths. Coverage at ~85% across 15 packages. All build gates pass.

### Session 35, Task 1 -- Next-day arrival indicator
Added isNextDay helper comparing departure/arrival dates. legArrival now appends (+N) marker for cross-day flights. Added arrival_next_day boolean to JSON output. 8 new tests covering same-day, next-day, multi-day, and JSON omitempty.

### Session 35, Task 2 -- Operating carrier display (codeshare)
legAirlines now appends "(op. XX)" when OperatingCarrier differs from marketing Airline. Added operating_carrier to JSON output. 5 new tests for codeshare, same-carrier, and empty-carrier cases.

### Session 35, Task 3 -- Richer result summary in chat
Expanded resultSummaryForChat from top-1 to top-3 results, each showing airline, duration, stops, and price. Graceful degradation for 1-2 results. 4 new tests.

### Session 35, Task 4 -- Preferred airlines filter
Added FilterByPreferredAirlines (positive filter: keep only matching airline/operating carrier). Wired end-to-end through search.Request, direct pipeline, multicity (FILTER + 4b), strategy adapter, CLI flag, and chat (tripParams, parse/merge/build/prompt/hint). 9 new tests.

### Session 35, Task 5 -- Ranker LLM response caching
Added in-memory cache to Ranker keyed by SHA-256 hash of weights + itinerary data. Identical candidate sets skip the LLM call. Introduced rankerLLM interface for testability. Removed TODO comment. 3 new tests verifying cache hit, miss on different itineraries, and miss on different weights.

## Session 35 -- 12:32 -- New display features, preferred airlines filter, ranker caching

Completed all 5 planned tasks in 5 commits with zero reverts and zero API calls. Used parallel worktree agents for Tasks 3 (chat summary) and 5 (ranker caching) while working on Tasks 1+2+4 sequentially on main. Key outcomes: (1) Users see next-day arrival markers (+N) and codeshare info (op. XX) in table/JSON output. (2) Chat LLM gets richer context with top-3 result summaries instead of just the cheapest. (3) New preferred_airlines positive filter complements existing avoid_airlines for full airline control. (4) Ranker caches LLM responses keyed by SHA-256 of weights+itineraries, saving tokens and latency on repeated rankings. 29 new tests total. All build gates pass.

### Session 36, Task 1 -- Multi-leg per-leg cabin class columns
Fixed multi-leg table to show "Leg 1 Cabin" / "Leg 2 Cabin" instead of single "Cabin" showing only leg 0. Same bug pattern as the CO2 per-leg fix from Day 28. 1 new test.

### Session 36, Task 2 -- Flight number in JSON output
Added flight_number (omitempty) to jsonLeg, populated from first segment's FlightNumber. Original plan was bags display but BagsIncluded is only populated by inactive Kiwi provider -- SerpAPI doesn't return bag data. 2 new tests.

### Session 36, Task 3 -- Fallback global hub stopovers
StopoversForRoute now returns filtered global hubs (IST, SIN, HKG, NRT, LHR, CDG, ICN, BKK) when no route-specific stopovers exist. Origin/destination airports excluded from fallback list. This enables multicity search for any route, not just the 3 hardcoded corridors. Updated search_test.go to reflect new behavior. 5 new tests.

### Session 36, Task 4 -- Combiner red-eye leg filtering
CombineLegs now rejects combinations where leg2 departs between 00:00-04:59. Reuses existing isRedEye helper from ranker.go (same package). 5 new subtests.

### Session 36, Task 5 -- Multi-leg trip summary footer
priceSummary now appends total trip duration range for multi-leg itineraries. Added formatTripDuration helper for compact "Xd Yh" display. Single-leg results unchanged. 3 new tests.

## Session 36 -- Multi-leg display fixes, fallback stopovers, combiner quality

Completed all 5 planned tasks in 5 commits with zero reverts and zero API calls. Used parallel worktree agents for Tasks 3 (stopovers) and 4 (combiner) while working on Tasks 1, 2, 5 sequentially on main. Task 91 was pivoted from bags display to flight number JSON (BagsIncluded only populated by inactive Kiwi provider). Key outcomes: (1) Multi-leg table now shows per-leg cabin class columns. (2) JSON output includes flight numbers for programmatic consumers. (3) Multicity search now works for any route via 8 global fallback hubs. (4) Combiner hard-filters red-eye leg2 departures. (5) Multi-leg price summary includes trip duration range. 16 new tests total. All build gates pass.

---

### Session 37, Task 1 -- Refactor stage 4b multi-city filtering
Extracted passesAllFilters helper to replace ~70 lines of verbose single-element slice wrapping in stage 4b. Net -56 lines from multicity.go. 14 table-driven tests for the helper.

### Session 37, Task 2 -- Add total_trip to JSON output
Added total_trip field to jsonItinerary, populated from itin.TotalTrip using formatTripDuration. Omitted when zero (single-leg). 2 new tests.

### Session 37, Task 3 -- Wire departure time CLI flags
Added --departure-after and --departure-before flags to searchCmd. These fields already existed on search.Request and worked in chat but had no CLI flags. No new tests needed -- existing filter tests cover the logic.

### Session 37, Task 4 -- Itinerary deduplication in multicity
Added deduplicateItineraries after price sort in the multicity pipeline. Keys by flight number + departure per leg, keeps cheapest. Re-sorts after dedup since map iteration is unordered. 3 new tests.

### Session 37, Task 5 -- Stopover duration control via CLI and chat
Users can now customize city stopover duration via --min-stopover/--max-stopover CLI flags and min_stopover_hours/max_stopover_hours in chat. Threaded through search.Request -> SearchParams -> CombineParams. Zero values use defaults. 7 new tests across combiner and chat.

## Session 37 -- Refactoring, deduplication, and user control

Completed all 5 planned tasks in 4 commits with zero reverts and zero API calls. Used parallel worktree for Task 1 (refactor) while working on Tasks 2-3 on main. Key outcomes: (1) Stage 4b filter code reduced from ~70 lines to ~15 via passesAllFilters helper. (2) JSON output includes total_trip duration for multi-leg itineraries. (3) CLI now has departure time and stopover duration flags (completing chat-CLI parity). (4) Multicity deduplicates identical-flight itineraries, keeping cheapest. (5) Users can control stopover city visit length. 26 new tests total. All build gates pass.

### Session 38, Task 1 -- Remove KiwiID from StopoverCity and SearchParams
Removed KiwiID field from StopoverCity struct, all KiwiID assignments in 3 route-specific + 8 global fallback entries, OriginKiwiID/DestinationKiwiID from SearchParams, and KiwiID refs in fetch goroutines and diagnostic_test.go. Left types.SearchRequest untouched (Kiwi provider reads it). No behavioral change for SerpAPI pipeline.

### Session 38, Task 2 -- Simplify fetchWithDualSort to single fetch
Replaced fetchWithDualSort (which made two identical API calls since SerpAPI ignores SortBy) with a direct fetchFromAllProviders call. Removed deduplicateFlights helper and all associated tests. Kiwi sort constants kept in config/routes.go since the inactive Kiwi provider still references them. Net removal: ~90 lines of dead code + tests.

### Session 38, Task 3 -- Add India-US route stopovers (DEL/BOM to JFK)
Added DELToJFKStopovers (8 cities) and BOMToJFKStopovers (7 cities) with curated Asia-Pacific and European hubs. Registered in stopoversMap. 2 new tests verify route-specific (not fallback) results and check expected airports. Implemented in parallel worktree, merged to main with KiwiID stripped (compatible with task 1).

### Session 38, Task 5 -- No-results filter suggestion in chat
Added filterSuggestion(tripParams) that checks 8 optional filter categories and suggests relaxing them when search returns zero results. Wired into chatLoop after the no-results message. 10 new tests (9 table-driven for active filters + 1 no-filters case). Implemented in parallel worktree.

### Session 38, Task 4 -- Consolidate stage 3 filter logging
Extracted applyBoth closure in stage 3 filter loop to eliminate repetitive before/after counting. Each filter application is now one line (plus a closure wrapper for parameterized filters). MaxStops kept separate since it uses different per-leg parameters. Same log output format preserved. Net reduction: ~15 lines while maintaining identical behavior.

## Session 38 -- Kiwi dead code cleanup, route expansion, chat UX

Completed all 5 planned tasks in 5 commits with zero reverts and zero API calls. Used parallel worktrees for tasks 3 (stopovers) and 5 (chat filter hints) while working sequentially on tasks 1-2 on main, then task 4.

Key outcomes:
1. Removed KiwiID from StopoverCity struct and OriginKiwiID/DestinationKiwiID from SearchParams (dead Kiwi-era code).
2. Replaced fetchWithDualSort (two identical SerpAPI calls) with single fetchFromAllProviders. Removed deduplicateFlights helper and 6 associated tests.
3. Added India-US route stopovers: DELToJFK (8 cities) and BOMToJFK (7 cities) with curated hubs.
4. Consolidated stage 3 filter logging with applyBoth closure.
5. Chat now suggests relaxing filters when search returns zero results.

Net lines: ~-100 Go code removed via cleanup, ~+200 added for new features. 12 new tests.

### Session 40, Task 1 -- Consolidate time-of-day filter functions
Extracted shared filterByTimeOfDay helper from FilterByDepartureTime and FilterByArrivalTime. Both functions shared identical parse/validate/iterate logic, differing only in which segment time they extracted. Net reduction: ~25 lines. All existing time filter tests pass unchanged.

### Session 40, Task 2 -- Single-flight filter predicates for passesAllFilters
Added 7 exported single-flight predicate functions (FlightPassesBlocked, FlightPassesAlliance, FlightPassesDepartureTime, FlightPassesArrivalTime, FlightPassesMaxDuration, FlightPassesAvoidAirlines, FlightPassesPreferredAirlines). Rewrote passesAllFilters to use early-return predicates instead of wrapping each flight in []Flight{f}. Also extracted parseAirlineCodes helper from FilterByAvoidAirlines/FilterByPreferredAirlines. 7 new predicate tests.

### Session 40, Task 3 -- Add India-UK route stopovers (DEL/BOM to LHR)
Added DELToLHRStopovers (6 cities) and BOMToLHRStopovers (6 cities): BKK, SIN, KUL, HKG, CMB, IST. All avoid Middle East blocked airspace. Total route-specific corridors: 7 (was 5). 2 new tests. Implemented in parallel worktree.

### Session 40, Task 4 -- Ranker cache hit/miss counters
Added hits/misses int fields to Ranker, CacheStats() method returning (hits, misses int). Incremented on each Rank() call. 3 new tests. Implemented in parallel worktree.

### Session 40, Task 5 -- Chat system prompt agent personality
Enhanced chatSystemPrompt to position booker as a proactive travel planning agent per VISION.md. Added guidance for suggesting stopovers, recommending nearby airports, explaining tradeoffs, and asking about flexibility. JSON extraction format unchanged. All 11 system prompt tests pass.

## Session 40 -- Filter refactoring, route expansion, observability, agent personality

Completed all 5 planned tasks in 4 commits with zero reverts and zero API calls. Used parallel worktrees for tasks 3 (stopovers) and 4 (ranker) while working sequentially on tasks 1-2 on main, then task 5.

Key outcomes:
1. Consolidated duplicate time-of-day filter logic into shared filterByTimeOfDay helper.
2. Added 7 single-flight predicates, eliminating slice-wrapping antipattern in passesAllFilters.
3. Added India-UK route stopovers: DEL/BOM->LHR (6 cities each) with curated hubs.
4. Ranker now tracks cache hit/miss stats for observability.
5. Chat system prompt now positions booker as a proactive travel agent per VISION.md.

12 new tests total. All build gates pass.

### Session 41, Task 1 -- Bidirectional route lookup in StopoversForRoute
Added reverse direction lookup to StopoversForRoute: when origin->dest not found, checks dest->origin and filters origin/dest airports from results. This doubles effective route coverage (7 routes -> 14). 2 new tests verify route-specific data (not fallback) via airports unique to each route (KUL, FRA, CMB).

### Session 41, Task 2 -- Add India-US West Coast stopovers (DEL/BOM to SFO)
Added DELToSFOStopovers (6 cities: NRT, ICN, HKG, BKK, SIN, IST) and BOMToSFOStopovers (5 cities: NRT, HKG, BKK, SIN, IST). Primary corridor via East Asia Pacific routing. Total route-specific corridors: 9 (was 7). 2 new tests.
