package config

// Kiwi via RapidAPI
const (
	KiwiBaseURL       = "https://kiwi-com-cheap-flights.p.rapidapi.com"
	KiwiRapidAPIHost  = "kiwi-com-cheap-flights.p.rapidapi.com"
	KiwiOneWayPath    = "/one-way"
	KiwiRoundTripPath = "/round-trip"
)

// Kiwi query parameter keys.
const (
	KiwiParamSource      = "source"
	KiwiParamDestination = "destination"
	KiwiParamCurrency    = "currency"
	KiwiParamLocale      = "locale"
	KiwiParamAdults      = "adults"
	KiwiParamChildren    = "children"
	KiwiParamInfants     = "infants"
	KiwiParamCabinClass  = "cabinClass"
	KiwiParamSortBy      = "sortBy"
	KiwiParamSortOrder   = "sortOrder"
	KiwiParamLimit       = "limit"
	KiwiParamTransport   = "transportTypes"
	KiwiParamProviders   = "contentProviders"
	KiwiParamHandbags    = "handbags"
	KiwiParamHoldbags    = "holdbags"
)

// Kiwi source/destination prefix types.
const (
	KiwiPrefixAirport = "Airport:"
	KiwiPrefixCity    = "City:"
	KiwiPrefixCountry = "Country:"
)

// Kiwi sort values.
const (
	KiwiSortByPrice   = "PRICE"
	KiwiSortByQuality = "QUALITY"
	KiwiSortAscending = "ASCENDING"
)

// Kiwi cabin class values.
const (
	KiwiCabinEconomy        = "ECONOMY"
	KiwiCabinPremiumEconomy = "PREMIUM_ECONOMY"
	KiwiCabinBusiness       = "BUSINESS"
	KiwiCabinFirst          = "FIRST"
)

// Kiwi transport types.
const (
	KiwiTransportFlight = "FLIGHT"
)

// Kiwi content providers.
const (
	KiwiContentFresh = "FRESH"
	KiwiContentKayak = "KAYAK"
	KiwiContentKiwi  = "KIWI"
)

// SerpAPI (Google Flights)
const (
	SerpAPIBaseURL    = "https://serpapi.com"
	SerpAPISearchPath = "/search.json"

	SerpAPIParamEngine    = "engine"
	SerpAPIParamAPIKey    = "api_key"
	SerpAPIParamDeparture = "departure_id"
	SerpAPIParamArrival   = "arrival_id"
	SerpAPIParamDate      = "outbound_date"
	SerpAPIParamType      = "type"
	SerpAPIParamAdults    = "adults"
	SerpAPIParamCurrency  = "currency"
	SerpAPIParamClass     = "travel_class"
	SerpAPIParamLanguage  = "hl"
	SerpAPIParamCountry   = "gl"
	SerpAPIParamStops     = "stops"

	SerpAPIEngineFlights = "google_flights"
	SerpAPITypeOneWay    = "2"
	SerpAPITypeRoundTrip = "1"
	SerpAPITypeMultiCity = "3"

	SerpAPIParamMultiCityJSON  = "multi_city_json"
	SerpAPIParamDepartureToken = "departure_token"

	SerpAPIClassEconomy        = "1"
	SerpAPIClassPremiumEconomy = "2"
	SerpAPIClassBusiness       = "3"
	SerpAPIClassFirst          = "4"
)

// HTTP header keys.
const (
	HeaderContentType  = "Content-Type"
	HeaderRapidAPIKey  = "x-rapidapi-key"
	HeaderRapidAPIHost = "x-rapidapi-host"
	ContentTypeJSON    = "application/json"
)

// OpenAI API
const (
	OpenAIBaseURL          = "https://api.openai.com/v1"
	OpenAIChatCompletions  = "/chat/completions"
	OpenAIModelDefault     = "gpt-4o-mini"
	OpenAIMaxTokensDefault = 4096
)

// Anuma API (OpenAI-compatible LLM router)
const (
	AnumaBaseURL         = "https://portal.anuma-dev.ai/api/v1"
	AnumaChatCompletions = "/chat/completions"
	AnumaModelDefault    = "openai/gpt-4o-mini"
	AnumaAuthHeader      = "X-API-Key"
)

// Default query values.
const (
	DefaultCurrency    = "usd"
	DefaultLocale      = "en"
	DefaultResultLimit = "50"
	DefaultHandbags    = "1"
	DefaultHoldbags    = "0"
	DefaultChildren    = "0"
	DefaultInfants     = "0"
)
