package setter

import (
	"context"
	"fmt"
	"sync"

	"github.com/carlosonunez/status/internal/pluginspec"
)

// StatusSetter applies a resolved Status to an external service (e.g. Slack,
// GitHub). Implementations register themselves in their package's init()
// function by calling the package-level Register function, and are wired into
// the binary via blank imports in cmd/status/main.go.
type StatusSetter interface {
	// Name returns the unique snake_case identifier used to reference this
	// setter in config (e.g. "slack", "github").
	Name() string

	// ParamSpecs describes the parameters this setter accepts.
	// Used by the "status integration list/show" commands.
	ParamSpecs() []pluginspec.ParamSpec

	// SetStatus applies s to the external service.
	SetStatus(ctx context.Context, s Status) (SetResult, error)
}

// Registry holds a set of named StatusSetters. Use NewRegistry in tests to get
// an isolated instance; never use the package-level default registry in tests.
type Registry struct {
	mu      sync.RWMutex
	setters map[string]StatusSetter
}

// NewRegistry returns an empty, isolated Registry.
func NewRegistry() *Registry {
	return &Registry{setters: make(map[string]StatusSetter)}
}

// Register adds s to the registry. Panics if a setter with the same name has
// already been registered — duplicate registration is a programming error.
func (r *Registry) Register(s StatusSetter) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, exists := r.setters[s.Name()]; exists {
		panic(fmt.Sprintf("setter: duplicate registration for %q", s.Name()))
	}
	r.setters[s.Name()] = s
}

// Get returns the StatusSetter registered under name, or an error if not found.
func (r *Registry) Get(name string) (StatusSetter, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	s, ok := r.setters[name]
	if !ok {
		return nil, fmt.Errorf("setter: no setter registered with name %q", name)
	}
	return s, nil
}

// All returns a snapshot of all registered StatusSetters in undefined order.
func (r *Registry) All() []StatusSetter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]StatusSetter, 0, len(r.setters))
	for _, s := range r.setters {
		out = append(out, s)
	}
	return out
}

// defaultRegistry is the package-level registry populated by plugin init() calls.
var defaultRegistry = NewRegistry()

// Register calls Register on the package-level default registry.
func Register(s StatusSetter) { defaultRegistry.Register(s) }

// Get calls Get on the package-level default registry.
func Get(name string) (StatusSetter, error) { return defaultRegistry.Get(name) }

// All calls All on the package-level default registry.
func All() []StatusSetter { return defaultRegistry.All() }

// DefaultRegistry returns the package-level default registry.
// Use this only when you need to pass the registry to a function that requires
// an explicit *Registry (e.g., plugin.DiscoverAllDefault).
func DefaultRegistry() *Registry { return defaultRegistry }
