// Package dummy provides a no-op EventGetter for use in tests and development.
// It is wired into the binary via a blank import in cmd/status/main.go.
package dummy

import (
	"context"
	"time"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/pluginspec"
)

const name = "dummy"

// DummyGetter is an EventGetter that returns a fixed, pre-configured slice of
// Events. It is intended for testing and local development.
type DummyGetter struct {
	events []getter.Event
}

// New creates a DummyGetter that returns the given events on every GetEvents call.
// Call with no arguments to get a getter that always returns an empty slice.
func New(events ...getter.Event) *DummyGetter {
	return &DummyGetter{events: events}
}

func (d *DummyGetter) Name() string { return name }
func (d *DummyGetter) Hidden() bool  { return true }

func (d *DummyGetter) ParamSpecs() []pluginspec.ParamSpec { return nil }

func (d *DummyGetter) GetEvents(_ context.Context, _ params.Params) ([]getter.Event, error) {
	out := make([]getter.Event, len(d.events))
	copy(out, d.events)
	return out, nil
}

// MinInterval returns one second — the smallest sensible polling floor for a
// getter that returns static data.
func (d *DummyGetter) MinInterval() time.Duration { return time.Second }

func init() {
	getter.Register(New())
}
