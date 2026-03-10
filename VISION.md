# Vision

> **PROTECTED** — This file must never be modified by the agent.

Booker is a flight booking **agent**, not a search engine. Unlike traditional flight search tools (Skyscanner, Trip.com, Google Flights) that present raw results to users, booker should understand what the user actually needs and act on their behalf.

## Core Differentiator

Traditional tools: user fills forms, gets a list, scrolls through options.
Booker: user describes intent, agent plans the trip, presents a recommendation.

## Architecture Direction

### 1. Conversational Interface

The current CLI is flags-based with no back-and-forth. The goal is a conversational layer where the agent can:

- Ask clarifying questions ("Is this a leisure trip or business? Flexible on dates?")
- Gather preferences incrementally instead of requiring all upfront
- Explain its reasoning ("I chose a stopover in Istanbul because it saves $400 and adds only 3 hours")

This should use the Anuma AI (`anuma.ai`) chat completions API, which is already wired up via the `llm` package.

### 2. Intelligent Strategy Selection

The `Picker` + `Strategy` pattern already exists. Expand this so the agent can:

- Choose between search strategies based on user context (already partially working)
- Combine strategies when appropriate (e.g., search direct AND multicity, then compare)
- Add new strategy types over time: flexible-date search, nearby-airport search, multi-destination trips

### 3. Dynamic, User-Aware Ranking

The current ranking uses fixed weight profiles (budget, comfort, balanced). The goal is ranking that adapts to the user:

- Infer preferences from conversation context ("I hate long layovers" -> penalize layover time)
- Learn from user feedback over sessions
- Consider factors beyond price: layover quality, airline reputation, connection risk
- The existing `Ranker` interface supports this — implementations can evolve without changing callers

### 4. Travel Itinerary Planning

Beyond single flight searches, the agent should plan complete trips:

- Multi-leg journeys with different dates and destinations
- Budget allocation across legs ("I have $2000 total for 3 flights")
- Travel time optimization (minimize total door-to-door time, not just flight time)
- Stopover recommendations based on interests, visa requirements, and cost savings

### 5. User Interaction Model

Build toward a session-based interaction where the agent:

1. Greets the user and asks about their trip
2. Gathers constraints through conversation (dates, budget, preferences, flexibility)
3. Runs searches using the best strategies
4. Presents curated options with explanations
5. Refines based on feedback ("Show me something cheaper" or "What if I leave a day later?")

This is the long-term target. Near-term steps: add a `chat` command alongside the existing `search` command, wire it to the LLM for multi-turn conversation, and have it build a `search.Request` from the dialogue.

## Technical Principles

- **Anuma AI for all LLM calls** — use the `llm` package with Anuma chat completions, do not add other LLM providers
- **SerpAPI for flight data** — single provider, keep it simple
- **Strategies are composable** — new search approaches plug into the existing `Strategy` interface
- **Ranking is swappable** — the `Ranker` interface allows experimentation without touching search logic
- **Cache everything** — flight data is expensive to fetch, use the cache layer aggressively

## What This Means for the Agent

When choosing what to work on, prioritize features that move booker from "search tool" toward "booking agent":

- Conversational features over more CLI flags
- Smarter strategy selection over new raw search types
- User-aware ranking over static profiles
- Trip planning capabilities over single-search improvements
- Integration quality (LLM + search + ranking working together) over isolated components
