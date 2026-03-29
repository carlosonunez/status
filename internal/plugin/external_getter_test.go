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

func TestExternalGetterAuthenticate(t *testing.T) {
	const fakeMetaNoAuth = `{"name":"mycal","type":"getter","min_interval":"1m"}`
	const fakeMetaWithAuth = `{"name":"mycal","type":"getter","min_interval":"1m","supports_auth":true}`
	const fakeAuthResp = `{"tokens":{"access_token":"tok123","token_secret":"sec456"}}`
	const fakeAuthErrResp = `{"error":"invalid credentials"}`

	authInvoker := func(authJSON string) plugin.InvokerFunc {
		return func(_ context.Context, _, flag string, _ io.Reader) ([]byte, error) {
			switch flag {
			case "--metadata":
				return []byte(fakeMetaWithAuth), nil
			case "--authenticate":
				return []byte(authJSON), nil
			default:
				return nil, fmt.Errorf("unexpected flag: %s", flag)
			}
		}
	}

	t.Run("authenticate_stores_tokens_when_supported", func(t *testing.T) {
		g, err := plugin.NewExternalGetterWithInvoker("/fake/status-getter-mycal", authInvoker(fakeAuthResp))
		require.NoError(t, err)
		assert.True(t, g.SupportsAuth())

		store := &fakeTokenStore{}
		err = g.Authenticate(context.Background(), store)
		require.NoError(t, err)
		assert.Equal(t, "tok123", store.tokens["mycal/access_token"])
		assert.Equal(t, "sec456", store.tokens["mycal/token_secret"])
	})

	t.Run("authenticate_returns_error_when_binary_errors", func(t *testing.T) {
		g, err := plugin.NewExternalGetterWithInvoker("/fake/status-getter-mycal", authInvoker(fakeAuthErrResp))
		require.NoError(t, err)

		store := &fakeTokenStore{}
		err = g.Authenticate(context.Background(), store)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid credentials")
	})

	t.Run("getter_without_supports_auth_returns_error", func(t *testing.T) {
		noAuthInvoker := func(_ context.Context, _, flag string, _ io.Reader) ([]byte, error) {
			if flag == "--metadata" {
				return []byte(fakeMetaNoAuth), nil
			}
			return nil, fmt.Errorf("unexpected flag: %s", flag)
		}
		g, err := plugin.NewExternalGetterWithInvoker("/fake/status-getter-mycal", plugin.InvokerFunc(noAuthInvoker))
		require.NoError(t, err)
		assert.False(t, g.SupportsAuth())

		store := &fakeTokenStore{}
		err = g.Authenticate(context.Background(), store)
		require.Error(t, err)
	})
}

// fakeTokenStore is an in-memory pluginspec.TokenStore for tests.
type fakeTokenStore struct {
	tokens map[string]string
}

func (s *fakeTokenStore) Get(service, key string) (string, error) {
	if s.tokens == nil {
		return "", fmt.Errorf("not found")
	}
	v, ok := s.tokens[service+"/"+key]
	if !ok {
		return "", fmt.Errorf("not found")
	}
	return v, nil
}

func (s *fakeTokenStore) Set(service, key, value string) error {
	if s.tokens == nil {
		s.tokens = make(map[string]string)
	}
	s.tokens[service+"/"+key] = value
	return nil
}

func (s *fakeTokenStore) Delete(service, key string) error {
	delete(s.tokens, service+"/"+key)
	return nil
}
