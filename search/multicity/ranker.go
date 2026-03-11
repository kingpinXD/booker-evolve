package multicity

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"booker/llm"
	"booker/search"
)

// MaxItinerariesToRank is the maximum number of candidates to send
// to the LLM. Sending too many wastes tokens and confuses the model.
// We pre-sort by price and pick the top N.
const MaxItinerariesToRank = 15

// RankingWeights controls how the LLM prioritizes different criteria
// when scoring itineraries. All weights should sum to 100.
type RankingWeights struct {
	Cost               int // cheaper = better
	AirlineConsistency int // same airline both legs preferred
	LayoverQuality     int // 1.5-3hr connections ideal
	FlightDuration     int // shorter in-air time preferred
	StopoverCity       int // how interesting the stopover city is
	Schedule           int // reasonable departure times
}

// Traveler profile presets.
var (
	// WeightsBudget prioritizes cost above all else.
	WeightsBudget = RankingWeights{
		Cost:               45,
		AirlineConsistency: 15,
		LayoverQuality:     10,
		FlightDuration:     15,
		StopoverCity:       10,
		Schedule:           5,
	}

	// WeightsComfort prioritizes airline quality, schedule, and layovers.
	WeightsComfort = RankingWeights{
		Cost:               20,
		AirlineConsistency: 25,
		LayoverQuality:     20,
		FlightDuration:     15,
		StopoverCity:       10,
		Schedule:           10,
	}

	// WeightsBalanced is a middle ground between budget and comfort.
	WeightsBalanced = RankingWeights{
		Cost:               35,
		AirlineConsistency: 20,
		LayoverQuality:     15,
		FlightDuration:     15,
		StopoverCity:       10,
		Schedule:           5,
	}
)

// Ranker uses an LLM to score and rank candidate itineraries based on
// soft preferences that are hard to encode as simple rules.
//
// The LLM receives a structured summary of each itinerary and returns
// a JSON array of scores with reasoning.
//
// TODO(iterate): Cache LLM responses for identical itinerary sets.
type Ranker struct {
	llm     *llm.Client
	weights RankingWeights
}

// NewRanker creates a ranker with the given weights profile.
// Pass WeightsBudget, WeightsComfort, or WeightsBalanced — or a custom RankingWeights.
func NewRanker(llmClient *llm.Client, weights RankingWeights) *Ranker {
	return &Ranker{llm: llmClient, weights: weights}
}

// RankResult is a single scored itinerary from the LLM.
type RankResult struct {
	Index     int     `json:"index"`
	Score     float64 `json:"score"`
	Reasoning string  `json:"reasoning"`
}

// Rank sends candidate itineraries to the LLM for scoring and returns
// them sorted by score (best first). The original itineraries are
// mutated to include Score and Reasoning fields.
func (r *Ranker) Rank(ctx context.Context, itineraries []search.Itinerary) ([]search.Itinerary, error) {
	if len(itineraries) == 0 {
		return nil, nil
	}

	// Cap the number of itineraries to avoid token waste.
	candidates := itineraries
	if len(candidates) > MaxItinerariesToRank {
		candidates = candidates[:MaxItinerariesToRank]
	}

	prompt := buildRankingPrompt(candidates)
	sysPrompt := buildSystemPrompt(r.weights)

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: sysPrompt},
		{Role: llm.RoleUser, Content: prompt},
	}

	response, err := r.llm.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("LLM ranking request: %w", err)
	}

	results, err := parseRankingResponse(response)
	if err != nil {
		return nil, fmt.Errorf("parsing LLM ranking response: %w", err)
	}

	// Apply scores back to itineraries.
	for _, res := range results {
		if res.Index >= 0 && res.Index < len(candidates) {
			candidates[res.Index].Score = res.Score
			candidates[res.Index].Reasoning = res.Reasoning
		}
	}

	// Sort by score descending (best first).
	sorted := make([]search.Itinerary, len(candidates))
	copy(sorted, candidates)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Score > sorted[i].Score {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return sorted, nil
}

// buildSystemPrompt generates the LLM system prompt with the given weights.
func buildSystemPrompt(w RankingWeights) string {
	return fmt.Sprintf(`You are a flight search assistant that ranks multi-city itineraries.

You will receive a list of candidate itineraries, each consisting of two flight legs
with a stopover city in between. Your job is to score each itinerary from 0 to 100
based on these criteria (in rough priority order):

1. TOTAL COST (%d%%) — Cheaper is better. But don't sacrifice everything for $20 savings.

2. AIRLINE CONSISTENCY (%d%%) — Same airline on both legs is strongly preferred.
   Same alliance is good. Completely different airlines means separate bookings,
   no baggage transfer, and more hassle.

3. LAYOVER QUALITY (%d%%) — Within each leg, short connections (1.5-3 hours) are ideal.
   Very tight connections (<1hr) are risky. Long airport waits (5+ hours) are bad.
   The multi-day stopover between legs is GOOD (that's the point of the trip).

4. FLIGHT DURATION (%d%%) — Shorter total in-air time is better. But a 1-hour
   difference doesn't matter much.

5. STOPOVER CITY (%d%%) — Consider how interesting/convenient the stopover city is
   for a 2-6 day visit.

6. SCHEDULE (%d%%) — Prefer reasonable departure times. Avoid 3 AM departures.

IMPORTANT: Respond ONLY with a JSON array. No markdown, no code fences, no explanation
outside the JSON. Each element must have:
  - "index": the 0-based itinerary index
  - "score": integer 0-100
  - "reasoning": one sentence explaining the score`,
		w.Cost, w.AirlineConsistency, w.LayoverQuality,
		w.FlightDuration, w.StopoverCity, w.Schedule)
}

// buildRankingPrompt creates a human-readable summary of itineraries
// for the LLM to evaluate.
func buildRankingPrompt(itineraries []search.Itinerary) string {
	var b strings.Builder
	b.WriteString("Please rank these itineraries:\n\n")

	for i, itin := range itineraries {
		fmt.Fprintf(&b, "=== ITINERARY %d ===\n", i)
		fmt.Fprintf(&b, "Total Price: $%.2f %s\n", itin.TotalPrice.Amount, itin.TotalPrice.Currency)
		fmt.Fprintf(&b, "Total In-Air Time: %s\n", formatDuration(itin.TotalTravel))
		fmt.Fprintf(&b, "Total Trip Duration: %s\n\n", formatDuration(itin.TotalTrip))

		for j, leg := range itin.Legs {
			fmt.Fprintf(&b, "  LEG %d: $%.2f\n", j+1, leg.Flight.Price.Amount)
			for _, seg := range leg.Flight.Outbound {
				airlineInfo := seg.AirlineName
				if tag := search.Alliance(seg.Airline); tag != "" {
					airlineInfo += " [" + tag + "]"
				}
				if isRedEye(seg.DepartureTime) {
					airlineInfo += " [Red-eye]"
				}
				if seg.Overnight {
					airlineInfo += " [Overnight]"
				}
				fmt.Fprintf(&b, "    %s %s→%s (%s→%s) depart %s arrive %s [%s] %s\n",
					seg.FlightNumber,
					seg.Origin, seg.Destination,
					seg.OriginCity, seg.DestinationCity,
					seg.DepartureTime.Format("Jan 02 15:04"),
					seg.ArrivalTime.Format("Jan 02 15:04"),
					formatDuration(seg.Duration),
					airlineInfo,
				)
				if seg.LayoverDuration > 0 {
					fmt.Fprintf(&b, "      ↳ layover: %s\n", formatDuration(seg.LayoverDuration))
				}
			}
			if leg.Stopover != nil {
				fmt.Fprintf(&b, "  --- STOPOVER: %s (%s) for %s ---\n",
					leg.Stopover.City, leg.Stopover.Airport,
					formatDuration(leg.Stopover.Duration))
				if leg.Stopover.Notes != "" {
					fmt.Fprintf(&b, "      Notes: %s\n", leg.Stopover.Notes)
				}
			}
		}
		b.WriteString("\n")
	}
	return b.String()
}

// isRedEye returns true for departures between 00:00 and 04:59.
func isRedEye(t time.Time) bool {
	return t.Hour() < 5
}

func formatDuration(d time.Duration) string {
	hours := int(d.Hours())
	mins := int(d.Minutes()) % 60
	if hours >= 24 {
		days := hours / 24
		hours %= 24
		return fmt.Sprintf("%dd %dh %dm", days, hours, mins)
	}
	return fmt.Sprintf("%dh %dm", hours, mins)
}

func parseRankingResponse(response string) ([]RankResult, error) {
	// The LLM should return raw JSON, but sometimes wraps in code fences.
	cleaned := response
	cleaned = strings.TrimPrefix(cleaned, "```json")
	cleaned = strings.TrimPrefix(cleaned, "```")
	cleaned = strings.TrimSuffix(cleaned, "```")
	cleaned = strings.TrimSpace(cleaned)

	var results []RankResult
	if err := json.Unmarshal([]byte(cleaned), &results); err != nil {
		return nil, fmt.Errorf("JSON unmarshal: %w (raw response: %s)", err, response)
	}
	return results, nil
}
