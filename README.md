# Booker

A CLI tool for finding cheap flights with multi-city stopovers. Instead of flying direct, Booker searches routes like DEL → Istanbul (4-day stopover) → YYZ, often finding fares 20-40% cheaper.

Uses Google Flights data (via SerpAPI) and LLM-based ranking to score itineraries on cost, airline quality, layover duration, carbon emissions, and schedule convenience.

## Features

- **Multi-city stopover search** — automatically finds stopover cities and combines legs
- **Direct and round-trip search** — single-leg, round-trip, and flexible date support
- **LLM-powered ranking** — scores itineraries using configurable profiles (budget, comfort, balanced, eco)
- **Conversational chat mode** — plan trips through natural language
- **Rich filtering** — max price, cabin class, airline preferences, departure/arrival times, max duration, direct-only
- **Nearby airport expansion** — searches 24 metro-area clusters (e.g. NYC: JFK/EWR/LGA)
- **Disk caching** — avoids redundant API calls

## Prerequisites

- **Go 1.23+**
- **SerpAPI key** — sign up at [serpapi.com](https://serpapi.com) (free tier available)
- **LLM API key**:
  - [Anuma AI](https://anuma.ai) key (primary, OpenAI-compatible router), or
  - OpenAI API key as fallback (GPT-4o-mini)

## Setup

1. Clone and build:

```bash
git clone https://github.com/kingpinXD/booker-evolve.git
cd booker-evolve
make build
```

2. Create a `.env` file in the project root:

```env
BOOKER_SERPAPI_KEY=your_serpapi_key
BOOKER_ANUMA_API_KEY=your_anuma_key
BOOKER_OPENAI_API_KEY=your_openai_key   # optional fallback if Anuma is unavailable
```

3. Verify everything works:

```bash
make test
```

## Quick Start

### Search for flights

```bash
# Direct flight search: Delhi to London, April 15
./booker search DEL LHR --date 2025-04-15

# Multi-city with stopover: Delhi → (stopover) → Toronto
# Leg 1 departs March 24, leg 2 departs March 30
./booker search DEL YYZ --date 2025-03-24 --leg2-date 2025-03-30

# Round-trip: Delhi to London
./booker search DEL LHR --date 2025-06-01 --return-date 2025-06-15

# With filters
./booker search DEL YYZ --date 2025-03-24 --leg2-date 2025-03-30 \
  --cabin business \
  --max-price 2000 \
  --preferred-airlines AC,UA \
  --departure-after 06:00 \
  --profile comfort

# JSON output
./booker search DEL LHR --date 2025-04-15 --format json

# Direct flights only
./booker search DEL LHR --date 2025-04-15 --max-stops 0
```

### Chat mode

Interactive trip planning through conversation:

```bash
./booker chat
```

Example conversation:
```
You: I want to fly from Delhi to Toronto in late March
Agent: [extracts params, searches flights, shows results]

You: Can you try business class?
Agent: [updates cabin, re-searches]

You: Compare options 1 and 3
Agent: [shows side-by-side comparison from cached results]
```

## CLI Reference

### `booker search <origin> <destination>`

| Flag | Default | Description |
|------|---------|-------------|
| `--date` | (required) | Departure date (YYYY-MM-DD) |
| `--leg2-date` | | Leg 2 departure for multi-city (YYYY-MM-DD) |
| `--return-date` | | Return date for round-trip (YYYY-MM-DD) |
| `--passengers` | 1 | Number of travelers |
| `--cabin` | economy | economy, premium_economy, business, first |
| `--flex-days` | 3 | Date flexibility +/- days |
| `--max-stops` | -1 | Max layovers per leg (-1 = any, 0 = direct) |
| `--max-price` | 0 | Max price in USD (0 = no limit) |
| `--max-results` | 5 | Results to display |
| `--profile` | budget | Ranking: budget, comfort, balanced, eco |
| `--currency` | CAD | Display currency |
| `--sort-by` | price | Sort: price, duration, departure, score |
| `--format` | table | Output: table or json |
| `--context` | | Natural language preferences |
| `--departure-after` | | Earliest departure (HH:MM) |
| `--departure-before` | | Latest departure (HH:MM) |
| `--arrival-after` | | Earliest arrival (HH:MM) |
| `--arrival-before` | | Latest arrival (HH:MM) |
| `--max-duration` | 0 | Max flight duration (e.g. 12h) |
| `--avoid-airlines` | | IATA codes to exclude (e.g. BA,LH) |
| `--preferred-airlines` | | IATA codes to keep (e.g. AC,UA) |
| `--min-stopover` | 0 | Min stopover duration (e.g. 24h) |
| `--max-stopover` | 0 | Max stopover duration (e.g. 168h) |
| `-v, --verbose` | false | Debug output |

### `booker chat`

| Flag | Default | Description |
|------|---------|-------------|
| `--currency` | CAD | Display currency |
| `--format` | table | Output: table or json |
| `--profile` | budget | Ranking profile |
| `-v, --verbose` | false | Debug output |

## How It Works

1. **Strategy selection** — an LLM picks the best search approach (direct, multi-city, or nearby airports) based on the route
2. **Stopover expansion** — for multi-city, generates candidate stopover cities from 25 pre-defined corridors (e.g. DEL→YYZ uses Istanbul, Dubai, Doha)
3. **Concurrent fetch** — searches all legs in parallel via SerpAPI (with disk caching)
4. **Filtering** — applies price, time, airline, duration, and stop filters
5. **Leg combination** — pairs outbound + onward flights with valid stopover gaps
6. **LLM ranking** — scores top candidates on cost, duration, airline quality, carbon, schedule, and connections

## Development

```bash
make build     # Build binary
make test      # Run all tests
make vet       # Go vet
make lint      # golangci-lint (requires golangci-lint installed)
make verify    # All checks: build + test + vet + lint + gofmt
make cover     # Test coverage report
```

## Project Structure

```
cmd/           CLI commands (search, chat) and display formatting
search/        Core search interfaces and strategies
  direct/      Direct flight search + flex-date + round-trip
  multicity/   Multi-city stopover search pipeline
  nearby/      Nearby airport expansion strategy
provider/      Flight data providers
  serpapi/      Google Flights via SerpAPI (active)
  cache/        Disk caching layer
llm/           LLM client (Anuma primary, OpenAI fallback)
types/         Normalized flight types (provider-agnostic)
config/        Configuration and API constants
currency/      Live exchange rate conversion
aggregator/    Generic concurrent fan-out
httpclient/    Shared HTTP client with retries
```

## License

MIT
