# LEARNINGS

> **APPEND-ONLY** — Only add new entries at the bottom. Never edit or remove existing entries.

Format: `## Lesson: [insight]` followed by a context paragraph explaining when and why this applies.

---

## Lesson: Run self-assessment before planning any work

The self-assessment revealed that 11 of 14 packages have zero tests and total coverage is 6.0%. Without this data, the session plan would have been guesswork. Always start an evolution session by measuring the current state -- build status, test count, coverage by package, lint issues -- so that tasks can be prioritized by actual gaps rather than assumptions.

## Lesson: Pure function tests are high-value, low-cost

search/filter.go and types/types.go reached 100% coverage with straightforward table-driven tests. No mocking, no API calls, no complex setup. When choosing what to test first, prioritize packages with pure functions -- they give the most coverage improvement for the least effort.

## Lesson: TDD red-green cycle catches interface mismatches early

Writing strategy_test.go before strategy.go caught that `toSearchParams` needed to be a method visible to the test (same package). The compile failure in the red phase confirmed the test was actually exercising the right surface area before any implementation existed.

## Lesson: Gate integration tests early to keep the default test suite fast and reliable

Adding `//go:build integration` to API-dependent tests on Day 1 unblocked all subsequent work. Before gating, `go test ./...` failed without API keys, which would have broken the build gate for every future commit. Do this first in any project that mixes unit and integration tests.

## Lesson: Always re-check CLAUDE.md directives before starting planned tasks

Task 6 (kiwi parser tests) was planned on Day 3 before the CLAUDE.md directive to ignore Kiwi was added. On Day 4, the self-assessment caught this and the task was correctly skipped. Review governance file directives at session start -- plans from earlier sessions may conflict with updated rules.

## Lesson: Check the full dependency graph before starting deferred work

Tasks 8 and 9 were deferred from Day 4. When resuming on Day 7, the self-assessment revealed that issue #3 (LLM Picker) was a missing dependency for Task 9 (CLI wiring) that had never been added to TODO.md. Scanning all open GitHub issues at session start catches these gaps before they become blockers mid-session.

## Lesson: Extract interfaces at consumption sites to make concrete types testable

Picker depended on *llm.Client directly. Defining ChatCompleter (a one-method interface) in picker.go -- right where it is consumed -- let tests inject a mockLLM without changing llm/client.go at all. In Go, define interfaces where they are used, not where they are implemented. This pattern also worked for currency.Converter: wrapping package-level globals behind a struct made the pure conversion logic testable without HTTP calls.

## Lesson: Parallel worktree agents need duplicate-aware merging

When running 3 test-writing agents in parallel worktrees for the same package (multicity), one agent duplicated test functions already present in ranker_test.go (TestFormatDuration, TestBuildSystemPrompt, contains helper). The fix was trivial -- remove duplicates after merge -- but the issue is predictable. When parallelizing test work within a single Go package, either split by file explicitly or plan for dedup during merge.

## Lesson: Worktree agents may write to the main tree instead of the worktree

When using isolation: "worktree", the agent may still write files to the main working directory instead of the worktree. After agent completion, always check `git status` in the main tree for uncommitted changes and copy any missing files from the worktree path before committing. Do not assume agents committed their work in the worktree branch.

## Lesson: Avoid threading per-search data through per-result interfaces

PriceInsights from SerpAPI is per-search (one value for the whole response), not per-flight. Attempting to thread it through provider.Provider, cache.Provider, and search.Strategy interfaces -- which all return per-flight/per-itinerary slices -- leads to over-engineering. The pragmatic approach: store it on the concrete provider with a getter and access it directly from the CLI. Keep per-search metadata separate from per-result data.

## Lesson: go-pretty table writer uppercases all header text

When capturing table output in tests via os.Pipe, remember that go-pretty's table.Writer renders all header strings in uppercase (e.g. "Score" becomes "SCORE"). Test assertions for header presence must match the uppercased form.

## Lesson: Separate testable logic from CLI wiring in chat commands

The chat command needs both a conversation loop (LLM calls, input/output) and a search pipeline (Picker, strategies, providers). By extracting chatLoop as a function that takes io.Reader/io.Writer and search.ChatCompleter, the entire conversation-to-search flow is testable with mocks. The cobra RunE function only wires up real dependencies. This separation also reuses ChatCompleter -- the same interface defined for Picker tests works for chat tests too.

## Lesson: Combine closely related tasks to reduce commit overhead

Tasks 9 (chat command scaffold) and 10 (wire chat to search) were planned as separate tasks, but the chat command's value comes from the full loop -- conversation + search execution. Implementing them as one commit kept the code coherent and avoided an intermediate state where the chat command exists but cannot actually search. Plan for logical units, not arbitrary splits.

## Lesson: FlexDays filtering is useless without multi-date search

The original direct strategy searched one date and filtered results by FlexDays. But SerpAPI returns flights only for the requested date, so the filter never expanded the result set. The fix: when FlexDays > 0, loop over each date in the range and make separate provider calls. This turns FlexDays from cosmetic into functional -- it actually finds cheaper flights on nearby dates. Always verify that filter parameters have upstream data to work with.

## Lesson: Worktree agents need gofmt verification on merge

Worktree agents produce correct, passing code, but may have minor gofmt violations (tab alignment, spacing). Always run `gofmt -l .` after rebasing worktree branches into main. This is cheap to fix but easy to miss if you only run `go test` and `go vet` without the lint pass.

## Lesson: Refactoring shared infrastructure unlocks simpler feature additions

Extracting buildPicker into cmd/infra.go reduced both runSearch and runChat by ~20 lines each, but the bigger win is that adding new strategies (like nearby-airport) to the picker now requires editing one function instead of two. When duplicate infrastructure accumulates across commands, consolidate early -- it compounds.

## Lesson: Test airport cluster membership bidirectionally

When writing tests for airport cluster data, test both directions: (1) a known code returns the expected siblings, and (2) every code in every cluster maps back to the correct number of siblings. The consistency test (TestNearbyAirports_AllClusters) caught a potential off-by-one before it shipped. Consistency checks on reference data are cheap and high-value.

## Lesson: Verify SerpAPI field availability before planning tasks

SerpAPI provides many fields (carbon_emissions, legroom, etc.) that aren't in our response struct yet. Before planning a "display X" task, check both (1) the SerpAPI docs for field existence and (2) the response struct + parser to see if parsing is needed. Tasks that only need display are quick (cabin class was already parsed). Tasks that need end-to-end parsing (carbon emissions: response.go + parser.go + types.go + display) take 3x longer but are still straightforward.

## Lesson: Watch for integer division truncation in unit conversions

When converting between units (grams to kg, cents to dollars, etc.) using integer types, division truncates toward zero. CarbonKg = grams/1000 gave 0 for 800 grams. Use rounding: (grams+500)/1000 yields the nearest integer. Always add a test for values just below the divisor boundary (e.g., 800g, 499g) to catch this early.

## Lesson: Multi-leg table columns must handle each leg independently

When adding per-leg data columns (CO2, arrival, cabin) to the multi-leg table layout, each leg needs its own column -- do not reuse a single column with hardcoded leg index 0. The single-leg layout uses one column; the multi-leg layout must iterate. The CO2 bug (only leg 0 shown) was a 3-line fix but invisible until someone checked leg 1 data. When adding any new per-leg column, always verify both layouts.

## Lesson: Parallel worktree agents work well when files are strictly disjoint

Session 28 ran two worktree agents in parallel: one for airports.go (pure data addition) and one for filter.go + chat.go + strategy.go (new filter feature). Both completed independently and rebased cleanly with zero conflicts. The key was ensuring no shared files between the two tasks. When tasks touch different files in the same package, parallelization is safe -- the risk is only when they modify the same file.

## Lesson: Multicity filters need separate per-flight and per-itinerary logic

MaxPrice in direct search filters individual flights (FilterByMaxPrice on each flight's price). In multicity, the meaningful filter is on total itinerary price (sum of both legs), not individual leg price. A $300 leg 1 + $500 leg 2 = $800 itinerary should be rejected by MaxPrice=700, even though each leg individually is below $700. Always consider the aggregation level when porting filters from direct to multicity.

## Lesson: Three-way parallel worktrees with sequential sub-tasks

When two tasks share files (65+66 both touch multicity/), put them in one worktree running sequentially. Other independent tasks get their own worktrees. This session used 3 worktrees: one for 65+66 (sequential), one for 67, one for 68. All rebased cleanly. The gofmt alignment issue after rebase is expected when struct fields are added in one branch and other fields already existed with different alignment.

## Lesson: Defined-but-unused constants reveal missing wiring

Session 31 found config.SerpAPIParamStops defined but never referenced in any provider call. Grepping for unused constants during self-assessment is a low-effort way to find features that were half-implemented. Similarly, SerpAPIParamReturnDate was never sent despite the round-trip type being set -- the constant existed, the request type was correct, but the parameter was omitted. When reviewing a codebase, scan for constants and struct fields that are defined but never read; they often indicate incomplete work that is easy to finish.

## Lesson: Batch closely-coupled filter tasks into a single commit

Tasks 82 (arrival time filter) and 83 (max duration filter) touched the exact same set of files (filter.go, strategy.go, direct.go, multicity.go, search.go, chat.go, chat_test.go). Implementing them separately would have required two nearly-identical wiring passes through 9 files. Implementing them together halved the wiring effort and produced a cleaner, single commit. When two tasks have >80% file overlap and follow the same pattern (add filter, add Request field, wire through pipelines), batch them.

## Lesson: Field renames in structs break test files referencing internal fields

Renaming Strategy.leg2Date to Strategy.defaultLeg2Date broke strategy_test.go which accessed the unexported field directly (s.leg2Date, &Strategy{leg2Date:}). When renaming struct fields, always grep the test files for direct references to the old name. Use replace_all to fix all occurrences in one pass rather than hunting them down individually.

## Lesson: Parallel worktree agents work well for disjoint package work

Session 35 ran Tasks 87 (cmd/chat.go only) and 89 (search/multicity/ranker.go only) as parallel worktree agents while main worked on Tasks 85+86+88 sequentially. Both parallel agents completed successfully with no merge conflicts. The key constraint is strict file disjointness -- the chat task touched chat.go/chat_test.go and the ranker task touched ranker.go/ranker_test.go, with zero overlap. Rebasing both branches into main after the sequential tasks completed cleanly. This pattern saves significant wall-clock time on 5-task sessions.

## Lesson: Verify parsed data is actually populated before adding display features

Session 36 planned to add baggage display (BagsIncluded) to the table and JSON output. During implementation, discovered that BagsIncluded is only populated by the inactive Kiwi provider -- SerpAPI doesn't return bag data at all. The field would always be zero for all real users. Before planning a "display parsed data" task, grep for where the data is actually set (not just defined) and verify the active provider populates it. This saved wasted effort by pivoting to flight number display instead.

## Lesson: Worktree agents may not commit -- copy uncommitted changes via file copy

Session 36 worktree agents completed their tasks but did not create commits on their branches. The worktree branches showed the same commit as the base. The changes existed as uncommitted modifications in the worktree directories. To integrate: cp the modified files from the worktree dir to the main repo, verify tests pass, then commit on main. This is simpler than trying to commit in the worktree and cherry-pick.

## Lesson: Cherry-picking worktree commits that touched governance files causes conflicts

Session 37 worktree agent committed on its branch. When cherry-picking into main, governance files (JOURNAL.md, LEARNINGS.md, TODO.md) conflicted because they were also modified on main. The Go code itself was clean. Better approach: copy only the Go files from the worktree and commit them on main, ignoring the worktree's governance file changes. This avoids conflict resolution entirely.

## Lesson: Dedup itineraries after sort requires re-sort

deduplicateItineraries uses a map to track best (cheapest) per key. After extracting results from the map, iteration order is random. Must re-sort by price after dedup. This is a common pattern when using maps for deduplication -- always re-sort if order matters.

## Lesson: Kiwi constants must stay in config even after removing Kiwi-specific code from the pipeline
The Kiwi provider (provider/kiwi/kiwi.go) directly references config.KiwiSortByQuality, config.KiwiParamSortBy, etc. Per CLAUDE.md, we don't modify Kiwi code. So even when removing Kiwi-era patterns from the active pipeline (like fetchWithDualSort), the config constants must remain. Always check cross-package references before removing constants.

## Lesson: Worktree merges with struct changes need manual KiwiID stripping
When a worktree branch adds stopover entries that include KiwiID fields, but main has already removed KiwiID from the struct, cherry-picking causes compile errors. The safest approach is to manually apply the worktree's Go changes to main, stripping the removed fields, rather than trying git rebase/cherry-pick.

## Lesson: Existing internal predicates make exported wrappers trivial
The search/filter.go package already had unexported single-flight predicates (isFlightBlocked, flightMatchesAlliance, flightMatchesAvoid, flightMatchesPreferred). Adding exported wrappers (FlightPasses*) is a thin layer over these -- the real refactoring value is in the consumer site (passesAllFilters) where slice-wrapping is eliminated. Check for existing internal helpers before designing new predicate APIs.

## Lesson: Adding optional struct fields with zero-value backward compat
When extending a struct used across multiple presets (e.g., RankingWeights), adding a new field with a zero value preserves all existing behavior. The system prompt conditionally includes the new criterion only when the field is non-zero, so existing profiles (budget/comfort/balanced) produce identical output. This pattern avoids migrating every call site.

## Lesson: Large file splits succeed cleanly when no logic changes
Extracting 25 functions and 4 types from cmd/search.go (768 lines) into cmd/display.go required zero test changes -- all existing tests continued passing because the functions stayed in the same package with identical signatures. The key is moving only display/formatting code with no behavioral changes. This kind of pure-mechanical refactor is safe to run in a parallel worktree alongside feature work in disjoint files.

## Lesson: Parsing and merging a config field does not mean it takes effect
Session 44 found that chat profile switching was dead code: tripParams.Profile was parsed from LLM JSON, merged between turns, and even displayed in refinement hints -- but the ranker's weights were fixed at startup and never updated. The fix was small (SetWeights method + interface), but the bug was invisible because all the wiring looked correct at each stage. When adding a user-facing setting, trace the full path from parse -> merge -> build request -> execute. If any link only stores the value without acting on it, the feature is broken despite appearing complete.

## Lesson: Nil optional dependencies silently degrade features
Session 46 found that CompositeStrategy accepted a nil ranker without error -- it simply skipped ranking, returning unscored results from "both" mode. The constructor signature `NewCompositeStrategy(ranker, strategies...)` accepted nil as a valid *Ranker, and the Rank method had a nil guard that silently returned. This meant the "integration quality" feature (merged results re-ranked as a unified set) was dead code since introduction. When a dependency is optional, audit every call site to ensure the non-nil case is actually wired. Grep for constructors that pass nil for interface/pointer params -- they may indicate incomplete feature plumbing rather than intentional optionality.

## Lesson: Expanding a function's return values is a safe, high-leverage change
Session 47 changed Picker.Pick from (Strategy, error) to (Strategy, string, error) to surface strategy reasoning. Go's compiler catches every call site that needs updating -- there is no risk of silent breakage. The pattern is: (1) change the signature, (2) compile to find all callers, (3) update each caller to use or discard the new value. This is safer than adding a field to a struct (which compiles silently even if no consumer reads it). When a function has useful internal information that callers should see, adding a return value is preferable to side channels (logging, struct fields) because the compiler enforces that all callers handle it.

## Lesson: Keyword-based inference is cheap and reliable as LLM fallback
When relying on LLM output for structured fields like ranking profile, a deterministic keyword scanner over conversation history provides a reliable fallback. The LLM may not always emit optional fields, but user intent is often clear from keywords ("cheapest", "comfortable", "green"). The scanner runs in O(messages * keywords) which is negligible, and the counter-based approach handles ambiguity gracefully by preferring the most-mentioned category.

## Lesson: Layover data enriches LLM context for tradeoff discussion
Changing chat summary from "1 stop" to "1 stop (3h IST)" gives the LLM concrete data points to explain tradeoffs ("the 3-hour Istanbul layover saves $200"). This required no new API calls -- segment LayoverDuration and Destination were already parsed from SerpAPI. The formatLayoverSummary function gracefully degrades to stop count when data is missing.

## Lesson: Fallback paths should inspect the request, not return a hardcoded default
The Picker fallback returned "direct" for every request, ignoring Leg2Date entirely. This meant LLM-less environments always used the wrong strategy for multi-city trips. Fallback/default paths are often written once during initial implementation and never revisited as the request shape gains new fields. When adding a field that changes which code path should execute (like Leg2Date selecting multicity vs direct), audit both the primary path and all fallback/error paths to ensure they respect the new field. A fallback that ignores request context is a silent correctness bug.

## Lesson: Go worktree test leakage with go test ./...
When using git worktrees inside the repo directory (.worktrees/), `go test ./...` picks up test files from all worktrees since Go follows the filesystem. Use `go test $(go list ./... | grep -v worktree)` to exclude them, or `golangci-lint run --skip-dirs '.worktrees'` for lint. This is critical when parallel worktree agents write tests that reference code not yet on the main branch.

## Lesson: Adding slice fields to structs breaks == comparison
Adding a []string field (like ClearFields) to a struct that was previously compared with == will cause compile errors. Use reflect.DeepEqual for struct comparison in tests when the struct has slice or map fields.

## Lesson: Intercept deterministic queries before the LLM to save latency and cost
When a user request can be answered entirely from local state (e.g. "compare options 1 and 3" when results are cached), detect the intent via simple keyword matching and short-circuit the LLM call. This is faster, cheaper, and more reliable than asking the LLM to reconstruct data it does not have. The pattern is: (1) cache results after each search, (2) define keyword sets for each interceptable intent, (3) parse parameters from the user message, (4) format a response directly from cached data. Reserve the LLM for open-ended reasoning where deterministic logic cannot suffice.

## Lesson: Zero-value fields create a "sticky parameter" trap in merge-based state
When merging partial updates into accumulated state (e.g. tripParams across chat turns), zero values are indistinguishable from "not set" in Go. This makes it impossible for users to reset a field to its zero value -- the merge always keeps the previous non-zero value. The fix is an explicit clear list: the user sends `clear_fields: ["direct_only", "max_price"]` and the merge function zeroes those fields on the previous state before applying the partial update. This pattern applies to any accumulator where zero is a valid user intent distinct from "no change."

## Lesson: Show the LLM what you show the user -- thread display data into conversation history

When a chat system displays computed insights to the user (e.g. price levels, benchmarks, suggestions) but does not add them to the LLM's conversation history, the LLM cannot reference that data in subsequent responses. Session 52 found that priceInsightHint was printed to the user but absent from the message history, so the LLM could not say "these prices are below typical for this route." The fix is simple: append the same data to the result summary that goes into history. The general rule is that any computed context shown to the user should also be visible to the LLM, otherwise the conversation becomes incoherent -- the user sees information the LLM does not know about.

## Lesson: Reflection eliminates per-field maintenance burden in struct merge functions
When a function like mergeParams has N if-blocks for N struct fields (one for clearing, one for zero-value merge), adding a new field requires modifying two code locations. Replacing the if-blocks with a reflection loop over struct fields using json tags reduces the function from ~113 lines to ~30 lines and automatically handles new fields. The key insight is that reflect.Value.IsZero() correctly identifies the zero value for all Go basic types (string "", int 0, bool false, float64 0.0, nil slice). Fields that need special treatment (e.g. ClearFields which is ephemeral and should not merge) are skipped by json tag name. This trades compile-time field safety for maintenance simplicity -- a reasonable tradeoff when all tests continue to pass and the struct has 20+ fields.

## Lesson: File splits are low-risk refactors that pay off immediately
When a file crosses ~800 lines with a mix of core logic and helpers, splitting into two files (core + helpers) reduces cognitive load for future changes. The key is a clean separation: keep the main loop and user-facing interaction code in the original file, move pure functions and utility helpers to a new file. Since Go packages are the compilation unit (not files), this is purely organizational with zero behavior change. All tests remain in the original test file and pass unchanged. The hardest part is getting imports right -- both files may need different subsets of the original import list.

## Lesson: Thread assistant-generated tips into LLM history, not just stdout
When a chat system generates tips (like stopover suggestions) and displays them to the user but does not append them to the conversation history, the LLM cannot follow up on them. The user sees "Flying via Bangkok saves money" and asks "how do I do this", but the LLM has no context about Bangkok. The fix: append the tip as an assistant message in history. This also means the LLM's "how do" prefixes in looksLikeHelp were masking the problem -- removing the broad prefix match lets contextual questions reach the LLM, which now has the context to answer them.

## Lesson: Per-operation timeouts prevent cascading hangs in interactive loops
A chat loop with a single session-level timeout (e.g. 5 minutes) lets individual operations hang for the entire budget. Adding a per-search timeout (2 minutes) with context.WithTimeout ensures one slow multicity search doesn't freeze the entire session. The key is wrapping context errors (DeadlineExceeded, Canceled) with user-friendly messages at the call site, so the user sees "search timed out" instead of raw Go errors.
