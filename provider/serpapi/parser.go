package serpapi

import (
	"fmt"
	"strings"
	"time"

	"booker/config"
	"booker/search"
	"booker/types"
)

const timeLayout = "2006-01-02 15:04"

// ParseResponse converts a SerpAPI response into normalized flights.
func ParseResponse(resp *Response) ([]types.Flight, error) {
	var flights []types.Flight
	for _, group := range append(resp.BestFlights, resp.OtherFlights...) {
		f, err := parseFlightGroup(group)
		if err != nil {
			continue
		}
		flights = append(flights, f)
	}
	return flights, nil
}

// ParsePriceInsights extracts price insights from a SerpAPI response.
func ParsePriceInsights(resp *Response) search.PriceInsights {
	var rng [2]float64
	if len(resp.PriceInsights.TypicalPriceRange) >= 2 {
		rng = [2]float64{float64(resp.PriceInsights.TypicalPriceRange[0]), float64(resp.PriceInsights.TypicalPriceRange[1])}
	}
	return search.PriceInsights{
		LowestPrice:       float64(resp.PriceInsights.LowestPrice),
		PriceLevel:        resp.PriceInsights.PriceLevel,
		TypicalPriceRange: rng,
	}
}

func parseFlightGroup(g FlightGroup) (types.Flight, error) {
	segments, err := parseSegments(g.Flights, g.Layovers)
	if err != nil {
		return types.Flight{}, err
	}

	return types.Flight{
		Provider:      config.ProviderSerpAPI,
		Price:         types.Money{Amount: float64(g.Price), Currency: "USD"},
		Outbound:      segments,
		TotalDuration: time.Duration(g.TotalDuration) * time.Minute,
		BookingURL:    g.BookingToken,
	}, nil
}

func parseSegments(segs []FlightSegment, layovers []Layover) ([]types.Segment, error) {
	out := make([]types.Segment, 0, len(segs))
	for i, seg := range segs {
		dep, err := time.Parse(timeLayout, seg.DepartureAirport.Time)
		if err != nil {
			return nil, fmt.Errorf("parsing departure time %q: %w", seg.DepartureAirport.Time, err)
		}
		arr, err := time.Parse(timeLayout, seg.ArrivalAirport.Time)
		if err != nil {
			return nil, fmt.Errorf("parsing arrival time %q: %w", seg.ArrivalAirport.Time, err)
		}

		airlineCode, _ := parseFlightNumber(seg.FlightNumber)

		var layoverDur time.Duration
		if i < len(layovers) {
			layoverDur = time.Duration(layovers[i].Duration) * time.Minute
		}

		out = append(out, types.Segment{
			Airline:         airlineCode,
			AirlineName:     seg.Airline,
			FlightNumber:    strings.ReplaceAll(seg.FlightNumber, " ", ""),
			Origin:          seg.DepartureAirport.ID,
			OriginName:      seg.DepartureAirport.Name,
			Destination:     seg.ArrivalAirport.ID,
			DestinationName: seg.ArrivalAirport.Name,
			DepartureTime:   dep,
			ArrivalTime:     arr,
			Duration:        time.Duration(seg.Duration) * time.Minute,
			CabinClass:      mapCabinClass(seg.TravelClass),
			LayoverDuration: layoverDur,
		})
	}
	return out, nil
}

// parseFlightNumber splits "TG 332" into ("TG", "332").
func parseFlightNumber(fn string) (airline, number string) {
	parts := strings.SplitN(fn, " ", 2)
	if len(parts) == 2 {
		return parts[0], parts[1]
	}
	return fn, ""
}

// ParseMultiCityResult converts a MultiCityResult into a search.Itinerary.
// The stopoverCity and stopoverAirport identify the intermediate city.
func ParseMultiCityResult(r MultiCityResult, stopoverCity, stopoverAirport string) (search.Itinerary, error) {
	leg1Flight, err := parseFlightGroup(r.Leg1)
	if err != nil {
		return search.Itinerary{}, fmt.Errorf("parsing leg1: %w", err)
	}

	leg2Flight, err := parseFlightGroup(r.Leg2)
	if err != nil {
		return search.Itinerary{}, fmt.Errorf("parsing leg2: %w", err)
	}

	// Use combined price from step 2 for both the itinerary total and leg1.
	// Individual leg prices are not meaningful since step 2 returns combined.
	totalPrice := types.Money{Amount: float64(r.Price), Currency: "USD"}
	leg1Flight.Price = totalPrice
	leg2Flight.Price = types.Money{Amount: 0, Currency: "USD"}

	// Compute stopover duration from leg1 arrival to leg2 departure.
	var stopoverDur time.Duration
	if len(leg1Flight.Outbound) > 0 && len(leg2Flight.Outbound) > 0 {
		leg1Arr := leg1Flight.Outbound[len(leg1Flight.Outbound)-1].ArrivalTime
		leg2Dep := leg2Flight.Outbound[0].DepartureTime
		stopoverDur = leg2Dep.Sub(leg1Arr)
	}

	totalTravel := leg1Flight.TotalDuration + leg2Flight.TotalDuration
	var totalTrip time.Duration
	if len(leg1Flight.Outbound) > 0 && len(leg2Flight.Outbound) > 0 {
		firstDep := leg1Flight.Outbound[0].DepartureTime
		lastArr := leg2Flight.Outbound[len(leg2Flight.Outbound)-1].ArrivalTime
		totalTrip = lastArr.Sub(firstDep)
	}

	return search.Itinerary{
		Legs: []search.Leg{
			{
				Flight: leg1Flight,
				Stopover: &search.Stopover{
					City:     stopoverCity,
					Airport:  stopoverAirport,
					Duration: stopoverDur,
				},
			},
			{Flight: leg2Flight},
		},
		TotalPrice:  totalPrice,
		TotalTravel: totalTravel,
		TotalTrip:   totalTrip,
	}, nil
}

func mapCabinClass(class string) types.CabinClass {
	switch strings.ToLower(class) {
	case "business":
		return types.CabinBusiness
	case "first":
		return types.CabinFirst
	case "premium economy":
		return types.CabinPremiumEconomy
	default:
		return types.CabinEconomy
	}
}
