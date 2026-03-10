package search

import (
	"context"
	"testing"
)

// fakeStrategy implements Strategy for testing.
type fakeStrategy struct {
	name string
	desc string
}

func (f *fakeStrategy) Name() string        { return f.name }
func (f *fakeStrategy) Description() string { return f.desc }
func (f *fakeStrategy) Search(_ context.Context, _ Request) ([]Itinerary, error) {
	return nil, nil
}

func TestPicker_SingleStrategy(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	p := NewPicker(nil, direct)

	got, err := p.Pick(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q", got.Name(), "direct")
	}
}

func TestPicker_FallbackNoContext(t *testing.T) {
	direct := &fakeStrategy{name: "direct", desc: "Direct flights"}
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	p := NewPicker(nil, direct, mc)

	// No context provided — should fall back to "direct".
	got, err := p.Pick(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "direct" {
		t.Errorf("got strategy %q, want %q", got.Name(), "direct")
	}
}

func TestPicker_FallbackWhenNoDirectExists(t *testing.T) {
	mc := &fakeStrategy{name: "multicity", desc: "Multi-city with stopover"}
	p := NewPicker(nil, mc)

	got, err := p.Pick(context.Background(), Request{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.Name() != "multicity" {
		t.Errorf("got strategy %q, want %q", got.Name(), "multicity")
	}
}
