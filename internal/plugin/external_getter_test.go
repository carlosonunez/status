package plugin_test

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// metadataInvoker returns a fake invoker that serves canned metadata JSON and
// canned get-events JSON for a getter plugin.
func metadataInvoker(metaJSON, eventsJSON string) plugin.InvokerFunc {
	return func(_ context.Context, _, flag string, _ io.Reader) ([]byte, error) {
		switch flag {
		case "--metadata":
			return []byte(metaJSON), nil
		case "--get-events":
			return []byte(eventsJSON), nil
		default:
			return nil, fmt.Errorf("unexpected flag: %s", flag)
		}
	}
}

// errorInvoker always returns an error, simulating a failing binary.
func errorGetterInvoker(msg string) plugin.InvokerFunc {
	return func(_ context.Context, _, _ string, _ io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("%s", msg)
	}
}

type externalGetterTest struct {
	TestName string
	Build    func(t *testing.T) (*plugin.ExternalGetter, error)
	Run      func(t *testing.T, g *plugin.ExternalGetter, buildErr error)
}

func (tc externalGetterTest) RunTest(t *testing.T) {
	t.Helper()
	g, err := tc.Build(t)
	tc.Run(t, g, err)
}

func TestExternalGetter(t *testing.T) {
	const fakeMeta = `{"name":"mycal","type":"getter","min_interval":"1m",` +
		`"param_specs":[{"name":"cal","description":"Calendar ID","required":true,"type":"string"}]}`
	const fakeEvents = `{"events":[{"calendar":"test-cal","title":"Fake Event",` +
		`"starts_at":"2026-01-01T09:00:00Z","ends_at":"2026-01-01T09:30:00Z","all_day":false}]}`

	tests := []externalGetterTest{
		{
			TestName: "name_from_metadata",
			Build: func(t *testing.T) (*plugin.ExternalGetter, error) {
				return plugin.NewExternalGetterWithInvoker("/fake/status-getter-mycal", metadataInvoker(fakeMeta, fakeEvents))
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter, buildErr error) {
				require.NoError(t, buildErr)
				assert.Equal(t, "mycal", g.Name())
			},
		},
		{
			TestName: "min_interval_parsed",
			Build: func(t *testing.T) (*plugin.ExternalGetter, error) {
				return plugin.NewExternalGetterWithInvoker("/fake/status-getter-mycal", metadataInvoker(fakeMeta, fakeEvents))
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter, buildErr error) {
				require.NoError(t, buildErr)
				assert.Equal(t, time.Minute, g.MinInterval())
			},
		},
		{
			TestName: "param_specs_from_metadata",
			Build: func(t *testing.T) (*plugin.ExternalGetter, error) {
				return plugin.NewExternalGetterWithInvoker("/fake/status-getter-mycal", metadataInvoker(fakeMeta, fakeEvents))
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter, buildErr error) {
				require.NoError(t, buildErr)
				specs := g.ParamSpecs()
				require.Len(t, specs, 1)
				assert.Equal(t, "cal", specs[0].Name)
			},
		},
		{
			TestName: "get_events_parses_response",
			Build: func(t *testing.T) (*plugin.ExternalGetter, error) {
				return plugin.NewExternalGetterWithInvoker("/fake/status-getter-mycal", metadataInvoker(fakeMeta, fakeEvents))
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter, buildErr error) {
				require.NoError(t, buildErr)
				events, err := g.GetEvents(context.Background(), params.FromMap(map[string]any{"cal": "primary"}))
				require.NoError(t, err)
				require.Len(t, events, 1)
				assert.Equal(t, "test-cal", events[0].Calendar)
				assert.Equal(t, "Fake Event", events[0].Title)
				assert.Equal(t, time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC), events[0].StartsAt)
			},
		},
		{
			TestName: "get_events_propagates_plugin_error",
			Build: func(t *testing.T) (*plugin.ExternalGetter, error) {
				errResp := `{"events":[],"error":"calendar not found"}`
				return plugin.NewExternalGetterWithInvoker("/fake/status-getter-mycal", metadataInvoker(fakeMeta, errResp))
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter, buildErr error) {
				require.NoError(t, buildErr)
				_, err := g.GetEvents(context.Background(), params.FromMap(nil))
				require.Error(t, err)
				assert.Contains(t, err.Error(), "calendar not found")
			},
		},
		{
			TestName: "constructor_returns_error_when_invoker_fails",
			Build: func(t *testing.T) (*plugin.ExternalGetter, error) {
				return plugin.NewExternalGetterWithInvoker("/fake/status-getter-broken", errorGetterInvoker("exec: no such file"))
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter, buildErr error) {
				require.Error(t, buildErr)
				assert.Nil(t, g)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}
