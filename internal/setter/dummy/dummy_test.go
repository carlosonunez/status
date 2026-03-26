// Tests use dummy.New() directly and never touch the global setter registry.
package dummy_test

import (
	"context"
	"testing"

	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/carlosonunez/status/internal/setter/dummy"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDummySetter_Name(t *testing.T) {
	assert.Equal(t, "dummy", dummy.New().Name())
}

func TestDummySetter_Received_InitiallyEmpty(t *testing.T) {
	assert.Empty(t, dummy.New().Received())
}

func TestDummySetter_SetStatus_RecordsStatus(t *testing.T) {
	d := dummy.New()
	s := setter.Status{Params: params.FromMap(map[string]any{"status_message": "In a meeting"})}
	result, err := d.SetStatus(context.Background(), s)
	require.NoError(t, err)
	assert.False(t, result.Skipped)
	assert.Equal(t, "ok", result.Response)

	got := d.Received()
	require.Len(t, got, 1)
	msg, ok := got[0].Params.String("status_message")
	require.True(t, ok)
	assert.Equal(t, "In a meeting", msg)
}

func TestDummySetter_SetStatus_MultipleCallsAccumulate(t *testing.T) {
	d := dummy.New()
	for i, msg := range []string{"first", "second", "third"} {
		s := setter.Status{Params: params.FromMap(map[string]any{"status_message": msg})}
		_, err := d.SetStatus(context.Background(), s)
		require.NoError(t, err, "call %d failed", i)
	}
	assert.Len(t, d.Received(), 3)
}

func TestDummySetter_Received_ReturnsCopyNotSlice(t *testing.T) {
	d := dummy.New()
	s := setter.Status{Params: params.FromMap(map[string]any{"status_message": "original"})}
	_, _ = d.SetStatus(context.Background(), s)

	got := d.Received()
	got[0] = setter.Status{Params: params.FromMap(map[string]any{"status_message": "mutated"})}

	got2 := d.Received()
	msg, _ := got2[0].Params.String("status_message")
	assert.Equal(t, "original", msg, "mutation of returned slice should not affect setter state")
}

func TestDummySetter_ParamSpecs(t *testing.T) {
	assert.Nil(t, dummy.New().ParamSpecs())
}
