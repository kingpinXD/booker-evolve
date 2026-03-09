package kiwi

import (
	"fmt"
	"strconv"
	"time"

	"booker/config"
	"booker/types"
)

const (
	kiwiTimeLayout  = "2006-01-02T15:04:05"
	kiwiBookingBase = "https://www.kiwi.com"

	// DefaultCurrency matches the currency we request from the API.
	DefaultCurrency = "USD"
)

// ParseResponse converts a Kiwi API Response into normalized Flight results.
func ParseResponse(resp *Response) ([]types.Flight, error) {
	carriers := buildCarrierLookup(resp.Metadata.Carriers)
	flights := make([]types.Flight, 0, len(resp.Itineraries))

	for i := range resp.Itineraries {
		f, err := parseItinerary(&resp.Itineraries[i], carriers)
		if err != nil {
			return nil, fmt.Errorf("itinerary %d: %w", i, err)
		}
		flights = append(flights, f)
	}
	return flights, nil
}

func parseItinerary(itin *Itinerary, carriers map[string]string) (types.Flight, error) {
	price, err := strconv.ParseFloat(itin.Price.Amount, 64)
	if err != nil {
		return types.Flight{}, fmt.Errorf("parsing price %q: %w", itin.Price.Amount, err)
	}

	outSector := itin.OutboundSector()
	if outSector == nil {
		return types.Flight{}, fmt.Errorf("no outbound sector found")
	}

	outbound, err := parseSector(outSector, carriers)
	if err != nil {
		return types.Flight{}, fmt.Errorf("outbound: %w", err)
	}

	var ret []types.Segment
	if itin.Inbound != nil {
		ret, err = parseSector(itin.Inbound, carriers)
		if err != nil {
			return types.Flight{}, fmt.Errorf("inbound: %w", err)
		}
	}

	bookingURL := extractBookingURL(itin)

	var seatsLeft int
	if itin.LastAvailable != nil {
		seatsLeft = itin.LastAvailable.SeatsLeft
	}

	return types.Flight{
		Provider:      config.ProviderKiwi,
		Price:         types.Money{Amount: price, Currency: DefaultCurrency},
		Outbound:      outbound,
		Return:        ret,
		TotalDuration: time.Duration(outSector.Duration) * time.Second,
		BookingURL:    bookingURL,
		SeatsLeft:     seatsLeft,
		BagsIncluded: types.BagsIncluded{
			HandBags:    itin.BagsInfo.IncludedHandBags,
			CheckedBags: itin.BagsInfo.IncludedCheckedBags,
		},
	}, nil
}

func parseSector(sector *Sector, carriers map[string]string) ([]types.Segment, error) {
	segments := make([]types.Segment, 0, len(sector.SectorSegments))
	for i, ss := range sector.SectorSegments {
		seg, err := parseSegment(&ss.Segment, carriers)
		if err != nil {
			return nil, fmt.Errorf("segment %d: %w", i, err)
		}

		if ss.Layover != nil {
			seg.LayoverDuration = time.Duration(ss.Layover.Duration) * time.Second
		}
		segments = append(segments, seg)
	}
	return segments, nil
}

func parseSegment(seg *Segment, carriers map[string]string) (types.Segment, error) {
	depTime, err := time.Parse(kiwiTimeLayout, seg.Source.LocalTime)
	if err != nil {
		return types.Segment{}, fmt.Errorf("parsing departure time %q: %w", seg.Source.LocalTime, err)
	}
	arrTime, err := time.Parse(kiwiTimeLayout, seg.Destination.LocalTime)
	if err != nil {
		return types.Segment{}, fmt.Errorf("parsing arrival time %q: %w", seg.Destination.LocalTime, err)
	}

	airlineName := carriers[seg.Carrier.Code]
	if airlineName == "" {
		airlineName = seg.Carrier.Name
	}

	return types.Segment{
		Airline:          seg.Carrier.Code,
		AirlineName:      airlineName,
		FlightNumber:     seg.Carrier.Code + seg.Code,
		Origin:           seg.Source.Station.Code,
		OriginName:       seg.Source.Station.Name,
		OriginCity:       seg.Source.Station.City.Name,
		Destination:      seg.Destination.Station.Code,
		DestinationName:  seg.Destination.Station.Name,
		DestinationCity:  seg.Destination.Station.City.Name,
		DepartureTime:    depTime,
		ArrivalTime:      arrTime,
		Duration:         time.Duration(seg.Duration) * time.Second,
		CabinClass:       mapCabinClass(seg.CabinClass),
		OperatingCarrier: seg.OperatingCarrier.Code,
	}, nil
}

func extractBookingURL(itin *Itinerary) string {
	if len(itin.BookingOptions.Edges) > 0 {
		path := itin.BookingOptions.Edges[0].Node.BookingURL
		if path != "" {
			return kiwiBookingBase + path
		}
	}
	return ""
}

func buildCarrierLookup(carriers []Carrier) map[string]string {
	m := make(map[string]string, len(carriers))
	for _, c := range carriers {
		m[c.Code] = c.Name
	}
	return m
}

func mapCabinClass(kiwi string) types.CabinClass {
	switch kiwi {
	case config.KiwiCabinBusiness:
		return types.CabinBusiness
	case config.KiwiCabinFirst:
		return types.CabinFirst
	case config.KiwiCabinPremiumEconomy:
		return types.CabinPremiumEconomy
	default:
		return types.CabinEconomy
	}
}
