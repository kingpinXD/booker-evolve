// Package search — airport cluster data for nearby-airport expansion.
package search

// airportClusters maps metro area names to the IATA codes of airports
// serving that area. Only metros with 2+ airports are included.
var airportClusters = map[string][]string{
	"New York":      {"JFK", "EWR", "LGA"},
	"London":        {"LHR", "LGW", "STN", "LTN"},
	"Tokyo":         {"NRT", "HND"},
	"Paris":         {"CDG", "ORY"},
	"Chicago":       {"ORD", "MDW"},
	"Los Angeles":   {"LAX", "SNA", "BUR", "LGB"},
	"San Francisco": {"SFO", "OAK", "SJC"},
	"Washington DC": {"IAD", "DCA", "BWI"},
	"Toronto":       {"YYZ", "YTZ"},
	"Shanghai":      {"PVG", "SHA"},
	"Seoul":         {"ICN", "GMP"},
	"Milan":         {"MXP", "LIN"},
	"Houston":       {"IAH", "HOU"},
	"Dallas":        {"DFW", "DAL"},
	"Bangkok":       {"BKK", "DMK"},
	"Istanbul":      {"IST", "SAW"},
	"Beijing":       {"PEK", "PKX"},
	"Osaka":         {"KIX", "ITM"},
	"Rome":          {"FCO", "CIA"},
	"Taipei":        {"TPE", "TSA"},
	"Miami":         {"MIA", "FLL"},
	"Sao Paulo":     {"GRU", "CGH", "VCP"},
}

// codeToCluster is a reverse index from IATA code to its cluster slice,
// built once at init time.
var codeToCluster map[string][]string

func init() {
	codeToCluster = make(map[string][]string, 60)
	for _, codes := range airportClusters {
		for _, c := range codes {
			codeToCluster[c] = codes
		}
	}
}

// NearbyAirports returns sibling airports for the given IATA code,
// excluding the code itself. Returns nil if the code is not in any cluster.
func NearbyAirports(code string) []string {
	cluster := codeToCluster[code]
	if cluster == nil {
		return nil
	}
	siblings := make([]string, 0, len(cluster)-1)
	for _, c := range cluster {
		if c != code {
			siblings = append(siblings, c)
		}
	}
	return siblings
}
