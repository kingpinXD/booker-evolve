// Package currency provides exchange rate conversion.
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

// Converter holds exchange rates and performs currency conversion.
type Converter struct {
	rates map[string]float64
}

// NewConverter creates a Converter with the given rates map.
func NewConverter(rates map[string]float64) *Converter {
	return &Converter{rates: rates}
}

// Rate returns the USD-to-target exchange rate.
func (c *Converter) Rate(target string) (float64, error) {
	r, ok := c.rates[target]
	if !ok {
		return 0, fmt.Errorf("unknown currency: %s", target)
	}
	return r, nil
}

// Convert converts a Money value to the target currency.
func (c *Converter) Convert(m types.Money, target string) (types.Money, error) {
	if m.Currency == target {
		return m, nil
	}
	r, err := c.Rate(target)
	if err != nil {
		return m, err
	}
	return types.Money{Amount: m.Amount * r, Currency: target}, nil
}

// Package-level default converter, initialized once via fetchRates.
var (
	once           sync.Once
	defaultConv    *Converter
	fallbackRates  = map[string]float64{"USD": 1.0, "CAD": 1.36, "EUR": 0.92, "GBP": 0.79}
)

type apiResponse struct {
	Result string             `json:"result"`
	Rates  map[string]float64 `json:"rates"`
}

func fetchRates() {
	once.Do(func() {
		resp, err := http.Get(rateURL)
		if err != nil {
			log.Printf("[currency] fetch failed, using fallback rates: %v", err)
			defaultConv = NewConverter(fallbackRates)
			return
		}
		defer func() { _ = resp.Body.Close() }()

		var data apiResponse
		if err := json.NewDecoder(resp.Body).Decode(&data); err != nil || data.Result != "success" {
			log.Printf("[currency] parse failed, using fallback rates: %v", err)
			defaultConv = NewConverter(fallbackRates)
			return
		}

		defaultConv = NewConverter(data.Rates)
		log.Printf("[currency] loaded %d live rates (USD→CAD: %.4f)", len(data.Rates), data.Rates["CAD"])
	})
}

// Rate returns the USD-to-target exchange rate using the default converter.
func Rate(target string) (float64, error) {
	fetchRates()
	return defaultConv.Rate(target)
}

// Convert converts a Money value to the target currency using the default converter.
func Convert(m types.Money, target string) (types.Money, error) {
	fetchRates()
	return defaultConv.Convert(m, target)
}
