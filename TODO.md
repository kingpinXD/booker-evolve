# TODO

## Task 165: Fix looksLikeHelp false positive
**Status:** done
**Plan:** Remove "how do"/"how does" prefixes from looksLikeHelp. Keep only exact/near-exact triggers.
- [x] Write test: "how do I do this" should NOT trigger help
- [x] Write test: "how do I set leg2_date" should NOT trigger help
- [x] Update looksLikeHelp to remove "how do"/"how does" prefixes
- [x] Verify existing help tests still pass
- [x] Run go build && go test ./cmd/... && go vet ./...

## Task 166: Conversational stopover suggestion with LLM history
**Status:** done
**Plan:** Make stopoverSuggestion conversational, add to LLM history, update system prompt for multi-city guidance.
- [x] Write test: stopoverSuggestion should not contain "leg2_date"
- [x] Write test: LLM history should contain stopover tip after results
- [x] Update stopoverSuggestion message to be conversational
- [x] Add stopover tip to LLM history in chatLoop
- [x] Update chatSystemPrompt with multi-city conversational guidance
- [x] Run go build && go test ./cmd/... && go vet ./...

## Task 167: Per-search timeout in chatSearch
**Status:** done
**Plan:** Add 2-minute per-search context timeout in chatSearch. Return friendly message on timeout.
- [x] Write test: chatSearch with cancelled context returns friendly error
- [x] Add context.WithTimeout in chatSearch
- [x] Wrap context.DeadlineExceeded with user-friendly message
- [x] Run go build && go test ./cmd/... && go vet ./...

## Task 168: Search progress feedback for multicity
**Status:** done
**Plan:** Show stopover cities being searched when multicity strategy is used.
- [x] Write test: multicity search output contains stopover city names
- [x] Detect multicity strategy in chatSearch by name
- [x] Call StopoversForRoute and display city list
- [x] Run go build && go test ./cmd/... && go vet ./...

## Task 169: Tests for new chat behavior
**Status:** done
**Plan:** Integration-style chatLoop tests for the new conversational flow.
- [x] Test: "how do I do this" after stopover tip goes to LLM
- [x] Test: LLM history contains stopover suggestion
- [x] Test: per-search timeout produces friendly error in chatLoop
- [x] Run full go build && go test ./... && go vet ./...
