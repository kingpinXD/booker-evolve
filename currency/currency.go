// Package currency provides live exchange rate conversion.
//
// Rates are fetched once from open.er-api.com (free, no key required)
// and cached in memory for the lifetime of the process.
package currency

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"booker/types"
)

const rateURL = "https://open.er-api.com/v6/latest/USD"

var (
	once  sync.Once
	rates map[string]float64
)

type apiResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

// fetchRates loads rates once, falling back to hardcoded values on failure.
func fetchRates() {
	once.Do(func() {
		resp, err := http.Get(rateURL)
		if err != nil {
			log.Printf("[currency] fetch failed, using fallback rates: %v", err)
			rates = map[string]float64{"USD": 1.0, "CAD": 1.36, "EUR": 0.92, "GBP": 0.79}
			return
		}
		defer func() { _ = resp.Body.Close() }()

		var data apiResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Result != "success" {
			log.Printf("[currency] parse failed, using fallback rates: %v", err)
			rates = map[string]float64{"USD": 1.0, "CAD": 1.36, "EUR": 0.92, "GBP": 0.79}
			return
		}

		rates = data.Rates
		log.Printf("[currency] loaded %d live rates (USD→CAD: %.4f)", len(rates), rates["CAD"])
	})
}

// Rate returns the USD→target exchange rate.
func Rate(target string) (float64, error) {
	fetchRates()
	r, ok := rates[target]
	if !ok {
		return 0, fmt.Errorf("unknown currency: %s", target)
	}
	return r, nil
}

// Convert converts a Money value to the target currency.
func Convert(m types.Money, target string) (types.Money, error) {
	if m.Currency == target {
		return m, nil
	}
	r, err := Rate(target)
	if err != nil {
		return m, err
	}
	return types.Money{Amount: m.Amount * r, Currency: target}, nil
}
