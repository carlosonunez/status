package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/carlosonunez/status/internal/cli"
	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ── fakes ─────────────────────────────────────────────────────────────────────

type authGetter struct {
	name    string
	called  *[]string
	authErr error
}

func (a *authGetter) Name() string                       { return a.name }
func (a *authGetter) ParamSpecs() []pluginspec.ParamSpec { return nil }
func (a *authGetter) MinInterval() time.Duration         { return time.Minute }
func (a *authGetter) GetEvents(_ context.Context, _ params.Params) ([]getter.Event, error) {
	return nil, nil
}
func (a *authGetter) Authenticate(_ context.Context, _ pluginspec.TokenStore) error {
	*a.called = append(*a.called, a.name)
	return a.authErr
}

type authSetter struct {
	name    string
	called  *[]string
	authErr error
}

func (a *authSetter) Name() string                       { return a.name }
func (a *authSetter) ParamSpecs() []pluginspec.ParamSpec { return nil }
func (a *authSetter) SetStatus(_ context.Context, _ setter.Status) (setter.SetResult, error) {
	return setter.SetResult{}, nil
}
func (a *authSetter) Authenticate(_ context.Context, _ pluginspec.TokenStore) error {
	*a.called = append(*a.called, a.name)
	return a.authErr
}

type fakeTokenStore struct{}

func (f *fakeTokenStore) Get(service, key string) (string, error) {
	return "", fmt.Errorf("not found")
}
func (f *fakeTokenStore) Set(service, key, value string) error { return nil }
func (f *fakeTokenStore) Delete(service, key string) error     { return nil }

// ── helpers ───────────────────────────────────────────────────────────────────

func runAuthCmd(
	t *testing.T,
	gr *getter.Registry,
	sr *setter.Registry,
	store pluginspec.TokenStore,
	args []string,
) (string, error) {
	t.Helper()
	cmd := cli.NewAuthCommandForTest(gr, sr, store)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	root := &cobra.Command{Use: "status"}
	root.AddCommand(cmd)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(append([]string{"auth"}, args...))
	err := root.Execute()
	return buf.String(), err
}

// ── tests ─────────────────────────────────────────────────────────────────────

type authLoginTest struct {
	TestName         string
	BuildGetters     func(called *[]string) []*authGetter
	BuildSetters     func(called *[]string) []*authSetter
	Args             []string
	WantCalled       []string
	WantOutputContains []string
	WantErr          bool
}

func (tc authLoginTest) RunTest(t *testing.T) {
	t.Helper()
	var called []string
	gr := getter.NewRegistry()
	sr := setter.NewRegistry()
	if tc.BuildGetters != nil {
		for _, g := range tc.BuildGetters(&called) {
			gr.Register(g)
		}
	}
	if tc.BuildSetters != nil {
		for _, s := range tc.BuildSetters(&called) {
			sr.Register(s)
		}
	}
	store := &fakeTokenStore{}
	out, err := runAuthCmd(t, gr, sr, store, tc.Args)
	if tc.WantErr {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
	assert.ElementsMatch(t, tc.WantCalled, called)
	for _, want := range tc.WantOutputContains {
		assert.Contains(t, out, want)
	}
}

func TestAuthLogin(t *testing.T) {
	tests := []authLoginTest{
		{
			TestName: "no_authenticators_prints_message",
			BuildGetters: func(called *[]string) []*authGetter { return nil },
			Args:         []string{"login", "--all"},
			WantCalled:   nil,
			WantOutputContains: []string{"No integrations require authentication"},
		},
		{
			TestName: "all_flag_calls_all_authenticators",
			BuildGetters: func(called *[]string) []*authGetter {
				return []*authGetter{
					{name: "getter_a", called: called},
				}
			},
			BuildSetters: func(called *[]string) []*authSetter {
				return []*authSetter{
					{name: "setter_b", called: called},
				}
			},
			Args:       []string{"login", "--all"},
			WantCalled: []string{"getter_a", "setter_b"},
		},
		{
			TestName: "integrations_flag_calls_only_named",
			BuildGetters: func(called *[]string) []*authGetter {
				return []*authGetter{
					{name: "getter_a", called: called},
					{name: "getter_b", called: called},
				}
			},
			Args:       []string{"login", "--integrations=getter_a"},
			WantCalled: []string{"getter_a"},
		},
		{
			TestName: "unknown_integration_returns_error",
			Args:     []string{"login", "--integrations=nonexistent"},
			WantErr:  true,
		},
		{
			TestName: "authenticate_error_is_propagated",
			BuildGetters: func(called *[]string) []*authGetter {
				return []*authGetter{
					{name: "bad_getter", called: called, authErr: fmt.Errorf("auth failed")},
				}
			},
			Args:    []string{"login", "--all"},
			WantErr: true,
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}
