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
