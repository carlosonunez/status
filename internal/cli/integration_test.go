package cli_test

import (
	"bytes"
	"context"
	"strings"
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

type fakeIntGetter struct {
	name   string
	specs  []pluginspec.ParamSpec
	desc   string
	source string
	auth   []pluginspec.AuthParam
	notes  []string
}

func (f *fakeIntGetter) Name() string                        { return f.name }
func (f *fakeIntGetter) ParamSpecs() []pluginspec.ParamSpec  { return f.specs }
func (f *fakeIntGetter) MinInterval() time.Duration          { return time.Minute }
func (f *fakeIntGetter) GetEvents(_ context.Context, _ params.Params) ([]getter.Event, error) {
	return nil, nil
}
func (f *fakeIntGetter) Description() string          { return f.desc }
func (f *fakeIntGetter) Source() string               { return f.source }
func (f *fakeIntGetter) AuthParams() []pluginspec.AuthParam { return f.auth }
func (f *fakeIntGetter) AuthNotes() []string          { return f.notes }

type fakeIntSetter struct {
	name   string
	specs  []pluginspec.ParamSpec
	desc   string
	source string
	auth   []pluginspec.AuthParam
	notes  []string
}

func (f *fakeIntSetter) Name() string                        { return f.name }
func (f *fakeIntSetter) ParamSpecs() []pluginspec.ParamSpec  { return f.specs }
func (f *fakeIntSetter) SetStatus(_ context.Context, _ setter.Status) (setter.SetResult, error) {
	return setter.SetResult{}, nil
}
func (f *fakeIntSetter) Description() string          { return f.desc }
func (f *fakeIntSetter) Source() string               { return f.source }
func (f *fakeIntSetter) AuthParams() []pluginspec.AuthParam { return f.auth }
func (f *fakeIntSetter) AuthNotes() []string          { return f.notes }

// ── helpers ───────────────────────────────────────────────────────────────────

func runIntegrationCmd(
	t *testing.T,
	gr *getter.Registry,
	sr *setter.Registry,
	args []string,
) (string, error) {
	t.Helper()
	cmd := cli.NewIntegrationCommandForTest(gr, sr)
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	// Suppress PersistentPreRunE (logging setup) by wrapping in a root-like parent.
	root := &cobra.Command{Use: "status"}
	root.AddCommand(cmd)
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(append([]string{"integration"}, args...))
	err := root.Execute()
	return buf.String(), err
}

func newTestRegistries(getters []*fakeIntGetter, setters []*fakeIntSetter) (*getter.Registry, *setter.Registry) {
	gr := getter.NewRegistry()
	sr := setter.NewRegistry()
	for _, g := range getters {
		gr.Register(g)
	}
	for _, s := range setters {
		sr.Register(s)
	}
	return gr, sr
}

// ── integration list tests ────────────────────────────────────────────────────

type integrationListTest struct {
	TestName        string
	Getters         []*fakeIntGetter
	Setters         []*fakeIntSetter
	Args            []string
	WantContains    []string
	WantNotContains []string
}

func (tc integrationListTest) RunTest(t *testing.T) {
	t.Helper()
	gr, sr := newTestRegistries(tc.Getters, tc.Setters)
	out, err := runIntegrationCmd(t, gr, sr, append([]string{"list"}, tc.Args...))
	require.NoError(t, err)
	for _, want := range tc.WantContains {
		assert.Contains(t, out, want, "expected output to contain %q", want)
	}
	for _, notWant := range tc.WantNotContains {
		assert.NotContains(t, out, notWant, "expected output NOT to contain %q", notWant)
	}
}

func TestIntegrationList(t *testing.T) {
	tests := []integrationListTest{
		{
			TestName: "shows_event_getters",
			Getters:  []*fakeIntGetter{{name: "my_getter"}},
			WantContains: []string{"event-getter", "my_getter", "builtin"},
		},
		{
			TestName: "shows_status_setters",
			Setters:  []*fakeIntSetter{{name: "my_setter"}},
			WantContains: []string{"status-setter", "my_setter", "builtin"},
		},
		{
			TestName: "filter_event_getter_hides_setters",
			Getters:  []*fakeIntGetter{{name: "my_getter"}},
			Setters:  []*fakeIntSetter{{name: "my_setter"}},
			Args:     []string{"--type=event-getter"},
			WantContains:    []string{"my_getter"},
			WantNotContains: []string{"my_setter"},
		},
		{
			TestName: "filter_status_setter_hides_getters",
			Getters:  []*fakeIntGetter{{name: "my_getter"}},
			Setters:  []*fakeIntSetter{{name: "my_setter"}},
			Args:     []string{"--type=status-setter"},
			WantContains:    []string{"my_setter"},
			WantNotContains: []string{"my_getter"},
		},
		{
			TestName: "shows_param_names",
			Getters: []*fakeIntGetter{{
				name:  "my_getter",
				specs: []pluginspec.ParamSpec{{Name: "event_name"}, {Name: "max_duration"}},
			}},
			WantContains: []string{"event_name", "max_duration"},
		},
		{
			TestName: "no_params_shows_none",
			Getters:  []*fakeIntGetter{{name: "my_getter"}},
			WantContains: []string{"(none)"},
		},
		{
			TestName: "shows_external_source",
			Getters:  []*fakeIntGetter{{name: "my_getter", source: "external"}},
			WantContains: []string{"external"},
		},
		{
			TestName: "json_format_contains_fields",
			Getters:  []*fakeIntGetter{{name: "my_getter"}},
			Args:     []string{"--format=json"},
			WantContains: []string{`"type"`, `"name"`, `"my_getter"`},
		},
		{
			TestName: "yaml_format_contains_fields",
			Getters:  []*fakeIntGetter{{name: "my_getter"}},
			Args:     []string{"--format=yaml"},
			WantContains: []string{"type:", "name:", "my_getter"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}

// ── integration show tests ────────────────────────────────────────────────────

type integrationShowTest struct {
	TestName        string
	Getters         []*fakeIntGetter
	Setters         []*fakeIntSetter
	Args            []string
	WantContains    []string
	WantNotContains []string
	WantErr         bool
}

func (tc integrationShowTest) RunTest(t *testing.T) {
	t.Helper()
	gr, sr := newTestRegistries(tc.Getters, tc.Setters)
	out, err := runIntegrationCmd(t, gr, sr, append([]string{"show"}, tc.Args...))
	if tc.WantErr {
		require.Error(t, err)
		return
	}
	require.NoError(t, err)
	for _, want := range tc.WantContains {
		assert.Contains(t, out, want, "expected output to contain %q", want)
	}
	for _, notWant := range tc.WantNotContains {
		assert.NotContains(t, out, notWant, "expected output NOT to contain %q", notWant)
	}
}

func TestIntegrationShow(t *testing.T) {
	tests := []integrationShowTest{
		{
			TestName: "shows_getter_by_name",
			Getters:  []*fakeIntGetter{{name: "my_getter"}},
			Args:     []string{"--name=my_getter"},
			WantContains: []string{"NAME: my_getter", "TYPE: event-getter"},
		},
		{
			TestName: "shows_setter_by_name",
			Setters:  []*fakeIntSetter{{name: "my_setter"}},
			Args:     []string{"--name=my_setter"},
			WantContains: []string{"NAME: my_setter", "TYPE: status-setter"},
		},
		{
			TestName: "shows_description",
			Getters:  []*fakeIntGetter{{name: "my_getter", desc: "Gets things."}},
			Args:     []string{"--name=my_getter"},
			WantContains: []string{"DESCRIPTION: Gets things."},
		},
		{
			TestName: "shows_params",
			Getters: []*fakeIntGetter{{
				name:  "my_getter",
				specs: []pluginspec.ParamSpec{{Name: "event_name", Type: "string", Required: true, Description: "The event name."}},
			}},
			Args:         []string{"--name=my_getter"},
			WantContains: []string{"event_name", "string", "The event name."},
		},
		{
			TestName: "shows_auth_params",
			Getters: []*fakeIntGetter{{
				name: "my_getter",
				auth: []pluginspec.AuthParam{{Name: "client_id", Required: true, EnvVar: "MY_CLIENT_ID", Description: "The client ID."}},
			}},
			Args:         []string{"--name=my_getter"},
			WantContains: []string{"AUTHENTICATION PARAMETERS", "client_id", "MY_CLIENT_ID"},
		},
		{
			TestName: "shows_auth_notes",
			Getters: []*fakeIntGetter{{
				name:  "my_getter",
				notes: []string{"You need admin access."},
			}},
			Args:         []string{"--name=my_getter"},
			WantContains: []string{"AUTHENTICATION NOTES", "You need admin access."},
		},
		{
			TestName: "all_flag_shows_everything",
			Getters:  []*fakeIntGetter{{name: "getter_a"}},
			Setters:  []*fakeIntSetter{{name: "setter_b"}},
			Args:     []string{"--all"},
			WantContains: []string{"getter_a", "setter_b"},
		},
		{
			TestName: "multiple_names",
			Getters:  []*fakeIntGetter{{name: "getter_a"}, {name: "getter_b"}},
			Args:     []string{"--name=getter_a,getter_b"},
			WantContains: []string{"getter_a", "getter_b"},
		},
		{
			TestName: "unknown_name_returns_error",
			Args:     []string{"--name=nonexistent"},
			WantErr:  true,
		},
		{
			TestName: "json_format",
			Getters:  []*fakeIntGetter{{name: "my_getter", desc: "A getter."}},
			Args:     []string{"--name=my_getter", "--format=json"},
			WantContains: []string{`"name"`, `"my_getter"`, `"description"`},
		},
		{
			TestName: "yaml_format",
			Getters:  []*fakeIntGetter{{name: "my_getter"}},
			Args:     []string{"--name=my_getter", "--format=yaml"},
			WantContains: []string{"name:", "my_getter", "type:"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}

// ── helpers shared with auth tests ───────────────────────────────────────────

// plainGetter is a minimal getter with no optional interfaces.
type plainGetter struct{ name string }

func (p *plainGetter) Name() string                        { return p.name }
func (p *plainGetter) ParamSpecs() []pluginspec.ParamSpec  { return nil }
func (p *plainGetter) MinInterval() time.Duration          { return time.Minute }
func (p *plainGetter) GetEvents(_ context.Context, _ params.Params) ([]getter.Event, error) {
	return nil, nil
}

// trimLines removes leading/trailing whitespace from each line (useful for
// loose output matching).
func trimLines(s string) string {
	var b strings.Builder
	for _, line := range strings.Split(s, "\n") {
		b.WriteString(strings.TrimSpace(line))
		b.WriteByte('\n')
	}
	return b.String()
}

var _ = trimLines // used in auth_test.go
