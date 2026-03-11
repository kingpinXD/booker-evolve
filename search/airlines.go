// Package search — airline alliance reference data for ranking.
package search

// codeToAlliance maps airline IATA codes to their alliance name.
var codeToAlliance = func() map[string]string {
	alliances := map[string][]string{
		"Star Alliance": {
			"AC", "LH", "UA", "NH", "SQ", "TK", "ET", "AI", "SK", "OS",
			"LX", "TP", "MS", "CA", "NZ", "AV", "OU", "SN", "A3", "SA",
			"ZH", "OZ", "BR", "CM", "LO", "TG", "HO", "JP",
		},
		"OneWorld": {
			"AA", "BA", "CX", "QF", "JL", "IB", "AY", "MH", "QR", "RJ",
			"S7", "UL", "AT", "FJ",
		},
		"SkyTeam": {
			"DL", "AF", "KL", "AM", "KE", "SU", "CI", "CZ", "MU", "ME",
			"OK", "AR", "RO", "VN", "GA", "SV", "UX", "XN",
		},
	}
	m := make(map[string]string, 60)
	for name, codes := range alliances {
		for _, c := range codes {
			m[c] = name
		}
	}
	return m
}()

// Alliance returns the alliance name for the given airline IATA code,
// or "" if the code is not a member of any major alliance.
func Alliance(code string) string {
	return codeToAlliance[code]
}

// SameAlliance reports whether two airline IATA codes belong to the
// same alliance. Returns false if either code is unknown.
func SameAlliance(a, b string) bool {
	allianceA := codeToAlliance[a]
	return allianceA != "" && allianceA == codeToAlliance[b]
}
