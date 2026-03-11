// multicity.go is the orchestrator for multi-city halt searches.
//
// # Pipeline Overview
//
// The search proceeds in 5 stages:
//
//	EXPAND → FETCH → FILTER → COMBINE → RANK
//
// Stage 1 — EXPAND:
//
//	For each stopover city in the route's stopover list (see stopovers.go),
//	generate two search requests: origin→stopover and stopover→destination.
//	All stopover cities are searched in parallel.
//
// Stage 2 — FETCH:
//
//	Hit the SerpAPI Google Flights endpoint for each leg. One-way searches
//	return flights for the specified departure date (with optional flex days).
//	Results are cached to minimize API calls.
//
// Stage 3 — FILTER (three passes):
//
//	3a. Blocked airlines/hubs: Remove flights using airlines or routing
//	    through airports affected by the Middle East airspace closures.
//	    See search/filter.go for the full block list.
//
//	3b. Date window: Post-filter leg1 to the user's departure window
//	    (target ± flex days). Leg2 is NOT date-filtered here — the combiner
//	    handles it by checking that leg2 departs within
//	    [leg1_arrival + minStay, leg1_arrival + maxStay].
//
//	3c. Layover quality (in combiner): Reject legs with airport layovers
//	    > 6 hours or < 1 hour. See combiner.go.
//
// Stage 4 — COMBINE:
//
//	For each (leg1, leg2) pair, check that the gap between leg1 arrival
//	and leg2 departure is a valid city stopover (2-4 days, not a 16-hour
//	airport layover). Build Itinerary structs from valid pairs.
//	See combiner.go for detailed logic.
//
// Stage 5 — RANK:
//
//	Send the top candidates (sorted by price) to the LLM for intelligent
//	scoring. The LLM evaluates: cost (35%), airline consistency (20%),
//	layover quality (15%), flight duration (15%), stopover city (10%),
//	schedule convenience (5%). See ranker.go for the prompt.
//
// # Date Handling
//
// SerpAPI Google Flights accepts a departure date parameter. Flex-date
// searches issue one request per date in the [target - flexDays, target +
// flexDays] window. The combiner then only pairs leg1+leg2 where the gap
// falls within the stopover city's min/max stay range.
//
// Leg2 dates are implicitly constrained by leg1 arrival + stopover duration.
//
// # Current Limitations
//
// TODO(iterate): Search each stopover city independently. A smarter
// approach: use the LLM to first pick 3-4 most promising stopover cities.
//
// TODO(iterate): Add support for more than one intermediate stop
// (e.g. DEL → BKK → NRT → YYZ for three-city itineraries).
//
// TODO(iterate): Deduplicate itineraries that differ only in price
// (same flights, different booking providers).
package multicity

import (
	"context"
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"booker/provider"
	"booker/search"
	"booker/types"
)

// DateLayout is the format for parsing the DepartureDate string.
const DateLayout = "2006-01-02"

// SearchParams defines the user's intent for a multi-city halt search.
type SearchParams struct {
	Origin      string // IATA code, e.g. "DEL"
	Destination string // IATA code, e.g. "YYZ"

	// DepartureDate is the target date in YYYY-MM-DD format.
	// The search is flexible: we look at ±FlexDays around this date.
	DepartureDate string
	Passengers    int
	CabinClass    types.CabinClass

	// Leg2Date is the target departure date for the second leg (YYYY-MM-DD).
	// This is the date the traveler leaves the stopover city.
	Leg2Date string

	// FlexDays is how many days before/after the target date we accept
	// for leg1 departure. E.g. FlexDays=3 means we accept Mar 23-29
	// for a target of Mar 26.
	FlexDays int

	// MaxLayoversLeg1 limits the number of connections on leg 1.
	// 0 = direct only, 1 = one stop max, -1 = no limit (default).
	MaxLayoversLeg1 int
	// MaxLayoversLeg2 limits the number of connections on leg 2.
	MaxLayoversLeg2 int

	// Stopovers overrides the default stopover cities when set.
	Stopovers []StopoverCity

	// MaxResults caps the number of itineraries returned after ranking.
	MaxResults int

	// PreferredAlliance filters flights to those from the given alliance
	// ("Star Alliance", "OneWorld", "SkyTeam"). Empty means no filter.
	PreferredAlliance string

	// MaxPrice filters combined itineraries whose total price exceeds this
	// amount (USD). 0 means no limit.
	MaxPrice float64

	// DepartureAfter filters flights departing before this time-of-day ("HH:MM").
	// Empty means no constraint.
	DepartureAfter string
	// DepartureBefore filters flights departing after this time-of-day ("HH:MM").
	// Empty means no constraint.
	DepartureBefore string

	// ArrivalAfter filters flights arriving before this time-of-day ("HH:MM").
	// Empty means no constraint.
	ArrivalAfter string
	// ArrivalBefore filters flights arriving after this time-of-day ("HH:MM").
	// Empty means no constraint.
	ArrivalBefore string

	// MaxDuration filters flights whose TotalDuration exceeds this.
	// 0 means no limit.
	MaxDuration time.Duration

	// AvoidAirlines is a comma-separated list of IATA codes to exclude.
	// Empty means no filter.
	AvoidAirlines string

	// PreferredAirlines is a comma-separated list of IATA codes to keep.
	// Empty means no filter.
	PreferredAirlines string

	// Context is the user's natural language preferences (e.g. "I hate long
	// layovers"). Forwarded to the LLM ranker so it can factor user intent
	// into scoring. Empty means no additional context.
	Context string

	// MinStopover overrides the default minimum stopover duration.
	// 0 means use the stopover city's default (typically 48h).
	MinStopover time.Duration
	// MaxStopover overrides the default maximum stopover duration.
	// 0 means use the stopover city's default (typically 144h).
	MaxStopover time.Duration
}

// Searcher orchestrates the multi-city halt search pipeline.
type Searcher struct {
	registry *provider.Registry
	ranker   *Ranker
}

// NewSearcher creates a multi-city searcher sharing the provided Ranker.
// Sharing a single Ranker instance lets callers update weights dynamically
// (e.g. when a chat user switches ranking profile mid-session).
func NewSearcher(registry *provider.Registry, ranker *Ranker) *Searcher {
	return &Searcher{
		registry: registry,
		ranker:   ranker,
	}
}

// Search executes the full pipeline: expand → fetch → filter → combine → rank.
func (s *Searcher) Search(ctx context.Context, params SearchParams) ([]search.Itinerary, error) {
	// Parse the target departure dates.
	targetDate, err := time.Parse(DateLayout, params.DepartureDate)
	if err != nil {
		return nil, fmt.Errorf("parsing departure date %q: %w", params.DepartureDate, err)
	}

	var leg2Date time.Time
	if params.Leg2Date != "" {
		leg2Date, err = time.Parse(DateLayout, params.Leg2Date)
		if err != nil {
			return nil, fmt.Errorf("parsing leg2 date %q: %w", params.Leg2Date, err)
		}
	}

	flexDays := params.FlexDays
	if flexDays <= 0 {
		flexDays = 3
	}
	dateEarliest := targetDate.AddDate(0, 0, -flexDays)
	dateLatest := targetDate.AddDate(0, 0, flexDays).Add(23*time.Hour + 59*time.Minute)

	log.Printf("[multicity] leg1 date window: %s to %s",
		dateEarliest.Format(DateLayout), dateLatest.Format(DateLayout))
	if !leg2Date.IsZero() {
		log.Printf("[multicity] leg2 date: %s", leg2Date.Format(DateLayout))
	}

	stopovers := params.Stopovers
	if len(stopovers) == 0 {
		stopovers = StopoversForRoute(params.Origin, params.Destination)
	}
	if len(stopovers) == 0 {
		return nil, fmt.Errorf("no stopover cities defined for %s → %s", params.Origin, params.Destination)
	}

	log.Printf("[multicity] searching %s → %s via %d stopover cities",
		params.Origin, params.Destination, len(stopovers))

	// ---------------------------------------------------------------
	// STAGE 1+2: EXPAND & FETCH
	// For each stopover city, search both legs in parallel.
	// Each goroutine makes 2 API calls (leg1 + leg2).
	//
	// Simultaneously, providers that support native multi-city search
	// (combined pricing) are queried in parallel.
	// ---------------------------------------------------------------
	type legPair struct {
		stopover StopoverCity
		leg1     []types.Flight
		leg2     []types.Flight
	}

	var (
		mu    sync.Mutex
		pairs []legPair
		wg    sync.WaitGroup
	)

	// Collect providers that support multi-city search.
	var mcProviders []provider.MultiCitySearcher
	for _, p := range s.registry.All() {
		if mc, ok := p.(provider.MultiCitySearcher); ok {
			mcProviders = append(mcProviders, mc)
		}
	}

	// Multi-city results are collected in parallel with one-way results.
	var (
		mcMu          sync.Mutex
		mcItineraries []search.Itinerary
	)

	for _, stop := range stopovers {
		// One-way fetch for this stopover.
		wg.Add(1)
		go func(stop StopoverCity) {
			defer wg.Done()

			log.Printf("[multicity]   fetching %s → %s → %s",
				params.Origin, stop.City, params.Destination)

			leg1Req := types.SearchRequest{
				Origin:        params.Origin,
				Destination:   stop.Airport,
				DepartureDate: targetDate,
				Passengers:    params.Passengers,
				CabinClass:    params.CabinClass,
			}
			leg1Results := s.fetchFromAllProviders(ctx, leg1Req)

			leg2Req := types.SearchRequest{
				Origin:        stop.Airport,
				Destination:   params.Destination,
				DepartureDate: leg2Date,
				Passengers:    params.Passengers,
				CabinClass:    params.CabinClass,
			}
			leg2Results := s.fetchFromAllProviders(ctx, leg2Req)

			mu.Lock()
			pairs = append(pairs, legPair{
				stopover: stop,
				leg1:     leg1Results,
				leg2:     leg2Results,
			})
			mu.Unlock()

			log.Printf("[multicity]   %s: %d leg1, %d leg2 (raw)",
				stop.City, len(leg1Results), len(leg2Results))
		}(stop)

		// Multi-city fetch for this stopover (runs in parallel).
		for _, mc := range mcProviders {
			wg.Add(1)
			go func(mc provider.MultiCitySearcher, stop StopoverCity) {
				defer wg.Done()

				mcReq := provider.MultiCityRequest{
					Origin:      params.Origin,
					Stopover:    stop.Airport,
					Destination: params.Destination,
					Leg1Date:    params.DepartureDate,
					Leg2Date:    params.Leg2Date,
					Passengers:  params.Passengers,
					CabinClass:  params.CabinClass,
					TopN:        3,
				}

				results, err := mc.SearchMultiCity(ctx, mcReq)
				if err != nil {
					log.Printf("[multicity] multi-city %s error: %v", stop.City, err)
					return
				}

				for _, r := range results {
					itin := buildMultiCityItinerary(r, stop)
					mcMu.Lock()
					mcItineraries = append(mcItineraries, itin)
					mcMu.Unlock()
				}

				log.Printf("[multicity] multi-city %s: %d itineraries", stop.City, len(results))
			}(mc, stop)
		}
	}
	wg.Wait()

	// ---------------------------------------------------------------
	// STAGE 3: FILTER
	//
	// 3a. Remove blocked airlines and hubs (Middle East closures).
	// 3b. Date-filter leg1 to the user's departure window.
	//     Leg2 is NOT date-filtered — the combiner constrains it
	//     via stopover duration (leg1_arrival + 2-4 days).
	// ---------------------------------------------------------------
	for i := range pairs {
		// Apply all per-flight filters to both legs.
		// applyBoth applies a filter to both legs and returns the drop counts.
		type filterFunc func([]types.Flight) []types.Flight
		applyBoth := func(f filterFunc) (int, int) {
			b1, b2 := len(pairs[i].leg1), len(pairs[i].leg2)
			pairs[i].leg1 = f(pairs[i].leg1)
			pairs[i].leg2 = f(pairs[i].leg2)
			return b1 - len(pairs[i].leg1), b2 - len(pairs[i].leg2)
		}

		blocked1, blocked2 := applyBoth(search.FilterFlights)
		zero1, zero2 := applyBoth(search.FilterZeroPrices)

		// Max layovers: different limits per leg.
		b1 := len(pairs[i].leg1)
		pairs[i].leg1 = search.FilterByMaxStops(pairs[i].leg1, params.MaxLayoversLeg1)
		stops1 := b1 - len(pairs[i].leg1)
		b2 := len(pairs[i].leg2)
		pairs[i].leg2 = search.FilterByMaxStops(pairs[i].leg2, params.MaxLayoversLeg2)
		stops2 := b2 - len(pairs[i].leg2)
		alliance1, alliance2 := applyBoth(func(f []types.Flight) []types.Flight {
			return search.FilterByAlliance(f, params.PreferredAlliance)
		})
		depTime1, depTime2 := applyBoth(func(f []types.Flight) []types.Flight {
			return search.FilterByDepartureTime(f, params.DepartureAfter, params.DepartureBefore)
		})
		applyBoth(func(f []types.Flight) []types.Flight {
			return search.FilterByArrivalTime(f, params.ArrivalAfter, params.ArrivalBefore)
		})
		applyBoth(func(f []types.Flight) []types.Flight {
			return search.FilterByMaxDuration(f, params.MaxDuration)
		})
		applyBoth(func(f []types.Flight) []types.Flight {
			return search.FilterByAvoidAirlines(f, params.AvoidAirlines)
		})
		applyBoth(func(f []types.Flight) []types.Flight {
			return search.FilterByPreferredAirlines(f, params.PreferredAirlines)
		})

		// Date window (leg1 only).
		beforeDate := len(pairs[i].leg1)
		pairs[i].leg1 = search.FilterByDateRange(pairs[i].leg1, dateEarliest, dateLatest)
		dateDrop := beforeDate - len(pairs[i].leg1)

		log.Printf("[multicity]   %s after filter: %d leg1 (-%d blocked, -%d $0, -%d stops, -%d alliance, -%d deptime, -%d date), %d leg2 (-%d blocked, -%d $0, -%d stops, -%d alliance, -%d deptime)",
			pairs[i].stopover.City,
			len(pairs[i].leg1), blocked1, zero1, stops1, alliance1, depTime1, dateDrop,
			len(pairs[i].leg2), blocked2, zero2, stops2, alliance2, depTime2)
	}

	// ---------------------------------------------------------------
	// STAGE 4: COMBINE
	// Pair leg1 + leg2 into complete itineraries.
	// Only combinations satisfying stopover duration and layover
	// constraints survive. See combiner.go for details.
	// ---------------------------------------------------------------
	var allItineraries []search.Itinerary
	for _, pair := range pairs {
		combined := CombineLegs(pair.leg1, pair.leg2, CombineParams{
			Stopover:        pair.stopover,
			MinStay:         params.MinStopover,
			MaxStay:         params.MaxStopover,
			DepartureAfter:  params.DepartureAfter,
			DepartureBefore: params.DepartureBefore,
		})
		allItineraries = append(allItineraries, combined...)
		if len(combined) > 0 {
			log.Printf("[multicity]   %s: %d valid itineraries (cheapest $%.0f)",
				pair.stopover.City, len(combined), combined[0].TotalPrice.Amount)
		} else {
			log.Printf("[multicity]   %s: 0 valid itineraries", pair.stopover.City)
		}
	}

	// ---------------------------------------------------------------
	// STAGE 4b: MERGE MULTI-CITY RESULTS
	// Filter and merge itineraries from native multi-city search.
	// These have combined pricing (often cheaper than two one-ways).
	// ---------------------------------------------------------------
	var filteredMC int
	for _, itin := range mcItineraries {
		leg1 := itin.Legs[0].Flight
		leg2 := itin.Legs[1].Flight

		// Apply the same per-flight filters as one-way results.
		if !passesAllFilters(leg1, params) || !passesAllFilters(leg2, params) {
			continue
		}
		// Leg-specific layover checks.
		if params.MaxLayoversLeg1 >= 0 && leg1.Stops() > params.MaxLayoversLeg1 {
			continue
		}
		if params.MaxLayoversLeg2 >= 0 && leg2.Stops() > params.MaxLayoversLeg2 {
			continue
		}
		// Itinerary-level max price check.
		if params.MaxPrice > 0 && itin.TotalPrice.Amount > params.MaxPrice {
			continue
		}

		allItineraries = append(allItineraries, itin)
		filteredMC++
	}

	log.Printf("[multicity] multi-city itineraries after filter: %d/%d", filteredMC, len(mcItineraries))

	// Filter combined itineraries by max total price.
	if params.MaxPrice > 0 {
		filtered := allItineraries[:0]
		for _, itin := range allItineraries {
			if itin.TotalPrice.Amount <= params.MaxPrice {
				filtered = append(filtered, itin)
			}
		}
		allItineraries = filtered
	}

	log.Printf("[multicity] total candidate itineraries: %d", len(allItineraries))

	if len(allItineraries) == 0 {
		return nil, nil
	}

	// Pre-sort by total price so the cheapest go to the LLM.
	sort.Slice(allItineraries, func(i, j int) bool {
		return allItineraries[i].TotalPrice.Amount < allItineraries[j].TotalPrice.Amount
	})

	// Deduplicate itineraries with identical flights, keeping cheapest.
	before := len(allItineraries)
	allItineraries = deduplicateItineraries(allItineraries)
	if dupes := before - len(allItineraries); dupes > 0 {
		log.Printf("[multicity] deduplicated %d itineraries", dupes)
	}

	// ---------------------------------------------------------------
	// STAGE 5: RANK
	// Send top candidates to the LLM for intelligent scoring.
	// Falls back to price-sorted results if LLM fails.
	// ---------------------------------------------------------------
	s.ranker.UserContext = params.Context
	ranked, err := s.ranker.Rank(ctx, allItineraries)
	if err != nil {
		log.Printf("[multicity] LLM ranking failed, falling back to price sort: %v", err)
		if params.MaxResults > 0 && len(allItineraries) > params.MaxResults {
			allItineraries = allItineraries[:params.MaxResults]
		}
		return allItineraries, nil
	}

	if params.MaxResults > 0 && len(ranked) > params.MaxResults {
		ranked = ranked[:params.MaxResults]
	}

	return ranked, nil
}

// passesAllFilters returns true if a single flight passes all generic
// per-flight filters (blocked airlines, zero price, alliance, departure/arrival
// time, max duration, avoid/preferred airlines). Leg-specific checks like
// max layovers and itinerary-level checks like max price are NOT included.
func passesAllFilters(f types.Flight, params SearchParams) bool {
	return search.FlightPassesBlocked(f) &&
		f.Price.Amount > 0 &&
		search.FlightPassesAlliance(f, params.PreferredAlliance) &&
		search.FlightPassesDepartureTime(f, params.DepartureAfter, params.DepartureBefore) &&
		search.FlightPassesArrivalTime(f, params.ArrivalAfter, params.ArrivalBefore) &&
		search.FlightPassesMaxDuration(f, params.MaxDuration) &&
		search.FlightPassesAvoidAirlines(f, params.AvoidAirlines) &&
		search.FlightPassesPreferredAirlines(f, params.PreferredAirlines)
}

// buildMultiCityItinerary converts a provider.MultiCityResult into a
// search.Itinerary. The flights are already parsed; we just need to
// compute stopover and trip durations.
func buildMultiCityItinerary(r provider.MultiCityResult, stop StopoverCity) search.Itinerary {
	var stopoverDur time.Duration
	if len(r.Leg1.Outbound) > 0 && len(r.Leg2.Outbound) > 0 {
		leg1Arr := r.Leg1.Outbound[len(r.Leg1.Outbound)-1].ArrivalTime
		leg2Dep := r.Leg2.Outbound[0].DepartureTime
		stopoverDur = leg2Dep.Sub(leg1Arr)
	}

	totalTravel := r.Leg1.TotalDuration + r.Leg2.TotalDuration
	var totalTrip time.Duration
	if len(r.Leg1.Outbound) > 0 && len(r.Leg2.Outbound) > 0 {
		firstDep := r.Leg1.Outbound[0].DepartureTime
		lastArr := r.Leg2.Outbound[len(r.Leg2.Outbound)-1].ArrivalTime
		totalTrip = lastArr.Sub(firstDep)
	}

	return search.Itinerary{
		Legs: []search.Leg{
			{
				Flight: r.Leg1,
				Stopover: &search.Stopover{
					City:     stop.City,
					Airport:  stop.Airport,
					Duration: stopoverDur,
				},
			},
			{Flight: r.Leg2},
		},
		TotalPrice:  r.Price,
		TotalTravel: totalTravel,
		TotalTrip:   totalTrip,
	}
}

func (s *Searcher) fetchFromAllProviders(ctx context.Context, req types.SearchRequest) []types.Flight {
	var allFlights []types.Flight
	for _, p := range s.registry.All() {
		results, err := p.Search(ctx, req)
		if err != nil {
			log.Printf("[multicity] provider %s error: %v", p.Name(), err)
			continue
		}
		allFlights = append(allFlights, results...)
	}
	return allFlights
}

// deduplicateItineraries removes itineraries with identical flights (same flight
// numbers and departure times across all legs), keeping the cheapest.
func deduplicateItineraries(itins []search.Itinerary) []search.Itinerary {
	type entry struct {
		idx   int
		price float64
	}
	best := make(map[string]entry)
	for i, itin := range itins {
		key := itineraryKey(itin)
		if prev, ok := best[key]; !ok || itin.TotalPrice.Amount < prev.price {
			best[key] = entry{idx: i, price: itin.TotalPrice.Amount}
		}
	}
	result := make([]search.Itinerary, 0, len(best))
	for _, e := range best {
		result = append(result, itins[e.idx])
	}
	// Re-sort by price since map iteration is unordered.
	sort.Slice(result, func(i, j int) bool {
		return result[i].TotalPrice.Amount < result[j].TotalPrice.Amount
	})
	return result
}

// itineraryKey builds a dedup key from each leg's first segment flight number + departure.
func itineraryKey(itin search.Itinerary) string {
	var key string
	for _, leg := range itin.Legs {
		if len(leg.Flight.Outbound) > 0 {
			seg := leg.Flight.Outbound[0]
			key += seg.FlightNumber + seg.DepartureTime.Format("2006-01-02T15:04") + "|"
		}
	}
	return key
}
