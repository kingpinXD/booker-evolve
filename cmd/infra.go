package cmd

import (
	"fmt"

	"booker/config"
	"booker/httpclient"
	"booker/llm"
	"booker/provider"
	"booker/provider/cache"
	"booker/provider/serpapi"
	"booker/search"
	"booker/search/direct"
	"booker/search/multicity"
	"booker/search/nearby"
)

// buildPicker creates the search infrastructure: provider registry, strategies,
// and picker. Returns the LLM client (for chat) and raw serpapi provider (for PriceInsights).
func buildPicker(weights multicity.RankingWeights, leg2Date string) (*search.Picker, *llm.Client, *serpapi.Provider, error) {
	cfg := config.Default()
	httpClient := httpclient.New(cfg.HTTP)

	registry := provider.NewRegistry()
	raw := serpapi.New(cfg.Providers[config.ProviderSerpAPI], httpClient)
	cached := cache.Wrap(raw, ".cache/flights", 0)
	if err := registry.Register(cached); err != nil {
		return nil, nil, nil, fmt.Errorf("registering serpapi: %w", err)
	}

	llmClient := llm.New(cfg.LLM, httpClient)
	ranker := multicity.NewRanker(llmClient, weights)

	directStrategy := direct.NewSearcher(registry, ranker)
	nearbyStrategy := nearby.NewSearcher(directStrategy)
	mcSearcher := multicity.NewSearcher(registry, llmClient, weights)
	mcStrategy := multicity.NewStrategy(mcSearcher, leg2Date)

	picker := search.NewPicker(llmClient, directStrategy, nearbyStrategy, mcStrategy)
	return picker, llmClient, raw, nil
}
