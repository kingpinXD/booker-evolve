package provider

import (
	"context"
	"sync"
	"testing"

	"booker/config"
	"booker/types"
)

// mockProvider implements Provider for testing.
type mockProvider struct {
	name config.ProviderName
}

func (m *mockProvider) Name() config.ProviderName { return m.name }
func (m *mockProvider) Search(context.Context, types.SearchRequest) ([]types.Flight, error) {
	return nil, nil
}

func TestRegister(t *testing.T) {
	r := NewRegistry()
	p := &mockProvider{name: "test"}

	if err := r.Register(p); err != nil {
		t.Fatalf("first Register failed: %v", err)
	}

	err := r.Register(p)
	if err == nil {
		t.Fatal("duplicate Register should return error")
	}
}

func TestGet(t *testing.T) {
	r := NewRegistry()
	p := &mockProvider{name: "test"}
	if err := r.Register(p); err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	got, ok := r.Get("test")
	if !ok || got != p {
		t.Fatal("Get should return registered provider")
	}

	_, ok = r.Get("missing")
	if ok {
		t.Fatal("Get should return false for unregistered name")
	}
}

func TestAll(t *testing.T) {
	r := NewRegistry()
	if all := r.All(); len(all) != 0 {
		t.Fatalf("empty registry All() returned %d providers", len(all))
	}

	if err := r.Register(&mockProvider{name: "a"}); err != nil {
		t.Fatalf("Register a failed: %v", err)
	}
	if err := r.Register(&mockProvider{name: "b"}); err != nil {
		t.Fatalf("Register b failed: %v", err)
	}

	all := r.All()
	if len(all) != 2 {
		t.Fatalf("All() returned %d, want 2", len(all))
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := NewRegistry()
	var wg sync.WaitGroup

	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			p := &mockProvider{name: config.ProviderName(string(rune('A' + n%26)))}
			_ = r.Register(p) // some will fail with duplicate — that's fine
			r.Get(p.Name())
			r.All()
		}(i)
	}
	wg.Wait()
}
