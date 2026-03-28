package getter_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeGetter is a minimal EventGetter for registry tests.
type fakeGetter struct{ name string }

func (f *fakeGetter) Name() string                        { return f.name }
func (f *fakeGetter) ParamSpecs() []pluginspec.ParamSpec  { return nil }
func (f *fakeGetter) MinInterval() time.Duration          { return 0 }
func (f *fakeGetter) GetEvents(_ context.Context, _ params.Params) ([]getter.Event, error) {
	return nil, nil
}

func newFake(name string) getter.EventGetter { return &fakeGetter{name: name} }

func TestRegistry_Register_StoresByName(t *testing.T) {
	r := getter.NewRegistry()
	r.Register(newFake("gcal"))
	g, err := r.Get("gcal")
	require.NoError(t, err)
	assert.Equal(t, "gcal", g.Name())
}

func TestRegistry_Register_PanicOnDuplicate(t *testing.T) {
	r := getter.NewRegistry()
	r.Register(newFake("gcal"))
	assert.Panics(t, func() { r.Register(newFake("gcal")) })
}

func TestRegistry_Get_UnknownNameReturnsError(t *testing.T) {
	r := getter.NewRegistry()
	_, err := r.Get("nonexistent")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "nonexistent")
}

func TestRegistry_All_ReturnsAllRegistered(t *testing.T) {
	r := getter.NewRegistry()
	for _, name := range []string{"a", "b", "c"} {
		r.Register(newFake(name))
	}
	all := r.All()
	assert.Len(t, all, 3)
	names := make([]string, len(all))
	for i, g := range all {
		names[i] = g.Name()
	}
	assert.ElementsMatch(t, []string{"a", "b", "c"}, names)
}

func TestRegistry_Isolation(t *testing.T) {
	r1 := getter.NewRegistry()
	r2 := getter.NewRegistry()
	r1.Register(newFake("only-in-r1"))
	_, err := r2.Get("only-in-r1")
	assert.Error(t, err, "r2 should not see getters registered in r1")
}

func TestRegistry_All_EmptyWhenNothingRegistered(t *testing.T) {
	r := getter.NewRegistry()
	assert.Empty(t, r.All())
}

func TestRegistry_Register_MultipleDistinctNames(t *testing.T) {
	r := getter.NewRegistry()
	for i := range 10 {
		r.Register(newFake(fmt.Sprintf("getter-%d", i)))
	}
	assert.Len(t, r.All(), 10)
}
