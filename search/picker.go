package search

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"booker/llm"
)

// ChatCompleter abstracts the LLM chat completion call so Picker can be tested
// with a mock. The concrete implementation is *llm.Client.
type ChatCompleter interface {
	ChatCompletion(ctx context.Context, messages []llm.Message) (string, error)
}

// Picker uses an LLM to select the best search strategy based on user context.
// Falls back to heuristics when no context is provided or the LLM is unavailable.
type Picker struct {
	llm        ChatCompleter
	strategies []Strategy
	ranker     Ranker
}

// NewPicker creates a strategy picker. Pass all available strategies.
func NewPicker(llmClient ChatCompleter, strategies ...Strategy) *Picker {
	return &Picker{llm: llmClient, strategies: strategies}
}

// SetRanker configures the ranker used by CompositeStrategy when the LLM
// picks "both" mode. Without a ranker, merged results are only deduplicated.
func (p *Picker) SetRanker(r Ranker) {
	p.ranker = r
}

// Pick analyzes the request context and returns the most appropriate strategy
// along with a human-readable reason for the choice.
func (p *Picker) Pick(ctx context.Context, req Request) (Strategy, string, error) {
	if len(p.strategies) == 0 {
		return nil, "", fmt.Errorf("no strategies registered")
	}
	if len(p.strategies) == 1 {
		return p.strategies[0], "only one strategy registered", nil
	}

	// No context or no LLM — use heuristic fallback.
	if req.Context == "" || p.llm == nil {
		s, reason := p.fallback(req)
		return s, reason, nil
	}

	picked, reason, err := p.pickWithLLM(ctx, req)
	if err != nil {
		s, _ := p.fallback(req)
		return s, "LLM unavailable, using default", nil
	}
	return picked, reason, nil
}

// fallback inspects the request shape and returns an appropriate strategy.
// When Leg2Date is set, prefers "multicity"; otherwise prefers "direct".
func (p *Picker) fallback(req Request) (Strategy, string) {
	preferred := "direct"
	reason := "default for single-leg route"
	if req.Leg2Date != "" {
		preferred = "multicity"
		reason = "default for multi-city route"
	}
	for _, s := range p.strategies {
		if s.Name() == preferred {
			return s, reason
		}
	}
	return p.strategies[0], reason
}

type pickerResult struct {
	Strategy string `json:"strategy"`
	Reason   string `json:"reason"`
}

func (p *Picker) pickWithLLM(ctx context.Context, req Request) (Strategy, string, error) {
	sysPrompt := p.buildSystemPrompt()
	userPrompt := fmt.Sprintf(
		"Route: %s -> %s\nDate: %s\nPassengers: %d\nCabin: %s\nContext: %s",
		req.Origin, req.Destination, req.DepartureDate,
		req.Passengers, req.CabinClass, req.Context,
	)

	messages := []llm.Message{
		{Role: llm.RoleSystem, Content: sysPrompt},
		{Role: llm.RoleUser, Content: userPrompt},
	}

	response, err := p.llm.ChatCompletion(ctx, messages)
	if err != nil {
		return nil, "", err
	}

	// Strip markdown code fences if present.
	cleaned := llm.StripCodeFences(response)

	var result pickerResult
	if err := json.Unmarshal([]byte(cleaned), &result); err != nil {
		return nil, "", fmt.Errorf("parsing picker response: %w", err)
	}

	if result.Strategy == "both" {
		return NewCompositeStrategy(p.ranker, p.strategies...), result.Reason, nil
	}

	for _, s := range p.strategies {
		if s.Name() == result.Strategy {
			return s, result.Reason, nil
		}
	}
	return nil, "", fmt.Errorf("unknown strategy %q from LLM", result.Strategy)
}

func (p *Picker) buildSystemPrompt() string {
	var b strings.Builder
	b.WriteString("You are a flight search strategy selector. Given the user's route and context, pick the best search strategy.\n\n")
	b.WriteString("Available strategies:\n")
	for _, s := range p.strategies {
		fmt.Fprintf(&b, "- %s: %s\n", s.Name(), s.Description())
	}
	b.WriteString("- both: Run all strategies in parallel and merge results. Use when the best approach is unclear or the user wants to compare options.\n")
	b.WriteString("\nRespond ONLY with JSON: {\"strategy\": \"<name>\", \"reason\": \"<one sentence>\"}")
	return b.String()
}
