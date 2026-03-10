package currency

import (
	"math"
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
