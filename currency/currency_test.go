package currency

import (
	"math"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"booker/types"
)

func testConverter() *Converter {
	return NewConverter(map[string]float64{
		"USD": 1.0,
		"CAD": 1.36,
		"EUR": 0.92,
	})
}

func TestConvert_SameCurrency(t *testing.T) {
	c := testConverter()
	m := types.Money{Amount: 100, Currency: "USD"}
	got, err := c.Convert(m, "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Amount != 100 || got.Currency != "USD" {
		t.Errorf("got %v, want {100 USD}", got)
	}
}

func TestConvert_USDToCAD(t *testing.T) {
	c := testConverter()
	m := types.Money{Amount: 100, Currency: "USD"}
	got, err := c.Convert(m, "CAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(got.Amount-136.0) > 0.01 {
		t.Errorf("got amount %.2f, want 136.00", got.Amount)
	}
	if got.Currency != "CAD" {
		t.Errorf("got currency %q, want %q", got.Currency, "CAD")
	}
}

func TestConvert_UnknownCurrency(t *testing.T) {
	c := testConverter()
	m := types.Money{Amount: 100, Currency: "USD"}
	_, err := c.Convert(m, "XYZ")
	if err == nil {
		t.Fatal("expected error for unknown currency, got nil")
	}
}

func TestRate_Known(t *testing.T) {
	c := testConverter()
	r, err := c.Rate("EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(r-0.92) > 0.001 {
		t.Errorf("got rate %.4f, want 0.92", r)
	}
}

func TestRate_Unknown(t *testing.T) {
	c := testConverter()
	_, err := c.Rate("XYZ")
	if err == nil {
		t.Fatal("expected error for unknown currency, got nil")
	}
}

// resetGlobals resets the package-level sync.Once and defaultConv so
// fetchRates can be exercised again in each test.
func resetGlobals(t *testing.T) {
	t.Helper()
	once = sync.Once{}
	defaultConv = nil
}

func TestFetchRates_Success(t *testing.T) {
	resetGlobals(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"success","rates":{"USD":1.0,"CAD":1.40,"EUR":0.88}}`))
	}))
	defer srv.Close()

	origURL := rateURL
	rateURL = srv.URL
	defer func() { rateURL = origURL }()

	fetchRates()

	if defaultConv == nil {
		t.Fatal("defaultConv is nil after fetchRates")
	}
	r, err := defaultConv.Rate("CAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(r-1.40) > 0.001 {
		t.Errorf("got CAD rate %.4f, want 1.40", r)
	}
}

func TestFetchRates_HTTPError(t *testing.T) {
	resetGlobals(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer srv.Close()

	origURL := rateURL
	rateURL = srv.URL
	defer func() { rateURL = origURL }()

	fetchRates()

	if defaultConv == nil {
		t.Fatal("defaultConv is nil after fetchRates with HTTP error")
	}
	// Should fall back to hardcoded rates.
	r, err := defaultConv.Rate("CAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(r-1.36) > 0.001 {
		t.Errorf("got fallback CAD rate %.4f, want 1.36", r)
	}
}

func TestFetchRates_MalformedJSON(t *testing.T) {
	resetGlobals(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	origURL := rateURL
	rateURL = srv.URL
	defer func() { rateURL = origURL }()

	fetchRates()

	if defaultConv == nil {
		t.Fatal("defaultConv is nil after fetchRates with malformed JSON")
	}
	r, err := defaultConv.Rate("EUR")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(r-0.92) > 0.001 {
		t.Errorf("got fallback EUR rate %.4f, want 0.92", r)
	}
}

func TestFetchRates_NonSuccessResult(t *testing.T) {
	resetGlobals(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"error","rates":{}}`))
	}))
	defer srv.Close()

	origURL := rateURL
	rateURL = srv.URL
	defer func() { rateURL = origURL }()

	fetchRates()

	if defaultConv == nil {
		t.Fatal("defaultConv is nil after fetchRates with non-success result")
	}
	// Should fall back to hardcoded rates.
	r, err := defaultConv.Rate("GBP")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(r-0.79) > 0.001 {
		t.Errorf("got fallback GBP rate %.4f, want 0.79", r)
	}
}

func TestPackageLevel_Rate(t *testing.T) {
	resetGlobals(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"success","rates":{"USD":1.0,"JPY":150.5}}`))
	}))
	defer srv.Close()

	origURL := rateURL
	rateURL = srv.URL
	defer func() { rateURL = origURL }()

	r, err := Rate("JPY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(r-150.5) > 0.001 {
		t.Errorf("got JPY rate %.4f, want 150.5", r)
	}

	// Unknown currency through package-level Rate.
	_, err = Rate("ZZZ")
	if err == nil {
		t.Fatal("expected error for unknown currency via package-level Rate")
	}
}

func TestPackageLevel_Convert(t *testing.T) {
	resetGlobals(t)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"result":"success","rates":{"USD":1.0,"CAD":1.35}}`))
	}))
	defer srv.Close()

	origURL := rateURL
	rateURL = srv.URL
	defer func() { rateURL = origURL }()

	m := types.Money{Amount: 200, Currency: "USD"}
	got, err := Convert(m, "CAD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if math.Abs(got.Amount-270.0) > 0.01 {
		t.Errorf("got amount %.2f, want 270.00", got.Amount)
	}
	if got.Currency != "CAD" {
		t.Errorf("got currency %q, want %q", got.Currency, "CAD")
	}

	// Same currency should return unchanged.
	same, err := Convert(m, "USD")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if same.Amount != 200 || same.Currency != "USD" {
		t.Errorf("same-currency convert: got %v, want {200 USD}", same)
	}
}
