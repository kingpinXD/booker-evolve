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
