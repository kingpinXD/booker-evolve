## Related Repositories

- **Anuma AI Portal** (backend): `/Users/tanmay/IdeaProjects/jedi2002/ai-portal`
  - Go backend for the Anuma LLM router service
  - OpenAI-compatible API at `POST /api/v1/chat/completions`
  - Auth via `X-API-Key: anuma_live_<key_id>.<secret>` header
  - Supports models: `auto`, `gpt-4o-mini`, `claude-3-5-sonnet`, `gpt-4o`, `o1`
  - Routes through Bifrost multi-provider gateway (OpenAI, Anthropic, Google, Mistral)

## Code Standards

### General Principles
- Follow Go best practices and idiomatic patterns
- Package comments required at top of files
- Follow TEST DRIVEN DEVELOPMENT when adding new features clarify the expected outcome always if you are not clear.
- Always use TODO's when working on adding new code 

### Coding Standards
- Write concise, readable, and maintainable code
- Naming: Variable/function names to clarify intent
- Functions should not be doing too many things
- All exported types/functions must have comments
- Prefer switch cases over nested if else statements
- Prefer early returns


### KISS: Keep It Simple Stupid
- Do not add unnecessary abstractions or indirection — could this be simpler?
- Do not add over-engineered solutions for simple problems
- Do not add clever one-liners that sacrifice readability
- Do not add premature optimization or generalization
- Do not add dead code, unused variables, redundant checks

### Modifying Existing Code
- **Always modify existing functions** instead of creating new duplicate functions with slightly different signatures
- When adding a new parameter to a function, update the existing function signature rather than creating a `FunctionWithNewParam` variant


**Important** when adding any code run an analysis to see if the number of lines of code added can be reduced in any way,
whether it is extracting common logic into functions, simplifying logic
**Important** prioritize simple readable code, overcomplicated ones.
It is okay to not handle cases which are unrealistic in favor of keeping the logic simple

## Evolution System

This project includes a self-evolving agent system. The agent (booker-agent) runs via scheduled Claude Code CLI sessions to autonomously improve the codebase.

### Governance Files
- `IDENTITY.md` — Agent constitution and rules (IMMUTABLE)
- `PERSONALITY.md` — Agent voice and style (IMMUTABLE)
- `JOURNAL.md` — Append-only session log
- `LEARNINGS.md` — Append-only self-reflection insights
- `SESSION_NUMBER` — Session counter

### Protected Files (never modify during evolution)
- `IDENTITY.md`, `PERSONALITY.md`
- `scripts/*`
- `.github/workflows/*`
- `skills/*`
- `CLAUDE.md`

### Flight Providers
- **SerpAPI is the only active provider.** Use it for all flight search work.
- **Ignore the Kiwi provider entirely** (`provider/kiwi/`). Do not fix, test, improve, or integrate Kiwi code. It is inactive and will be removed later.
- If a task involves Kiwi, skip it.

### API Usage
- **SerpAPI** has a limited free tier. Minimize API calls during evolution sessions.
- **Always use the caching layer** (`provider/cache/`) — check cache before making live requests.
- **Session limit: 6 SerpAPI calls max.** If a task would exceed this, skip it and log the reason.
- When developing new features that touch flight search, write tests against cached/mocked data first. Only make live API calls if absolutely necessary.
- Prefer working with existing cached data in `.cache/flights/` during development.

### Development Direction
- **Prefer new features to retroactive coverage.** Adding tests for existing stable code is a low priority. Focus on building new capabilities (more routes, better ranking, flexible dates, multi-city improvements) and writing tests for that new code.
- **Refactoring is welcome.** The codebase can be reorganized to be more idiomatic Go — better package boundaries, cleaner interfaces, simpler function signatures. If something feels awkward or overengineered, improve it.
- Coverage for existing code is fine as a fallback when there are no features or refactors to work on.

### GitHub Issues
- **Completely ignore closed issues.** Do not read, comment on, or reference closed issues. Only work on open issues.
- Do not add comments to closed issues under any circumstances.

### Safety Rules
- Every code change must pass: `go build && go test ./... && go vet ./...`
- Maximum 3 fix attempts per failure, then auto-revert
- Budget caps per phase prevent runaway API spending
- Issue content is sanitized before prompt injection

