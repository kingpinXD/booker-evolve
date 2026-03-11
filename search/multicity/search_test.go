package multicity

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"booker/config"
	"booker/httpclient"
	"booker/llm"
	"booker/provider"
	"booker/search"
	"booker/types"
)

// mockProv implements provider.Provider for unit tests.
type mockProv struct {
	name    config.ProviderName
	flights []types.Flight
	err     error
}

func (m *mockProv) Name() config.ProviderName { return m.name }

func (m *mockProv) Search(_ context.Context, _ types.SearchRequest) ([]types.Flight, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.flights, nil
}

// mockMultiCityProv implements both provider.Provider and provider.MultiCitySearcher.
type mockMultiCityProv struct {
	mockProv
	mcResults []provider.MultiCityResult
	mcErr     error
}

func (m *mockMultiCityProv) SearchMultiCity(_ context.Context, _ provider.MultiCityRequest) ([]provider.MultiCityResult, error) {
	if m.mcErr != nil {
		return nil, m.mcErr
	}
	return m.mcResults, nil
}

// newTestSearcher creates a Searcher backed by a mock provider and an httptest LLM server.
// The returned cleanup function must be called to shut down the server.
func newTestSearcher(t *testing.T, flights []types.Flight, llmHandler http.HandlerFunc) *Searcher {
	t.Helper()

	reg := provider.NewRegistry()
	if err := reg.Register(&mockProv{name: "mock", flights: flights}); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(llmHandler)
	t.Cleanup(srv.Close)

	httpCfg := config.HTTPConfig{
		Timeout:      5 * time.Second,
		MaxRetries:   1,
		MaxIdleConns: 2,
	}
	llmClient := llm.New(config.LLMConfig{
		BaseURL:    srv.URL,
		APIKey:     "test-key",
		Model:      "test-model",
		MaxTokens:  100,
		AuthHeader: "Authorization",
		Provider:   "test",
	}, httpclient.New(httpCfg))

	return NewSearcher(reg, llmClient, WeightsBudget)
}

// llmRankingHandler returns an http.HandlerFunc that responds with valid ranking JSON.
func llmRankingHandler(count int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var results []RankResult
		for i := 0; i < count; i++ {
			results = append(results, RankResult{
				Index:     i,
				Score:     float64(90 - i*5),
				Reasoning: fmt.Sprintf("itinerary %d", i),
			})
		}
		resp := map[string]any{
			"choices": []map[string]any{
				{"message": map[string]string{
					"role":    "assistant",
					"content": mustJSON(results),
				}},
			},
			"usage": map[string]int{
				"prompt_tokens":     100,
				"completion_tokens": 50,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}
}

// llmErrorHandler returns an http.HandlerFunc that returns a 500 error.
func llmErrorHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"message":"LLM unavailable"}}`))
	}
}

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

// testStopover returns a single stopover city for testing.
func testStopover() StopoverCity {
	return StopoverCity{
		City:    "Hong Kong",
		Airport: "HKG",
		Region:  "east_asia",
		MinStay: types.DefaultMinStopover,
		MaxStay: types.DefaultMaxStopover,
	}
}

// validLeg1 creates a flight suitable for leg1 (origin -> stopover) that departs
// on the basetime date and passes all filters.
func validLeg1() types.Flight {
	return makeFlight("CX", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300)
}

// validLeg2 creates a flight suitable for leg2 (stopover -> destination) that departs
// 3 days after basetime (within the default stopover window).
func validLeg2() types.Flight {
	return makeFlight("AC", "HKG", "YYZ",
		basetime.Add(72*time.Hour), basetime.Add(88*time.Hour), 500)
}

func TestSearch_OneStopover(t *testing.T) {
	// The mock provider returns both leg1 and leg2 flights on every search call.
	// The pipeline should combine them into at least one itinerary.
	flights := []types.Flight{validLeg1(), validLeg2()}
	searcher := newTestSearcher(t, flights, llmRankingHandler(15))

	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search() returned 0 itineraries, expected at least 1")
	}

	// Verify the itinerary structure.
	itin := results[0]
	if len(itin.Legs) != 2 {
		t.Fatalf("expected 2 legs, got %d", len(itin.Legs))
	}
	if itin.Legs[0].Stopover == nil || itin.Legs[0].Stopover.City != "Hong Kong" {
		t.Error("first leg should have Hong Kong stopover")
	}
	if itin.TotalPrice.Amount <= 0 {
		t.Errorf("TotalPrice = %.0f, expected > 0", itin.TotalPrice.Amount)
	}
}

func TestSearch_InvalidDepartureDate(t *testing.T) {
	searcher := newTestSearcher(t, nil, llmErrorHandler())

	_, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: "not-a-date",
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
	})
	if err == nil {
		t.Fatal("expected error for invalid departure date")
	}
}

func TestSearch_InvalidLeg2Date(t *testing.T) {
	searcher := newTestSearcher(t, nil, llmErrorHandler())

	_, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Leg2Date:      "bad-date",
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
	})
	if err == nil {
		t.Fatal("expected error for invalid leg2 date")
	}
}

func TestSearch_EmptyStopoversUnknownRoute(t *testing.T) {
	// When params.Stopovers is empty and route is unknown, StopoversForRoute
	// returns nil, which causes Search to return an error.
	searcher := newTestSearcher(t, nil, llmErrorHandler())

	_, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "XXX",
		Destination:   "YYY",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		Stopovers:     []StopoverCity{}, // triggers fallback
	})
	if err == nil {
		t.Fatal("expected error for unknown route with no stopovers")
	}
}

func TestSearch_MaxResultsCap(t *testing.T) {
	// Create multiple distinguishable flights that will produce many itineraries.
	var flights []types.Flight
	for i := 0; i < 5; i++ {
		offset := time.Duration(i) * time.Hour
		flights = append(flights,
			makeFlight("CX", "DEL", "HKG",
				basetime.Add(offset), basetime.Add(8*time.Hour+offset), 300+float64(i*10)),
		)
	}
	for i := 0; i < 5; i++ {
		offset := time.Duration(i) * time.Hour
		flights = append(flights,
			makeFlight("AC", "HKG", "YYZ",
				basetime.Add(72*time.Hour+offset), basetime.Add(88*time.Hour+offset), 500+float64(i*10)),
		)
	}

	searcher := newTestSearcher(t, flights, llmRankingHandler(15))

	maxResults := 2
	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
		MaxResults:    maxResults,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) > maxResults {
		t.Errorf("Search() returned %d results, want <= %d", len(results), maxResults)
	}
}

func TestSearch_RankerFallback(t *testing.T) {
	// When the LLM is unavailable, Search should fall back to price-sorted results.
	flights := []types.Flight{validLeg1(), validLeg2()}
	searcher := newTestSearcher(t, flights, llmErrorHandler())

	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
	})
	if err != nil {
		t.Fatalf("Search() should not error on LLM failure, got: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search() returned 0 results, expected fallback to price sort")
	}

	// Verify scores are zero (unranked).
	for i, r := range results {
		if r.Score != 0 {
			t.Errorf("result[%d].Score = %.0f, want 0 (unranked)", i, r.Score)
		}
	}
}

func TestSearch_RankerFallbackWithMaxResults(t *testing.T) {
	// Verify MaxResults is applied even on the fallback path.
	var flights []types.Flight
	for i := 0; i < 4; i++ {
		offset := time.Duration(i) * time.Hour
		flights = append(flights,
			makeFlight("CX", "DEL", "HKG",
				basetime.Add(offset), basetime.Add(8*time.Hour+offset), 300+float64(i*10)),
		)
	}
	for i := 0; i < 4; i++ {
		offset := time.Duration(i) * time.Hour
		flights = append(flights,
			makeFlight("AC", "HKG", "YYZ",
				basetime.Add(72*time.Hour+offset), basetime.Add(88*time.Hour+offset), 500+float64(i*10)),
		)
	}

	searcher := newTestSearcher(t, flights, llmErrorHandler())

	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
		MaxResults:    2,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) > 2 {
		t.Errorf("Search() returned %d results, want <= 2 on fallback path", len(results))
	}
}

func TestSearch_FlexDaysDefault(t *testing.T) {
	// FlexDays=0 should default to 3.
	flights := []types.Flight{validLeg1(), validLeg2()}
	searcher := newTestSearcher(t, flights, llmRankingHandler(15))

	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      0, // should default to 3
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search() returned 0 results with default FlexDays")
	}
}

func TestSearch_NoMatchingFlights(t *testing.T) {
	// Provider returns flights but none that form valid itinerary pairs.
	// Leg2 departs only 10 hours after leg1 arrives (below 48h MinStay).
	flights := []types.Flight{
		makeFlight("CX", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300),
		makeFlight("AC", "HKG", "YYZ", basetime.Add(10*time.Hour), basetime.Add(26*time.Hour), 500),
	}
	searcher := newTestSearcher(t, flights, llmErrorHandler())

	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	// No valid combinations, so nil result without error.
	if results != nil {
		t.Errorf("expected nil results for no matching flights, got %d", len(results))
	}
}

func TestSearch_MultiCityProvider(t *testing.T) {
	// Verify that multi-city provider results are merged into the output.
	leg1Dep := basetime
	leg1Arr := basetime.Add(8 * time.Hour)
	leg2Dep := basetime.Add(72 * time.Hour)
	leg2Arr := basetime.Add(88 * time.Hour)

	mcProv := &mockMultiCityProv{
		mockProv: mockProv{
			name:    "mc-mock",
			flights: nil, // no one-way results
		},
		mcResults: []provider.MultiCityResult{
			{
				Leg1: types.Flight{
					Outbound:      []types.Segment{{Airline: "CX", Origin: "DEL", Destination: "HKG", DepartureTime: leg1Dep, ArrivalTime: leg1Arr}},
					TotalDuration: 8 * time.Hour,
					Price:         types.Money{Amount: 250, Currency: "USD"},
				},
				Leg2: types.Flight{
					Outbound:      []types.Segment{{Airline: "AC", Origin: "HKG", Destination: "YYZ", DepartureTime: leg2Dep, ArrivalTime: leg2Arr}},
					TotalDuration: 16 * time.Hour,
					Price:         types.Money{Amount: 400, Currency: "USD"},
				},
				Price: types.Money{Amount: 650, Currency: "USD"},
			},
		},
	}

	reg := provider.NewRegistry()
	if err := reg.Register(mcProv); err != nil {
		t.Fatal(err)
	}

	srv := httptest.NewServer(llmRankingHandler(15))
	t.Cleanup(srv.Close)

	httpCfg := config.HTTPConfig{Timeout: 5 * time.Second, MaxRetries: 1, MaxIdleConns: 2}
	llmClient := llm.New(config.LLMConfig{
		BaseURL:    srv.URL,
		APIKey:     "test",
		Model:      "test",
		MaxTokens:  100,
		AuthHeader: "Authorization",
		Provider:   "test",
	}, httpclient.New(httpCfg))

	searcher := NewSearcher(reg, llmClient, WeightsBudget)

	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("expected multi-city results to be merged")
	}
}

func TestSearch_PreferredAllianceFilter(t *testing.T) {
	// AC is Star Alliance, AA is OneWorld.
	// With PreferredAlliance="Star Alliance", only AC flights should survive.
	// Use different departure times to avoid dedup collisions.
	flights := []types.Flight{
		makeFlight("AC", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300),
		makeFlight("AA", "DEL", "HKG", basetime.Add(1*time.Hour), basetime.Add(9*time.Hour), 280),
		makeFlight("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), basetime.Add(88*time.Hour), 500),
		makeFlight("AA", "HKG", "YYZ", basetime.Add(73*time.Hour), basetime.Add(89*time.Hour), 480),
	}
	searcher := newTestSearcher(t, flights, llmRankingHandler(15))

	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:            "DEL",
		Destination:       "YYZ",
		DepartureDate:     basetime.Format(DateLayout),
		Passengers:        1,
		CabinClass:        types.CabinEconomy,
		Stopovers:         []StopoverCity{testStopover()},
		FlexDays:          3,
		PreferredAlliance: "Star Alliance",
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("Search() returned 0 results, expected at least 1")
	}

	// Every leg in every itinerary should be Star Alliance.
	for i, itin := range results {
		for j, leg := range itin.Legs {
			for _, seg := range leg.Flight.Outbound {
				alliance := search.Alliance(seg.Airline)
				if alliance != "Star Alliance" {
					t.Errorf("result[%d].leg[%d] airline %s is %q, want Star Alliance",
						i, j, seg.Airline, alliance)
				}
			}
		}
	}
}

func TestSearch_MaxPriceFilter(t *testing.T) {
	// Leg1 = $300, Leg2 = $500, total = $800.
	flights := []types.Flight{validLeg1(), validLeg2()}

	// MaxPrice=700 should filter out the $800 combined itinerary.
	searcher := newTestSearcher(t, flights, llmRankingHandler(15))
	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
		MaxPrice:      700,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("MaxPrice=700: got %d results, want 0 (combined price is $800)", len(results))
	}

	// MaxPrice=900 should keep the $800 combined itinerary.
	searcher2 := newTestSearcher(t, flights, llmRankingHandler(15))
	results2, err := searcher2.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
		MaxPrice:      900,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if len(results2) == 0 {
		t.Fatal("MaxPrice=900: got 0 results, expected at least 1 (combined price is $800)")
	}
}

func TestSearch_ZeroPriceFiltered(t *testing.T) {
	// Flights with $0 price should be filtered out.
	flights := []types.Flight{
		makeFlight("CX", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 0),
		makeFlight("AC", "HKG", "YYZ", basetime.Add(72*time.Hour), basetime.Add(88*time.Hour), 0),
	}
	searcher := newTestSearcher(t, flights, llmErrorHandler())

	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for zero-price flights, got %d", len(results))
	}
}

func TestSearch_BlockedAirlineFiltered(t *testing.T) {
	// Flights on blocked airlines (e.g. Emirates "EK") should be filtered out.
	flights := []types.Flight{
		makeFlight("EK", "DEL", "HKG", basetime, basetime.Add(8*time.Hour), 300),
		makeFlight("EK", "HKG", "YYZ", basetime.Add(72*time.Hour), basetime.Add(88*time.Hour), 500),
	}
	searcher := newTestSearcher(t, flights, llmErrorHandler())

	results, err := searcher.Search(context.Background(), SearchParams{
		Origin:        "DEL",
		Destination:   "YYZ",
		DepartureDate: basetime.Format(DateLayout),
		Passengers:    1,
		CabinClass:    types.CabinEconomy,
		Stopovers:     []StopoverCity{testStopover()},
		FlexDays:      3,
	})
	if err != nil {
		t.Fatalf("Search() error: %v", err)
	}
	if results != nil {
		t.Errorf("expected nil results for blocked airlines, got %d", len(results))
	}
}
