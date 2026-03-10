package aggregator

import (
	"context"
	"errors"
	"testing"

	"booker/config"
	"booker/provider"
	"booker/types"
)

// mockProvider implements provider.Provider for testing.
type mockProvider struct {
	name    config.ProviderName
	flights []types.Flight
	err     error
}

func (m *mockProvider) Name() config.ProviderName { return m.name }
func (m *mockProvider) Search(_ context.Context, _ types.SearchRequest) ([]types.Flight, error) {
	return m.flights, m.err
}

func TestSearch_SingleProvider(t *testing.T) {
	reg := provider.NewRegistry()
	_ = reg.Register(&mockProvider{
		name: "test",
		flights: []types.Flight{
			{Price: types.Money{Amount: 300, Currency: "USD"}},
			{Price: types.Money{Amount: 100, Currency: "USD"}},
			{Price: types.Money{Amount: 200, Currency: "USD"}},
		},
	})

	agg := New(reg)
	result := agg.Search(context.Background(), types.SearchRequest{})

	if len(result.Flights) != 3 {
		t.Fatalf("got %d flights, want 3", len(result.Flights))
	}
	// Verify sorted by price ascending.
	for i := 1; i < len(result.Flights); i++ {
		if result.Flights[i].Price.Amount < result.Flights[i-1].Price.Amount {
			t.Errorf("flights not sorted: index %d ($%.0f) < index %d ($%.0f)",
				i, result.Flights[i].Price.Amount, i-1, result.Flights[i-1].Price.Amount)
		}
	}
	if len(result.Errors) != 0 {
		t.Errorf("got %d errors, want 0", len(result.Errors))
	}
}

func TestSearch_MultipleProvidersMergedAndSorted(t *testing.T) {
	reg := provider.NewRegistry()
	_ = reg.Register(&mockProvider{
		name:    "expensive",
		flights: []types.Flight{{Price: types.Money{Amount: 500, Currency: "USD"}}},
	})
	_ = reg.Register(&mockProvider{
		name:    "cheap",
		flights: []types.Flight{{Price: types.Money{Amount: 50, Currency: "USD"}}},
	})

	agg := New(reg)
	result := agg.Search(context.Background(), types.SearchRequest{})

	if len(result.Flights) != 2 {
		t.Fatalf("got %d flights, want 2", len(result.Flights))
	}
	if result.Flights[0].Price.Amount != 50 {
		t.Errorf("first flight price = %.0f, want 50", result.Flights[0].Price.Amount)
	}
	if result.Flights[1].Price.Amount != 500 {
		t.Errorf("second flight price = %.0f, want 500", result.Flights[1].Price.Amount)
	}
}

func TestSearch_PartialFailure(t *testing.T) {
	reg := provider.NewRegistry()
	_ = reg.Register(&mockProvider{
		name:    "good",
		flights: []types.Flight{{Price: types.Money{Amount: 200, Currency: "USD"}}},
	})
	_ = reg.Register(&mockProvider{
		name: "bad",
		err:  errors.New("provider down"),
	})

	agg := New(reg)
	result := agg.Search(context.Background(), types.SearchRequest{})

	if len(result.Flights) != 1 {
		t.Fatalf("got %d flights, want 1", len(result.Flights))
	}
	if len(result.Errors) != 1 {
		t.Fatalf("got %d errors, want 1", len(result.Errors))
	}
	if result.Errors[0].Provider != "bad" {
		t.Errorf("error provider = %q, want %q", result.Errors[0].Provider, "bad")
	}
}

func TestSearch_AllProvidersFail(t *testing.T) {
	reg := provider.NewRegistry()
	_ = reg.Register(&mockProvider{name: "fail1", err: errors.New("err1")})
	_ = reg.Register(&mockProvider{name: "fail2", err: errors.New("err2")})

	agg := New(reg)
	result := agg.Search(context.Background(), types.SearchRequest{})

	if len(result.Flights) != 0 {
		t.Errorf("got %d flights, want 0", len(result.Flights))
	}
	if len(result.Errors) != 2 {
		t.Errorf("got %d errors, want 2", len(result.Errors))
	}
}

func TestSearch_EmptyRegistry(t *testing.T) {
	reg := provider.NewRegistry()
	agg := New(reg)
	result := agg.Search(context.Background(), types.SearchRequest{})

	if len(result.Flights) != 0 {
		t.Errorf("got %d flights, want 0", len(result.Flights))
	}
	if len(result.Errors) != 0 {
		t.Errorf("got %d errors, want 0", len(result.Errors))
	}
}
