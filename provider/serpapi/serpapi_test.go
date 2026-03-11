package serpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"booker/config"
	"booker/httpclient"
	"booker/provider"
	"booker/types"
)

// cannedOneWayResponse is a minimal SerpAPI response with one best and one other flight.
var cannedOneWayResponse = Response{
	BestFlights: []FlightGroup{{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{Name: "JFK Airport", ID: "JFK", Time: "2026-04-10 08:00"},
			ArrivalAirport:   Airport{Name: "LHR Airport", ID: "LHR", Time: "2026-04-10 20:00"},
			Duration:         420,
			Airline:          "British Airways",
			FlightNumber:     "BA 117",
			TravelClass:      "Economy",
		}},
		TotalDuration: 420,
		Price:         450,
		BookingToken:  "book_best",
	}},
	OtherFlights: []FlightGroup{{
		Flights: []FlightSegment{{
			DepartureAirport: Airport{Name: "JFK Airport", ID: "JFK", Time: "2026-04-10 14:00"},
			ArrivalAirport:   Airport{Name: "LHR Airport", ID: "LHR", Time: "2026-04-11 02:00"},
			Duration:         420,
			Airline:          "Virgin Atlantic",
			FlightNumber:     "VS 10",
			TravelClass:      "Economy",
		}},
		TotalDuration: 420,
		Price:         380,
		BookingToken:  "book_other",
	}},
}

// newTestProvider creates a Provider pointed at the given httptest server URL.
func newTestProvider(serverURL string) *Provider {
	cfg := config.ProviderConfig{
		APIKey:  "test-key-123",
		BaseURL: serverURL,
	}
	httpCfg := config.HTTPConfig{
		Timeout:         5 * time.Second,
		MaxIdleConns:    2,
		IdleConnTimeout: 10 * time.Second,
		MaxRetries:      1,
		RetryBaseDelay:  10 * time.Millisecond,
	}
	return New(cfg, httpclient.New(httpCfg))
}

func TestSearch_OneWay(t *testing.T) {
	var captured url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cannedOneWayResponse)
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	dep, _ := time.Parse("2006-01-02", "2026-04-10")
	req := types.SearchRequest{
		Origin:        "JFK",
		Destination:   "LHR",
		DepartureDate: dep,
		Passengers:    2,
		CabinClass:    types.CabinEconomy,
	}

	flights, err := p.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify URL params sent to server.
	if got := captured.Get(config.SerpAPIParamEngine); got != config.SerpAPIEngineFlights {
		t.Errorf("engine = %q, want %q", got, config.SerpAPIEngineFlights)
	}
	if got := captured.Get(config.SerpAPIParamAPIKey); got != "test-key-123" {
		t.Errorf("api_key = %q, want %q", got, "test-key-123")
	}
	if got := captured.Get(config.SerpAPIParamDeparture); got != "JFK" {
		t.Errorf("departure_id = %q, want %q", got, "JFK")
	}
	if got := captured.Get(config.SerpAPIParamArrival); got != "LHR" {
		t.Errorf("arrival_id = %q, want %q", got, "LHR")
	}
	if got := captured.Get(config.SerpAPIParamDate); got != "2026-04-10" {
		t.Errorf("outbound_date = %q, want %q", got, "2026-04-10")
	}
	if got := captured.Get(config.SerpAPIParamType); got != config.SerpAPITypeOneWay {
		t.Errorf("type = %q, want %q (one-way)", got, config.SerpAPITypeOneWay)
	}
	if got := captured.Get(config.SerpAPIParamAdults); got != "2" {
		t.Errorf("adults = %q, want %q", got, "2")
	}
	if got := captured.Get(config.SerpAPIParamClass); got != config.SerpAPIClassEconomy {
		t.Errorf("travel_class = %q, want %q", got, config.SerpAPIClassEconomy)
	}
	// MaxStops defaults to 0 (direct-only), so stops=0 should be present.
	if got := captured.Get(config.SerpAPIParamStops); got != "0" {
		t.Errorf("stops = %q, want %q", got, "0")
	}

	// Verify flights returned.
	if len(flights) != 2 {
		t.Fatalf("flights = %d, want 2", len(flights))
	}
	if flights[0].Price.Amount != 450 {
		t.Errorf("flights[0].Price = %v, want 450", flights[0].Price.Amount)
	}
	if flights[1].Price.Amount != 380 {
		t.Errorf("flights[1].Price = %v, want 380", flights[1].Price.Amount)
	}
	if flights[0].Provider != config.ProviderSerpAPI {
		t.Errorf("provider = %q, want %q", flights[0].Provider, config.ProviderSerpAPI)
	}
}

func TestSearch_DirectOnlyStopsParam(t *testing.T) {
	var captured url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cannedOneWayResponse)
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	dep, _ := time.Parse("2006-01-02", "2026-04-10")
	req := types.SearchRequest{
		Origin:        "JFK",
		Destination:   "LHR",
		DepartureDate: dep,
		Passengers:    1,
		MaxStops:      0, // direct-only
	}

	_, err := p.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := captured.Get(config.SerpAPIParamStops); got != "0" {
		t.Errorf("stops = %q, want %q", got, "0")
	}
}

func TestSearch_NonDirectNoStopsParam(t *testing.T) {
	var captured url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cannedOneWayResponse)
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	dep, _ := time.Parse("2006-01-02", "2026-04-10")
	req := types.SearchRequest{
		Origin:        "JFK",
		Destination:   "LHR",
		DepartureDate: dep,
		Passengers:    1,
		MaxStops:      1, // allow 1 stop
	}

	_, err := p.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if captured.Has(config.SerpAPIParamStops) {
		t.Errorf("stops param should be absent for MaxStops=1, got %q", captured.Get(config.SerpAPIParamStops))
	}
}

func TestSearch_RoundTrip(t *testing.T) {
	var captured url.Values
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		captured = r.URL.Query()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(cannedOneWayResponse)
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	dep, _ := time.Parse("2006-01-02", "2026-04-10")
	ret, _ := time.Parse("2006-01-02", "2026-04-17")
	req := types.SearchRequest{
		Origin:        "JFK",
		Destination:   "LHR",
		DepartureDate: dep,
		ReturnDate:    ret,
		Passengers:    1,
		CabinClass:    types.CabinBusiness,
	}

	_, err := p.Search(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got := captured.Get(config.SerpAPIParamType); got != config.SerpAPITypeRoundTrip {
		t.Errorf("type = %q, want %q (round-trip)", got, config.SerpAPITypeRoundTrip)
	}
	if got := captured.Get(config.SerpAPIParamClass); got != config.SerpAPIClassBusiness {
		t.Errorf("travel_class = %q, want %q", got, config.SerpAPIClassBusiness)
	}
}

func TestSearch_ServerError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	dep, _ := time.Parse("2006-01-02", "2026-04-10")
	req := types.SearchRequest{
		Origin:        "JFK",
		Destination:   "LHR",
		DepartureDate: dep,
		Passengers:    1,
	}

	_, err := p.Search(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for 400 response, got nil")
	}
}

func TestMapCabinClassToSerpAPI(t *testing.T) {
	tests := []struct {
		input types.CabinClass
		want  string
	}{
		{types.CabinEconomy, config.SerpAPIClassEconomy},
		{types.CabinPremiumEconomy, config.SerpAPIClassPremiumEconomy},
		{types.CabinBusiness, config.SerpAPIClassBusiness},
		{types.CabinFirst, config.SerpAPIClassFirst},
		{"unknown", config.SerpAPIClassEconomy}, // default
	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			got := mapCabinClassToSerpAPI(tt.input)
			if got != tt.want {
				t.Errorf("mapCabinClassToSerpAPI(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestBuildMultiCityJSON(t *testing.T) {
	req := MultiCityRequest{
		Origin:      "DEL",
		Stopover:    "HKG",
		Destination: "YYZ",
		Leg1Date:    "2026-03-24",
		Leg2Date:    "2026-03-30",
	}

	result, err := buildMultiCityJSON(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var legs []map[string]string
	if err := json.Unmarshal([]byte(result), &legs); err != nil {
		t.Fatalf("invalid JSON output: %v", err)
	}

	if len(legs) != 2 {
		t.Fatalf("legs = %d, want 2", len(legs))
	}

	// Leg 1: DEL -> HKG
	if legs[0]["departure_id"] != "DEL" {
		t.Errorf("leg1 departure_id = %q, want DEL", legs[0]["departure_id"])
	}
	if legs[0]["arrival_id"] != "HKG" {
		t.Errorf("leg1 arrival_id = %q, want HKG", legs[0]["arrival_id"])
	}
	if legs[0]["date"] != "2026-03-24" {
		t.Errorf("leg1 date = %q, want 2026-03-24", legs[0]["date"])
	}

	// Leg 2: HKG -> YYZ
	if legs[1]["departure_id"] != "HKG" {
		t.Errorf("leg2 departure_id = %q, want HKG", legs[1]["departure_id"])
	}
	if legs[1]["arrival_id"] != "YYZ" {
		t.Errorf("leg2 arrival_id = %q, want YYZ", legs[1]["arrival_id"])
	}
	if legs[1]["date"] != "2026-03-30" {
		t.Errorf("leg2 date = %q, want 2026-03-30", legs[1]["date"])
	}
}

func TestSearchMultiCity_TwoStepFlow(t *testing.T) {
	// Step 1 response: two leg1 options with departure tokens.
	step1Resp := Response{
		BestFlights: []FlightGroup{{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{Name: "DEL Airport", ID: "DEL", Time: "2026-03-24 06:00"},
				ArrivalAirport:   Airport{Name: "HKG Airport", ID: "HKG", Time: "2026-03-24 14:00"},
				Duration:         480,
				Airline:          "Cathay Pacific",
				FlightNumber:     "CX 694",
				TravelClass:      "Economy",
			}},
			TotalDuration:  480,
			Price:          300,
			DepartureToken: "token_leg1_cheap",
		}},
		OtherFlights: []FlightGroup{{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{Name: "DEL Airport", ID: "DEL", Time: "2026-03-24 10:00"},
				ArrivalAirport:   Airport{Name: "HKG Airport", ID: "HKG", Time: "2026-03-24 18:00"},
				Duration:         480,
				Airline:          "Air India",
				FlightNumber:     "AI 310",
				TravelClass:      "Economy",
			}},
			TotalDuration:  480,
			Price:          400,
			DepartureToken: "token_leg1_expensive",
		}},
	}

	// Step 2 response: one leg2 option with combined price.
	step2Resp := Response{
		BestFlights: []FlightGroup{{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{Name: "HKG Airport", ID: "HKG", Time: "2026-03-27 10:00"},
				ArrivalAirport:   Airport{Name: "YYZ Airport", ID: "YYZ", Time: "2026-03-27 22:00"},
				Duration:         720,
				Airline:          "Air Canada",
				FlightNumber:     "AC 16",
				TravelClass:      "Economy",
			}},
			TotalDuration: 720,
			Price:         800,
		}},
	}

	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")

		// Step 1 has no departure_token param; step 2 does.
		if r.URL.Query().Get(config.SerpAPIParamDepartureToken) != "" {
			_ = json.NewEncoder(w).Encode(step2Resp)
		} else {
			_ = json.NewEncoder(w).Encode(step1Resp)
		}
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	req := provider.MultiCityRequest{
		Origin:      "DEL",
		Stopover:    "HKG",
		Destination: "YYZ",
		Leg1Date:    "2026-03-24",
		Leg2Date:    "2026-03-27",
		Passengers:  1,
		CabinClass:  types.CabinEconomy,
		TopN:        2,
	}

	results, err := p.SearchMultiCity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// 1 step1 request + 2 step2 requests (one per leg1 option) = 3 total.
	if requestCount != 3 {
		t.Errorf("request count = %d, want 3", requestCount)
	}

	// 2 leg1 options x 1 leg2 option each = 2 results.
	if len(results) != 2 {
		t.Fatalf("results = %d, want 2", len(results))
	}

	// Price should be the combined step2 price.
	for i, r := range results {
		if r.Price.Amount != 800 {
			t.Errorf("results[%d].Price = %v, want 800", i, r.Price.Amount)
		}
		if r.Price.Currency != "USD" {
			t.Errorf("results[%d].Price.Currency = %q, want USD", i, r.Price.Currency)
		}
		if len(r.Leg1.Outbound) != 1 {
			t.Errorf("results[%d].Leg1 segments = %d, want 1", i, len(r.Leg1.Outbound))
		}
		if len(r.Leg2.Outbound) != 1 {
			t.Errorf("results[%d].Leg2 segments = %d, want 1", i, len(r.Leg2.Outbound))
		}
	}

	// Leg1 carries total price, leg2 has zero price.
	if results[0].Leg1.Price.Amount != 800 {
		t.Errorf("leg1 price = %v, want 800", results[0].Leg1.Price.Amount)
	}
	if results[0].Leg2.Price.Amount != 0 {
		t.Errorf("leg2 price = %v, want 0", results[0].Leg2.Price.Amount)
	}
}

func TestSearchMultiCity_EmptyStep1(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(Response{}) // empty response
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	req := provider.MultiCityRequest{
		Origin:      "DEL",
		Stopover:    "HKG",
		Destination: "YYZ",
		Leg1Date:    "2026-03-24",
		Leg2Date:    "2026-03-30",
		Passengers:  1,
		CabinClass:  types.CabinEconomy,
	}

	results, err := p.SearchMultiCity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("results = %d, want 0 (empty step1)", len(results))
	}
}

func TestSearchMultiCity_NoDepartureToken(t *testing.T) {
	// Step1 returns flights but without departure tokens -- step2 should be skipped.
	step1Resp := Response{
		BestFlights: []FlightGroup{{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 06:00"},
				ArrivalAirport:   Airport{ID: "HKG", Time: "2026-03-24 14:00"},
				Duration:         480,
				FlightNumber:     "CX 694",
			}},
			TotalDuration:  480,
			Price:          300,
			DepartureToken: "", // no token
		}},
	}

	requestCount := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(step1Resp)
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	req := provider.MultiCityRequest{
		Origin:      "DEL",
		Stopover:    "HKG",
		Destination: "YYZ",
		Leg1Date:    "2026-03-24",
		Leg2Date:    "2026-03-30",
		Passengers:  1,
		CabinClass:  types.CabinEconomy,
	}

	results, err := p.SearchMultiCity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Only 1 request (step1); step2 skipped because no departure token.
	if requestCount != 1 {
		t.Errorf("request count = %d, want 1 (step2 skipped)", requestCount)
	}
	if len(results) != 0 {
		t.Errorf("results = %d, want 0", len(results))
	}
}

func TestSearchMultiCity_TopNDefault(t *testing.T) {
	// Step 1: 5 options to verify TopN defaults to 3.
	var step1Groups []FlightGroup
	for i := range 5 {
		step1Groups = append(step1Groups, FlightGroup{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 06:00"},
				ArrivalAirport:   Airport{ID: "HKG", Time: "2026-03-24 14:00"},
				Duration:         480,
				FlightNumber:     "CX 100",
			}},
			TotalDuration:  480,
			Price:          200 + i*100,
			DepartureToken: "token_" + string(rune('A'+i)),
		})
	}

	step2Resp := Response{
		BestFlights: []FlightGroup{{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{ID: "HKG", Time: "2026-03-27 10:00"},
				ArrivalAirport:   Airport{ID: "YYZ", Time: "2026-03-27 22:00"},
				Duration:         720,
				FlightNumber:     "AC 16",
			}},
			TotalDuration: 720,
			Price:         900,
		}},
	}

	step2Count := 0
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get(config.SerpAPIParamDepartureToken) != "" {
			step2Count++
			_ = json.NewEncoder(w).Encode(step2Resp)
		} else {
			_ = json.NewEncoder(w).Encode(Response{OtherFlights: step1Groups})
		}
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	// TopN = 0 should default to 3.
	req := provider.MultiCityRequest{
		Origin:      "DEL",
		Stopover:    "HKG",
		Destination: "YYZ",
		Leg1Date:    "2026-03-24",
		Leg2Date:    "2026-03-27",
		Passengers:  1,
		CabinClass:  types.CabinEconomy,
		TopN:        0,
	}

	results, err := p.SearchMultiCity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// TopN defaults to 3, so only 3 step2 requests.
	if step2Count != 3 {
		t.Errorf("step2 requests = %d, want 3 (TopN default)", step2Count)
	}
	if len(results) != 3 {
		t.Errorf("results = %d, want 3", len(results))
	}
}

func TestSearchMultiCity_Step2Error(t *testing.T) {
	step1Resp := Response{
		BestFlights: []FlightGroup{{
			Flights: []FlightSegment{{
				DepartureAirport: Airport{ID: "DEL", Time: "2026-03-24 06:00"},
				ArrivalAirport:   Airport{ID: "HKG", Time: "2026-03-24 14:00"},
				Duration:         480,
				FlightNumber:     "CX 694",
			}},
			TotalDuration:  480,
			Price:          300,
			DepartureToken: "token_a",
		}},
	}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Query().Get(config.SerpAPIParamDepartureToken) != "" {
			// Step 2 returns a client error.
			http.Error(w, "bad request", http.StatusBadRequest)
		} else {
			_ = json.NewEncoder(w).Encode(step1Resp)
		}
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	req := provider.MultiCityRequest{
		Origin:      "DEL",
		Stopover:    "HKG",
		Destination: "YYZ",
		Leg1Date:    "2026-03-24",
		Leg2Date:    "2026-03-30",
		Passengers:  1,
		CabinClass:  types.CabinEconomy,
	}

	// Step2 errors are logged and skipped, not returned.
	results, err := p.SearchMultiCity(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("results = %d, want 0 (step2 failed)", len(results))
	}
}

func TestProviderName(t *testing.T) {
	p := &Provider{}
	if p.Name() != config.ProviderSerpAPI {
		t.Errorf("Name() = %q, want %q", p.Name(), config.ProviderSerpAPI)
	}
}

func TestSearch_Step1Error(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "bad request", http.StatusBadRequest)
	}))
	defer ts.Close()

	p := newTestProvider(ts.URL)
	req := provider.MultiCityRequest{
		Origin:      "DEL",
		Stopover:    "HKG",
		Destination: "YYZ",
		Leg1Date:    "2026-03-24",
		Leg2Date:    "2026-03-30",
		Passengers:  1,
		CabinClass:  types.CabinEconomy,
	}

	_, err := p.SearchMultiCity(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for step1 failure, got nil")
	}
}
