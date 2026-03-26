// Tests use dummy.New() directly and never touch the global getter registry.
package dummy_test

import (
	"context"
	"testing"
	"time"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/getter/dummy"
	"github.com/carlosonunez/status/internal/params"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDummyGetter_Name(t *testing.T) {
	assert.Equal(t, "dummy", dummy.New().Name())
}

func TestDummyGetter_GetEvents_ReturnsConfiguredEvents(t *testing.T) {
	events := []getter.Event{
		{Title: "Team standup", Calendar: "Work"},
		{Title: "Lunch", Calendar: "Personal"},
	}
	d := dummy.New(events...)
	got, err := d.GetEvents(context.Background(), params.FromMap(nil))
	require.NoError(t, err)
	assert.Equal(t, events, got)
}

func TestDummyGetter_GetEvents_EmptyByDefault(t *testing.T) {
	d := dummy.New()
	got, err := d.GetEvents(context.Background(), params.FromMap(nil))
	require.NoError(t, err)
	assert.Empty(t, got)
}

func TestDummyGetter_GetEvents_ReturnsCopyNotSlice(t *testing.T) {
	original := getter.Event{Title: "Original"}
	d := dummy.New(original)
	got, _ := d.GetEvents(context.Background(), params.FromMap(nil))
	got[0].Title = "Modified"
	got2, _ := d.GetEvents(context.Background(), params.FromMap(nil))
	assert.Equal(t, "Original", got2[0].Title, "mutation of returned slice should not affect getter")
}

func TestDummyGetter_MinInterval(t *testing.T) {
	assert.Equal(t, time.Second, dummy.New().MinInterval())
}

func TestDummyGetter_ParamSpecs(t *testing.T) {
	assert.Nil(t, dummy.New().ParamSpecs())
}
