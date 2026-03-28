package plugin_test

import (
	"context"
	"fmt"
	"io"
	"testing"

	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/plugin"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setterInvoker returns a fake invoker that serves canned metadata JSON and
// canned set-status JSON for a setter plugin.
func setterInvoker(metaJSON, resultJSON string) plugin.InvokerFunc {
	return func(_ context.Context, _, flag string, _ io.Reader) ([]byte, error) {
		switch flag {
		case "--metadata":
			return []byte(metaJSON), nil
		case "--set-status":
			return []byte(resultJSON), nil
		default:
			return nil, fmt.Errorf("unexpected flag: %s", flag)
		}
	}
}

// errorSetterInvoker always returns an error, simulating a failing binary.
func errorSetterInvoker(msg string) plugin.InvokerFunc {
	return func(_ context.Context, _, _ string, _ io.Reader) ([]byte, error) {
		return nil, fmt.Errorf("%s", msg)
	}
}

type externalSetterTest struct {
	TestName string
	Build    func(t *testing.T) (*plugin.ExternalSetter, error)
	Run      func(t *testing.T, s *plugin.ExternalSetter, buildErr error)
}

func (tc externalSetterTest) RunTest(t *testing.T) {
	t.Helper()
	s, err := tc.Build(t)
	tc.Run(t, s, err)
}

func TestExternalSetter(t *testing.T) {
	const fakeMeta = `{"name":"myslack","type":"setter",` +
		`"param_specs":[{"name":"status_message","description":"Message","required":true,"type":"string"}]}`
	const fakeResult = `{"skipped":false,"response":"ok"}`

	tests := []externalSetterTest{
		{
			TestName: "name_from_metadata",
			Build: func(t *testing.T) (*plugin.ExternalSetter, error) {
				return plugin.NewExternalSetterWithInvoker("/fake/status-setter-myslack", setterInvoker(fakeMeta, fakeResult))
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter, buildErr error) {
				require.NoError(t, buildErr)
				assert.Equal(t, "myslack", s.Name())
			},
		},
		{
			TestName: "param_specs_from_metadata",
			Build: func(t *testing.T) (*plugin.ExternalSetter, error) {
				return plugin.NewExternalSetterWithInvoker("/fake/status-setter-myslack", setterInvoker(fakeMeta, fakeResult))
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter, buildErr error) {
				require.NoError(t, buildErr)
				specs := s.ParamSpecs()
				require.Len(t, specs, 1)
				assert.Equal(t, "status_message", specs[0].Name)
			},
		},
		{
			TestName: "set_status_parses_response",
			Build: func(t *testing.T) (*plugin.ExternalSetter, error) {
				return plugin.NewExternalSetterWithInvoker("/fake/status-setter-myslack", setterInvoker(fakeMeta, fakeResult))
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter, buildErr error) {
				require.NoError(t, buildErr)
				st := setter.Status{Params: params.FromMap(map[string]any{"status_message": "In a meeting"})}
				result, err := s.SetStatus(context.Background(), st)
				require.NoError(t, err)
				assert.False(t, result.Skipped)
				assert.Equal(t, "ok", result.Response)
			},
		},
		{
			TestName: "set_status_skipped_response",
			Build: func(t *testing.T) (*plugin.ExternalSetter, error) {
				skippedResult := `{"skipped":true}`
				return plugin.NewExternalSetterWithInvoker("/fake/status-setter-myslack", setterInvoker(fakeMeta, skippedResult))
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter, buildErr error) {
				require.NoError(t, buildErr)
				st := setter.Status{Params: params.FromMap(map[string]any{"status_message": "busy"})}
				result, err := s.SetStatus(context.Background(), st)
				require.NoError(t, err)
				assert.True(t, result.Skipped)
			},
		},
		{
			TestName: "set_status_propagates_plugin_error",
			Build: func(t *testing.T) (*plugin.ExternalSetter, error) {
				errResult := `{"skipped":false,"error":"rate limited"}`
				return plugin.NewExternalSetterWithInvoker("/fake/status-setter-myslack", setterInvoker(fakeMeta, errResult))
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter, buildErr error) {
				require.NoError(t, buildErr)
				st := setter.Status{Params: params.FromMap(map[string]any{"status_message": "busy"})}
				_, err := s.SetStatus(context.Background(), st)
				require.Error(t, err)
				assert.Contains(t, err.Error(), "rate limited")
			},
		},
		{
			TestName: "constructor_returns_error_when_invoker_fails",
			Build: func(t *testing.T) (*plugin.ExternalSetter, error) {
				return plugin.NewExternalSetterWithInvoker("/fake/status-setter-broken", errorSetterInvoker("exec: no such file"))
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter, buildErr error) {
				require.Error(t, buildErr)
				assert.Nil(t, s)
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}
