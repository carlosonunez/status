// Package dummy provides a recording StatusSetter for use in tests and
// development. It is wired into the binary via a blank import in
// cmd/status/main.go.
package dummy

import (
	"context"
	"sync"

	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/carlosonunez/status/internal/setter"
)

const name = "dummy"

// DummySetter is a StatusSetter that records every Status it receives.
// It is intended for testing and local development.
type DummySetter struct {
	mu       sync.Mutex
	received []setter.Status
}

// New creates a DummySetter with an empty received list.
func New() *DummySetter { return &DummySetter{} }

// Received returns a copy of all statuses received so far, in call order.
func (d *DummySetter) Received() []setter.Status {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]setter.Status, len(d.received))
	copy(out, d.received)
	return out
}

func (d *DummySetter) Name() string { return name }

func (d *DummySetter) ParamSpecs() []pluginspec.ParamSpec { return nil }

func (d *DummySetter) SetStatus(_ context.Context, s setter.Status) (setter.SetResult, error) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.received = append(d.received, s)
	return setter.SetResult{Response: "ok"}, nil
}

func init() {
	setter.Register(New())
}
