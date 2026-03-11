package serpapi

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sort"
	"strconv"

	"booker/config"
	"booker/httpclient"
	"booker/provider"
	"booker/search"
	"booker/types"
)

// Provider implements provider.Provider for SerpAPI Google Flights.
type Provider struct {
	cfg          config.ProviderConfig
	http         *httpclient.Client
	lastInsights search.PriceInsights
}

// New creates a SerpAPI provider.
func New(cfg config.ProviderConfig, http *httpclient.Client) *Provider {
	return &Provider{cfg: cfg, http: http}
}

// Name returns the provider identifier.
func (p *Provider) Name() config.ProviderName {
	return config.ProviderSerpAPI
}

// LastPriceInsights returns price insights from the most recent Search call.
func (p *Provider) LastPriceInsights() search.PriceInsights {
	return p.lastInsights
}

// Search queries Google Flights via SerpAPI and returns normalized flights.
func (p *Provider) Search(ctx context.Context, req types.SearchRequest) ([]types.Flight, error) {
	tripType := config.SerpAPITypeOneWay
	if req.IsRoundTrip() {
		tripType = config.SerpAPITypeRoundTrip
	}

	params := map[string]string{
		config.SerpAPIParamEngine:    config.SerpAPIEngineFlights,
		config.SerpAPIParamAPIKey:    p.cfg.APIKey,
		config.SerpAPIParamDeparture: req.Origin,
		config.SerpAPIParamArrival:   req.Destination,
		config.SerpAPIParamDate:      req.DepartureDate.Format("2006-01-02"),
		config.SerpAPIParamType:      tripType,
		config.SerpAPIParamAdults:    strconv.Itoa(req.Passengers),
		config.SerpAPIParamCurrency:  "USD",
		config.SerpAPIParamClass:     mapCabinClassToSerpAPI(req.CabinClass),
		config.SerpAPIParamLanguage:  "en",
		config.SerpAPIParamCountry:   "us",
	}

	if req.MaxStops == 0 {
		params[config.SerpAPIParamStops] = "0"
	}

	url, err := httpclient.BuildURL(p.cfg.BaseURL, config.SerpAPISearchPath, params)
	if err != nil {
		return nil, fmt.Errorf("building URL: %w", err)
	}

	log.Printf("[serpapi] GET %s→%s %s", req.Origin, req.Destination, req.DepartureDate.Format("2006-01-02"))

	var resp Response
	if err := p.http.GetJSON(ctx, url, nil, &resp); err != nil {
		return nil, fmt.Errorf("serpapi request: %w", err)
	}

	flights, err := ParseResponse(&resp)
	if err != nil {
		return nil, fmt.Errorf("parsing response: %w", err)
	}

	p.lastInsights = ParsePriceInsights(&resp)

	log.Printf("[serpapi] %s→%s: %d flights (%d best, %d other)",
		req.Origin, req.Destination, len(flights),
		len(resp.BestFlights), len(resp.OtherFlights))

	return flights, nil
}

// MultiCityRequest defines the input for a multi-city search.
type MultiCityRequest struct {
	Origin      string // IATA code for leg 1 origin
	Stopover    string // IATA code for stopover city
	Destination string // IATA code for leg 2 destination
	Leg1Date    string // YYYY-MM-DD
	Leg2Date    string // YYYY-MM-DD
	Passengers  int
	CabinClass  types.CabinClass
	TopN        int // how many leg1 options to expand in step 2 (default 3)
}

// searchMultiCityInternal performs a two-step Google Flights multi-city search.
// Step 1 fetches leg 1 options with departure tokens.
// Step 2 uses each token to fetch leg 2 options with combined pricing.
func (p *Provider) searchMultiCityInternal(ctx context.Context, req MultiCityRequest) ([]MultiCityResult, error) {
	if req.TopN <= 0 {
		req.TopN = 3
	}

	multiCityJSON, err := buildMultiCityJSON(req)
	if err != nil {
		return nil, fmt.Errorf("building multi_city_json: %w", err)
	}

	params := map[string]string{
		config.SerpAPIParamEngine:        config.SerpAPIEngineFlights,
		config.SerpAPIParamAPIKey:        p.cfg.APIKey,
		config.SerpAPIParamType:          config.SerpAPITypeMultiCity,
		config.SerpAPIParamMultiCityJSON: multiCityJSON,
		config.SerpAPIParamAdults:        strconv.Itoa(req.Passengers),
		config.SerpAPIParamCurrency:      "USD",
		config.SerpAPIParamClass:         mapCabinClassToSerpAPI(req.CabinClass),
		config.SerpAPIParamLanguage:      "en",
		config.SerpAPIParamCountry:       "us",
	}

	// Step 1: get leg 1 options with departure tokens.
	url, err := httpclient.BuildURL(p.cfg.BaseURL, config.SerpAPISearchPath, params)
	if err != nil {
		return nil, fmt.Errorf("building step1 URL: %w", err)
	}

	log.Printf("[serpapi] multi-city step1: %s→%s→%s", req.Origin, req.Stopover, req.Destination)

	var step1Resp Response
	if err := p.http.GetJSON(ctx, url, nil, &step1Resp); err != nil {
		return nil, fmt.Errorf("step1 request: %w", err)
	}

	var leg1Options []FlightGroup
	leg1Options = append(leg1Options, step1Resp.BestFlights...)
	leg1Options = append(leg1Options, step1Resp.OtherFlights...)
	log.Printf("[serpapi] step1: %d leg1 options", len(leg1Options))

	if len(leg1Options) == 0 {
		return nil, nil
	}

	// Sort by price and take top N for step 2.
	sort.Slice(leg1Options, func(i, j int) bool {
		return leg1Options[i].Price < leg1Options[j].Price
	})
	if len(leg1Options) > req.TopN {
		leg1Options = leg1Options[:req.TopN]
	}

	// Step 2: for each leg1 option, fetch leg 2 using its departure token.
	var results []MultiCityResult
	for _, leg1 := range leg1Options {
		if leg1.DepartureToken == "" {
			continue
		}

		params[config.SerpAPIParamDepartureToken] = leg1.DepartureToken
		step2URL, err := httpclient.BuildURL(p.cfg.BaseURL, config.SerpAPISearchPath, params)
		if err != nil {
			log.Printf("[serpapi] step2 URL error: %v", err)
			continue
		}

		var step2Resp Response
		if err := p.http.GetJSON(ctx, step2URL, nil, &step2Resp); err != nil {
			log.Printf("[serpapi] step2 request error: %v", err)
			continue
		}

		var leg2Options []FlightGroup
		leg2Options = append(leg2Options, step2Resp.BestFlights...)
		leg2Options = append(leg2Options, step2Resp.OtherFlights...)
		log.Printf("[serpapi] step2 for leg1 $%d: %d leg2 options", leg1.Price, len(leg2Options))

		for _, leg2 := range leg2Options {
			results = append(results, MultiCityResult{
				Leg1:  leg1,
				Leg2:  leg2,
				Price: leg2.Price, // step 2 returns combined price
			})
		}
	}
	delete(params, config.SerpAPIParamDepartureToken)

	return results, nil
}

// buildMultiCityJSON produces the JSON array for the multi_city_json parameter.
func buildMultiCityJSON(req MultiCityRequest) (string, error) {
	legs := []map[string]string{
		{"departure_id": req.Origin, "arrival_id": req.Stopover, "date": req.Leg1Date},
		{"departure_id": req.Stopover, "arrival_id": req.Destination, "date": req.Leg2Date},
	}
	b, err := json.Marshal(legs)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// SearchMultiCity implements provider.MultiCitySearcher by delegating
// to the internal two-step search and converting the results.
func (p *Provider) SearchMultiCity(ctx context.Context, req provider.MultiCityRequest) ([]provider.MultiCityResult, error) {
	internal := MultiCityRequest{
		Origin:      req.Origin,
		Stopover:    req.Stopover,
		Destination: req.Destination,
		Leg1Date:    req.Leg1Date,
		Leg2Date:    req.Leg2Date,
		Passengers:  req.Passengers,
		CabinClass:  req.CabinClass,
		TopN:        req.TopN,
	}

	results, err := p.searchMultiCityInternal(ctx, internal)
	if err != nil {
		return nil, err
	}

	var out []provider.MultiCityResult
	for _, r := range results {
		leg1, err := parseFlightGroup(r.Leg1)
		if err != nil {
			log.Printf("[serpapi] skipping multi-city result (leg1 parse): %v", err)
			continue
		}
		leg2, err := parseFlightGroup(r.Leg2)
		if err != nil {
			log.Printf("[serpapi] skipping multi-city result (leg2 parse): %v", err)
			continue
		}

		totalPrice := types.Money{Amount: float64(r.Price), Currency: "USD"}
		leg1.Price = totalPrice
		leg2.Price = types.Money{Amount: 0, Currency: "USD"}

		out = append(out, provider.MultiCityResult{
			Leg1:  leg1,
			Leg2:  leg2,
			Price: totalPrice,
		})
	}

	return out, nil
}

// Compile-time check that Provider implements provider.MultiCitySearcher.
var _ provider.MultiCitySearcher = (*Provider)(nil)

func mapCabinClassToSerpAPI(c types.CabinClass) string {
	switch c {
	case types.CabinBusiness:
		return config.SerpAPIClassBusiness
	case types.CabinFirst:
		return config.SerpAPIClassFirst
	case types.CabinPremiumEconomy:
		return config.SerpAPIClassPremiumEconomy
	default:
		return config.SerpAPIClassEconomy
	}
}
