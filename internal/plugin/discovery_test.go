package plugin_test

import (
	"testing"
	"testing/fstest"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/plugin"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeGetterFactory returns a factory that creates ExternalGetters directly
// from the binary name without any filesystem or subprocess interaction.
func fakeGetterFactory(inv plugin.InvokerFunc) func(string) (*plugin.ExternalGetter, error) {
	return func(path string) (*plugin.ExternalGetter, error) {
		return plugin.NewExternalGetterWithInvoker(path, inv)
	}
}

// fakeSetterFactory returns a factory that creates ExternalSetters directly
// from the binary name without any filesystem or subprocess interaction.
func fakeSetterFactory(inv plugin.InvokerFunc) func(string) (*plugin.ExternalSetter, error) {
	return func(path string) (*plugin.ExternalSetter, error) {
		return plugin.NewExternalSetterWithInvoker(path, inv)
	}
}

type discoveryTest struct {
	TestName   string
	FS         fstest.MapFS
	GetterInv  plugin.InvokerFunc
	SetterInv  plugin.InvokerFunc
	Run        func(t *testing.T, gr *getter.Registry, sr *setter.Registry)
}

func (tc discoveryTest) RunTest(t *testing.T) {
	t.Helper()
	gr := getter.NewRegistry()
	sr := setter.NewRegistry()
	gf := fakeGetterFactory(tc.GetterInv)
	sf := fakeSetterFactory(tc.SetterInv)
	require.NoError(t, plugin.DiscoverFS(tc.FS, "/fake/bin", gr, sr, gf, sf))
	tc.Run(t, gr, sr)
}

func noopGetterInv() plugin.InvokerFunc {
	return metadataInvoker(
		`{"name":"testcal","type":"getter","min_interval":"1m"}`,
		`{"events":[]}`,
	)
}

func noopSetterInv() plugin.InvokerFunc {
	return setterInvoker(
		`{"name":"testslack","type":"setter"}`,
		`{"skipped":false,"response":"ok"}`,
	)
}

func TestDiscoverAll(t *testing.T) {
	tests := []discoveryTest{
		{
			TestName:  "discovers_getter_binary",
			FS:        fstest.MapFS{"status-getter-testcal": &fstest.MapFile{Mode: 0o755}},
			GetterInv: noopGetterInv(),
			SetterInv: noopSetterInv(),
			Run: func(t *testing.T, gr *getter.Registry, sr *setter.Registry) {
				g, err := gr.Get("testcal")
				require.NoError(t, err)
				assert.Equal(t, "testcal", g.Name())
				assert.Empty(t, sr.All())
			},
		},
		{
			TestName:  "discovers_setter_binary",
			FS:        fstest.MapFS{"status-setter-testslack": &fstest.MapFile{Mode: 0o755}},
			GetterInv: noopGetterInv(),
			SetterInv: noopSetterInv(),
			Run: func(t *testing.T, gr *getter.Registry, sr *setter.Registry) {
				s, err := sr.Get("testslack")
				require.NoError(t, err)
				assert.Equal(t, "testslack", s.Name())
				assert.Empty(t, gr.All())
			},
		},
		{
			TestName: "discovers_both_getter_and_setter",
			FS: fstest.MapFS{
				"status-getter-testcal":   &fstest.MapFile{Mode: 0o755},
				"status-setter-testslack": &fstest.MapFile{Mode: 0o755},
			},
			GetterInv: noopGetterInv(),
			SetterInv: noopSetterInv(),
			Run: func(t *testing.T, gr *getter.Registry, sr *setter.Registry) {
				assert.Len(t, gr.All(), 1)
				assert.Len(t, sr.All(), 1)
			},
		},
		{
			TestName:  "ignores_non_plugin_files",
			FS:        fstest.MapFS{"something-else": &fstest.MapFile{Mode: 0o755}},
			GetterInv: noopGetterInv(),
			SetterInv: noopSetterInv(),
			Run: func(t *testing.T, gr *getter.Registry, sr *setter.Registry) {
				assert.Empty(t, gr.All())
				assert.Empty(t, sr.All())
			},
		},
		{
			TestName:  "empty_fs_registers_nothing",
			FS:        fstest.MapFS{},
			GetterInv: noopGetterInv(),
			SetterInv: noopSetterInv(),
			Run: func(t *testing.T, gr *getter.Registry, sr *setter.Registry) {
				assert.Empty(t, gr.All())
				assert.Empty(t, sr.All())
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}
