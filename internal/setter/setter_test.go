package setter_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeSetter is a minimal StatusSetter for registry tests.
type fakeSetter struct{ name string }

func (f *fakeSetter) Name() string                         { return f.name }
func (f *fakeSetter) ParamSpecs() []pluginspec.ParamSpec   { return nil }
func (f *fakeSetter) SetStatus(_ context.Context, _ setter.Status) (setter.SetResult, error) {
	return setter.SetResult{}, nil
}

func newFake(name string) setter.StatusSetter { return &fakeSetter{name: name} }

func TestRegistry_Register_StoresByName(t *testing.T) {
	r := setter.NewRegistry()
	r.Register(newFake("slack"))
	s, err := r.Get("slack")
	require.NoError(t, err)
	assert.Equal(t, "slack", s.Name())
}

func TestRegistry_Register_PanicOnDuplicate(t *testing.T) {
	r := setter.NewRegistry()
	r.Register(newFake("slack"))
	assert.Panics(t, func() { r.Register(newFake("slack")) })
}

func TestRegistry_Get_UnknownNameReturnsError(t *testing.T) {
	r := setter.NewRegistry()
	_, err := r.Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestRegistry_All_ReturnsAllRegistered(t *testing.T) {
	r := setter.NewRegistry()
	for _, name := range []string{"slack", "github", "whatsapp"} {
		r.Register(newFake(name))
	}
	all := r.All()
	assert.Len(t, all, 3)
	names := make([]string, len(all))
	for i, s := range all {
		names[i] = s.Name()
	}
	assert.ElementsMatch(t, []string{"slack", "github", "whatsapp"}, names)
}

func TestRegistry_Isolation(t *testing.T) {
	r1 := setter.NewRegistry()
	r2 := setter.NewRegistry()
	r1.Register(newFake("only-in-r1"))
	_, err := r2.Get("only-in-r1")
	assert.Error(t, err, "r2 should not see setters registered in r1")
}

func TestRegistry_All_EmptyWhenNothingRegistered(t *testing.T) {
	r := setter.NewRegistry()
	assert.Empty(t, r.All())
}

func TestRegistry_Register_MultipleDistinctNames(t *testing.T) {
	r := setter.NewRegistry()
	for i := range 10 {
		r.Register(newFake(fmt.Sprintf("setter-%d", i)))
	}
	assert.Len(t, r.All(), 10)
}
