package getter

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/pluginspec"
)

// EventGetter retrieves calendar or scheduling events from an external source.
// Implementations register themselves in their package's init() function by
// calling the package-level Register function, and are wired into the binary
// via blank imports in cmd/status/main.go.
type EventGetter interface {
	// Name returns the unique snake_case identifier used to reference this
	// getter in config (e.g. "google_calendar", "tripit").
	Name() string

	// ParamSpecs describes the parameters this getter accepts.
	// Used by the "status integration list/show" commands.
	ParamSpecs() []pluginspec.ParamSpec

	// GetEvents fetches currently-active events using the supplied params.
	GetEvents(ctx context.Context, p params.Params) ([]Event, error)

	// MinInterval is the shortest polling period this getter supports.
	// The engine will not call GetEvents more frequently than this.
	MinInterval() time.Duration
}

// Registry holds a set of named EventGetters. Use NewRegistry in tests to get
// an isolated instance; never use the package-level default registry in tests.
type Registry struct {
	mu      sync.RWMutex
	getters map[string]EventGetter
}

// NewRegistry returns an empty, isolated Registry.
func NewRegistry() *Registry {
	return &Registry{getters: make(map[string]EventGetter)}
}

// Register adds g to the registry. Panics if a getter with the same name has
// already been registered — duplicate registration is a programming error.
func (r *Registry) Register(g EventGetter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.getters[g.Name()]; exists {
		panic(fmt.Sprintf("getter: duplicate registration for %q", g.Name()))
	}
	r.getters[g.Name()] = g
}

// Get returns the EventGetter registered under name, or an error if not found.
func (r *Registry) Get(name string) (EventGetter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	g, ok := r.getters[name]
	if !ok {
		return nil, fmt.Errorf("getter: no getter registered with name %q", name)
	}
	return g, nil
}

// All returns a snapshot of all registered EventGetters in undefined order.
func (r *Registry) All() []EventGetter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]EventGetter, 0, len(r.getters))
	for _, g := range r.getters {
		out = append(out, g)
	}
	return out
}

// defaultRegistry is the package-level registry populated by plugin init() calls.
var defaultRegistry = NewRegistry()

// Register calls Register on the package-level default registry.
// Plugin packages call this from their init() function.
func Register(g EventGetter) { defaultRegistry.Register(g) }

// Get calls Get on the package-level default registry.
func Get(name string) (EventGetter, error) { return defaultRegistry.Get(name) }

// All calls All on the package-level default registry.
func All() []EventGetter { return defaultRegistry.All() }

// DefaultRegistry returns the package-level default registry.
// Use this only when you need to pass the registry to a function that requires
// an explicit *Registry (e.g., plugin.DiscoverAllDefault).
func DefaultRegistry() *Registry { return defaultRegistry }
