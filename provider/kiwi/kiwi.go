package kiwi

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"booker/config"
	"booker/httpclient"
	"booker/types"
)

// Provider implements provider.Provider for the Kiwi RapidAPI.
type Provider struct {
	cfg  config.ProviderConfig
	http *httpclient.Client
}

func New(cfg config.ProviderConfig, http *httpclient.Client) *Provider {
	return &Provider{cfg: cfg, http: http}
}

func (p *Provider) Name() config.ProviderName {
	return config.ProviderKiwi
}

// Search queries the Kiwi one-way endpoint and returns normalized flights.
// The origin/destination in req should be IATA codes. We also accept
// Kiwi-format IDs via OriginKiwiID/DestinationKiwiID if set in the request.
func (p *Provider) Search(ctx context.Context, req types.SearchRequest) ([]types.Flight, error) {
	// Build source/destination in Kiwi format.
	source := config.KiwiPrefixAirport + req.Origin
	dest := config.KiwiPrefixAirport + req.Destination

	// If KiwiID overrides are provided, use those instead.
	if req.OriginKiwiID != "" {
		source = req.OriginKiwiID
	}
	if req.DestinationKiwiID != "" {
		dest = req.DestinationKiwiID
	}

	params := map[string]string{
		config.KiwiParamSource:      source,
		config.KiwiParamDestination: dest,
		config.KiwiParamCurrency:    config.DefaultCurrency,
		config.KiwiParamLocale:      config.DefaultLocale,
		config.KiwiParamAdults:      strconv.Itoa(req.Passengers),
		config.KiwiParamChildren:    config.DefaultChildren,
		config.KiwiParamInfants:     config.DefaultInfants,
		config.KiwiParamCabinClass:  mapCabinClassToKiwi(req.CabinClass),
		config.KiwiParamSortBy:      sortBy(req.SortBy),
		config.KiwiParamSortOrder:   config.KiwiSortAscending,
		config.KiwiParamLimit:       config.DefaultResultLimit,
		config.KiwiParamTransport:   config.KiwiTransportFlight,
		config.KiwiParamProviders:   config.KiwiContentFresh + "," + config.KiwiContentKayak + "," + config.KiwiContentKiwi,
		config.KiwiParamHandbags:    config.DefaultHandbags,
		config.KiwiParamHoldbags:    config.DefaultHoldbags,
	}

	path := config.KiwiOneWayPath
	if req.IsRoundTrip() {
		path = config.KiwiRoundTripPath
	}

	url, err := httpclient.BuildURL(p.cfg.BaseURL, path, params)
	if err != nil {
		return nil, fmt.Errorf("building URL: %w", err)
	}

	headers := map[string]string{
		config.HeaderRapidAPIKey:  p.cfg.APIKey,
		config.HeaderRapidAPIHost: p.cfg.APIHost,
	}

	log.Printf("[kiwi] GET %s", url)

	var resp Response
	if err := p.http.GetJSON(ctx, url, headers, &resp); err != nil {
		return nil, fmt.Errorf("kiwi API request: %w", err)
	}

	log.Printf("[kiwi] %s→%s: %d itineraries returned", source, dest, len(resp.Itineraries))

	return ParseResponse(&resp)
}

func sortBy(hint string) string {
	switch hint {
	case config.KiwiSortByPrice, config.KiwiSortByQuality:
		return hint
	default:
		return config.KiwiSortByQuality
	}
}

func mapCabinClassToKiwi(c types.CabinClass) string {
	switch c {
	case types.CabinBusiness:
		return config.KiwiCabinBusiness
	case types.CabinFirst:
		return config.KiwiCabinFirst
	case types.CabinPremiumEconomy:
		return config.KiwiCabinPremiumEconomy
	default:
		return config.KiwiCabinEconomy
	}
}
