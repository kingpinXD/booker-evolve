package kiwi

// response.go contains the typed structs that map to the Kiwi RapidAPI
// JSON response. Only fields we consume are included.

// Response is the top-level API response.
type Response struct {
	Metadata    Metadata    `json:"metadata"`
	Itineraries []Itinerary `json:"itineraries"`
}

// Metadata contains carrier lookup info and result counts.
type Metadata struct {
	Carriers        []Carrier `json:"carriers"`
	ItinerariesCount int      `json:"itinerariesCount"`
	HasMorePending  bool      `json:"hasMorePending"`
}

// Carrier is an airline entry from the metadata lookup table.
type Carrier struct {
	ID   string `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

// Itinerary is a single flight option (one-way or round-trip).
// One-way results use "sector", round-trip uses "outbound"/"inbound".
type Itinerary struct {
	ID             string          `json:"id"`
	Typename       string          `json:"__typename"` // "ItineraryOneWay" or "ItineraryReturn"
	Price          Price           `json:"price"`
	PriceEur       Price           `json:"priceEur"`
	Provider       ProviderInfo    `json:"provider"`
	BagsInfo       BagsInfo        `json:"bagsInfo"`
	BookingOptions BookingOptions  `json:"bookingOptions"`
	Sector         *Sector         `json:"sector"`   // one-way flights
	Outbound       *Sector         `json:"outbound"` // round-trip outbound
	Inbound        *Sector         `json:"inbound"`  // round-trip inbound
	Stopover       *StopoverInfo   `json:"stopover"`
	LastAvailable  *LastAvailable  `json:"lastAvailable"`
	TravelHack     TravelHack      `json:"travelHack"`
}

// OutboundSector returns the outbound sector regardless of whether this
// is a one-way (uses Sector) or round-trip (uses Outbound) itinerary.
func (it *Itinerary) OutboundSector() *Sector {
	if it.Sector != nil {
		return it.Sector
	}
	return it.Outbound
}

// Price holds a monetary amount as a string from the API.
type Price struct {
	Amount              string `json:"amount"`
	PriceBeforeDiscount string `json:"priceBeforeDiscount"`
}

// ProviderInfo identifies who is selling the ticket.
type ProviderInfo struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// BagsInfo describes included and purchasable baggage.
type BagsInfo struct {
	IncludedCheckedBags int          `json:"includedCheckedBags"`
	IncludedHandBags    int          `json:"includedHandBags"`
	CheckedBagTiers     []BagTier    `json:"checkedBagTiers"`
	HandBagTiers        []BagTier    `json:"handBagTiers"`
}

// BagTier is a purchasable baggage option.
type BagTier struct {
	TierPrice Price  `json:"tierPrice"`
	Bags      []Bag  `json:"bags"`
}

// Bag describes a single bag's weight.
type Bag struct {
	Weight BagWeight `json:"weight"`
}

// BagWeight holds the weight value in kg.
type BagWeight struct {
	Value float64 `json:"value"`
}

// BookingOptions contains the booking URL edges.
type BookingOptions struct {
	Edges []BookingEdge `json:"edges"`
}

// BookingEdge wraps a single booking option node.
type BookingEdge struct {
	Node BookingNode `json:"node"`
}

// BookingNode holds the booking URL and price for one option.
type BookingNode struct {
	BookingURL        string       `json:"bookingUrl"`
	ItineraryProvider ProviderInfo `json:"itineraryProvider"`
	Price             Price        `json:"price"`
}

// Sector represents one direction of travel (outbound or inbound),
// containing one or more segments with optional layovers.
type Sector struct {
	ID              string           `json:"id"`
	SectorSegments  []SectorSegment  `json:"sectorSegments"`
	Duration        int              `json:"duration"` // total seconds
}

// SectorSegment pairs a flight segment with its following layover (if any).
type SectorSegment struct {
	Segment Segment  `json:"segment"`
	Layover *Layover `json:"layover"`
}

// Segment is a single non-stop flight.
type Segment struct {
	ID               string       `json:"id"`
	Source           StopPoint    `json:"source"`
	Destination      StopPoint    `json:"destination"`
	Duration         int          `json:"duration"` // seconds
	Type             string       `json:"type"`     // "FLIGHT"
	Code             string       `json:"code"`     // flight number portion, e.g. "5967"
	Carrier          CarrierRef   `json:"carrier"`
	OperatingCarrier CarrierRef   `json:"operatingCarrier"`
	CabinClass       string       `json:"cabinClass"` // "ECONOMY", "BUSINESS", etc.
}

// StopPoint is an origin or destination with time and station info.
type StopPoint struct {
	LocalTime string  `json:"localTime"` // "2026-04-29T20:50:00"
	UTCTime   string  `json:"utcTime"`
	Station   Station `json:"station"`
}

// Station is an airport or other transport stop.
type Station struct {
	ID       string  `json:"id"`
	LegacyID string  `json:"legacyId"` // IATA code, e.g. "STN"
	Name     string  `json:"name"`     // "London Stansted"
	Code     string  `json:"code"`     // IATA code
	Type     string  `json:"type"`     // "AIRPORT"
	GPS      GPS     `json:"gps"`
	City     CityRef `json:"city"`
	Country  CountryRef `json:"country"`
}

// GPS coordinates.
type GPS struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

// CityRef is a reference to a city.
type CityRef struct {
	LegacyID string `json:"legacyId"` // e.g. "london_gb"
	Name     string `json:"name"`
	ID       string `json:"id"`
}

// CountryRef is a reference to a country.
type CountryRef struct {
	Code string `json:"code"` // ISO 3166, e.g. "GB"
	ID   string `json:"id"`
}

// CarrierRef identifies an airline within a segment.
type CarrierRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Code string `json:"code"` // IATA carrier code, e.g. "FR"
}

// Layover describes a connection between segments.
type Layover struct {
	Duration     int    `json:"duration"` // seconds
	IsBagRecheck bool   `json:"isBagRecheck"`
	TransferType string `json:"transferType"`
}

// StopoverInfo describes the stopover at the destination (for round trips).
type StopoverInfo struct {
	NightsCount int `json:"nightsCount"`
}

// LastAvailable indicates seat scarcity.
type LastAvailable struct {
	SeatsLeft int `json:"seatsLeft"`
}

// TravelHack flags for special routing tricks.
type TravelHack struct {
	IsTrueHiddenCity     bool `json:"isTrueHiddenCity"`
	IsVirtualInterlining bool `json:"isVirtualInterlining"`
	IsThrowawayTicket    bool `json:"isThrowawayTicket"`
}
