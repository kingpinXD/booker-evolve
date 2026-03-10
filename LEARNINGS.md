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
