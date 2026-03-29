package pluginsdk_test

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/carlosonunez/status/pkg/pluginsdk"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fakes ---

type fakeGetter struct{}

func (fakeGetter) Metadata() pluginsdk.GetterMetadata {
	return pluginsdk.GetterMetadata{
		Name:        "fake-getter",
		MinInterval: "30s",
		ParamSpecs: []pluginsdk.ParamSpec{
			{Name: "calendar_id", Description: "Calendar ID", Required: true, Type: "string"},
		},
	}
}

func (fakeGetter) GetEvents(params map[string]any) ([]pluginsdk.Event, error) {
	cal, _ := params["calendar_id"].(string)
	return []pluginsdk.Event{
		{
			Calendar: cal,
			Title:    "Standup",
			StartsAt: time.Date(2026, 1, 1, 9, 0, 0, 0, time.UTC),
			EndsAt:   time.Date(2026, 1, 1, 9, 30, 0, 0, time.UTC),
		},
	}, nil
}

// fakeGetterWithAuth also implements AuthHandler.
type fakeGetterWithAuth struct{ fakeGetter }

func (fakeGetterWithAuth) Metadata() pluginsdk.GetterMetadata {
	m := fakeGetter{}.Metadata()
	m.Name = "fake-getter-auth"
	m.SupportsAuth = true
	return m
}

func (fakeGetterWithAuth) Authenticate(params map[string]any, _ io.Writer) (map[string]string, error) {
	key, _ := params["key"].(string)
	return map[string]string{"token": "tok-" + key}, nil
}

type fakeSetter struct{}

func (fakeSetter) Metadata() pluginsdk.SetterMetadata {
	return pluginsdk.SetterMetadata{
		Name:       "fake-setter",
		ParamSpecs: []pluginsdk.ParamSpec{{Name: "status_message", Description: "Message", Required: true, Type: "string"}},
	}
}

func (fakeSetter) SetStatus(params map[string]any) (pluginsdk.SetResult, error) {
	msg, _ := params["status_message"].(string)
	return pluginsdk.SetResult{Response: "set: " + msg}, nil
}

// fakeSetterWithAuth also implements AuthHandler.
type fakeSetterWithAuth struct{ fakeSetter }

func (fakeSetterWithAuth) Authenticate(_ map[string]any, _ io.Writer) (map[string]string, error) {
	return map[string]string{"setter_token": "stok"}, nil
}

// --- ServeGetter tests ---

type serveGetterTest struct {
	TestName   string
	Args       []string
	Stdin      string
	WantStatus int
	CheckOut   func(t *testing.T, out string)
}

func (tc serveGetterTest) RunTest(t *testing.T) {
	t.Helper()
	var out bytes.Buffer
	in := strings.NewReader(tc.Stdin)
	got := pluginsdk.ServeGetter(fakeGetter{}, tc.Args, in, &out)
	assert.Equal(t, tc.WantStatus, got, "[%s] exit status", tc.TestName)
	if tc.CheckOut != nil {
		tc.CheckOut(t, out.String())
	}
}

func TestServeGetter(t *testing.T) {
	tests := []serveGetterTest{
		{
			TestName:   "metadata_flag",
			Args:       []string{"--metadata"},
			WantStatus: 0,
			CheckOut: func(t *testing.T, out string) {
				var resp struct {
					Name        string `json:"name"`
					Type        string `json:"type"`
					MinInterval string `json:"min_interval"`
				}
				require.NoError(t, json.Unmarshal([]byte(out), &resp))
				assert.Equal(t, "fake-getter", resp.Name)
				assert.Equal(t, "getter", resp.Type)
				assert.Equal(t, "30s", resp.MinInterval)
			},
		},
		{
			TestName:   "get_events_flag",
			Args:       []string{"--get-events"},
			Stdin:      `{"params":{"calendar_id":"primary"}}`,
			WantStatus: 0,
			CheckOut: func(t *testing.T, out string) {
				var resp struct {
					Events []struct {
						Calendar string `json:"calendar"`
						Title    string `json:"title"`
					} `json:"events"`
					Error string `json:"error,omitempty"`
				}
				require.NoError(t, json.Unmarshal([]byte(out), &resp))
				require.Len(t, resp.Events, 1)
				assert.Equal(t, "primary", resp.Events[0].Calendar)
				assert.Equal(t, "Standup", resp.Events[0].Title)
				assert.Empty(t, resp.Error)
			},
		},
		{
			TestName:   "unknown_flag_returns_nonzero",
			Args:       []string{"--unknown"},
			WantStatus: 1,
		},
		{
			TestName:   "no_args_returns_nonzero",
			Args:       []string{},
			WantStatus: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}

// --- authenticate tests (getter) ---

func TestServeGetterAuthenticate(t *testing.T) {
	t.Run("authenticate_with_auth_handler", func(t *testing.T) {
		var out bytes.Buffer
		in := strings.NewReader(`{"params":{"key":"mykey"}}`)
		code := pluginsdk.ServeGetter(fakeGetterWithAuth{}, []string{"--authenticate"}, in, &out)
		assert.Equal(t, 0, code)
		var resp struct {
			Tokens map[string]string `json:"tokens"`
			Error  string            `json:"error,omitempty"`
		}
		require.NoError(t, json.Unmarshal([]byte(out.String()), &resp))
		assert.Equal(t, "tok-mykey", resp.Tokens["token"])
		assert.Empty(t, resp.Error)
	})

	t.Run("authenticate_without_auth_handler_returns_error", func(t *testing.T) {
		var out bytes.Buffer
		in := strings.NewReader(`{"params":{}}`)
		code := pluginsdk.ServeGetter(fakeGetter{}, []string{"--authenticate"}, in, &out)
		assert.Equal(t, 1, code)
		var resp struct {
			Error string `json:"error"`
		}
		require.NoError(t, json.Unmarshal([]byte(out.String()), &resp))
		assert.NotEmpty(t, resp.Error)
	})

	t.Run("metadata_with_auth_handler_sets_supports_auth", func(t *testing.T) {
		var out bytes.Buffer
		code := pluginsdk.ServeGetter(fakeGetterWithAuth{}, []string{"--metadata"}, strings.NewReader(""), &out)
		assert.Equal(t, 0, code)
		var resp struct {
			SupportsAuth bool `json:"supports_auth"`
		}
		require.NoError(t, json.Unmarshal([]byte(out.String()), &resp))
		assert.True(t, resp.SupportsAuth)
	})

	t.Run("metadata_without_auth_handler_omits_supports_auth", func(t *testing.T) {
		var out bytes.Buffer
		code := pluginsdk.ServeGetter(fakeGetter{}, []string{"--metadata"}, strings.NewReader(""), &out)
		assert.Equal(t, 0, code)
		var resp struct {
			SupportsAuth bool `json:"supports_auth"`
		}
		require.NoError(t, json.Unmarshal([]byte(out.String()), &resp))
		assert.False(t, resp.SupportsAuth)
	})
}

// --- ServeSetter tests ---

type serveSetterTest struct {
	TestName   string
	Args       []string
	Stdin      string
	WantStatus int
	CheckOut   func(t *testing.T, out string)
}

func (tc serveSetterTest) RunTest(t *testing.T) {
	t.Helper()
	var out bytes.Buffer
	in := strings.NewReader(tc.Stdin)
	got := pluginsdk.ServeSetter(fakeSetter{}, tc.Args, in, &out)
	assert.Equal(t, tc.WantStatus, got, "[%s] exit status", tc.TestName)
	if tc.CheckOut != nil {
		tc.CheckOut(t, out.String())
	}
}

func TestServeSetter(t *testing.T) {
	tests := []serveSetterTest{
		{
			TestName:   "metadata_flag",
			Args:       []string{"--metadata"},
			WantStatus: 0,
			CheckOut: func(t *testing.T, out string) {
				var resp struct {
					Name string `json:"name"`
					Type string `json:"type"`
				}
				require.NoError(t, json.Unmarshal([]byte(out), &resp))
				assert.Equal(t, "fake-setter", resp.Name)
				assert.Equal(t, "setter", resp.Type)
			},
		},
		{
			TestName:   "set_status_flag",
			Args:       []string{"--set-status"},
			Stdin:      `{"params":{"status_message":"In a meeting"}}`,
			WantStatus: 0,
			CheckOut: func(t *testing.T, out string) {
				var resp struct {
					Skipped  bool   `json:"skipped"`
					Response string `json:"response"`
					Error    string `json:"error,omitempty"`
				}
				require.NoError(t, json.Unmarshal([]byte(out), &resp))
				assert.False(t, resp.Skipped)
				assert.Equal(t, "set: In a meeting", resp.Response)
				assert.Empty(t, resp.Error)
			},
		},
		{
			TestName:   "unknown_flag_returns_nonzero",
			Args:       []string{"--unknown"},
			WantStatus: 1,
		},
		{
			TestName:   "no_args_returns_nonzero",
			Args:       []string{},
			WantStatus: 1,
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}

// --- authenticate tests (setter) ---

func TestServeSetterAuthenticate(t *testing.T) {
	t.Run("authenticate_with_auth_handler", func(t *testing.T) {
		var out bytes.Buffer
		in := strings.NewReader(`{"params":{}}`)
		code := pluginsdk.ServeSetter(fakeSetterWithAuth{}, []string{"--authenticate"}, in, &out)
		assert.Equal(t, 0, code)
		var resp struct {
			Tokens map[string]string `json:"tokens"`
			Error  string            `json:"error,omitempty"`
		}
		require.NoError(t, json.Unmarshal([]byte(out.String()), &resp))
		assert.Equal(t, "stok", resp.Tokens["setter_token"])
	})

	t.Run("authenticate_without_auth_handler_returns_error", func(t *testing.T) {
		var out bytes.Buffer
		in := strings.NewReader(`{"params":{}}`)
		code := pluginsdk.ServeSetter(fakeSetter{}, []string{"--authenticate"}, in, &out)
		assert.Equal(t, 1, code)
	})
}
