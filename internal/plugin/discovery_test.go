package plugin_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/plugin"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type discoveryTest struct {
	TestName string
	Setup    func(t *testing.T, dir string)
	Run      func(t *testing.T, gr *getter.Registry, sr *setter.Registry)
}

func (tc discoveryTest) RunTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	tc.Setup(t, dir)
	gr := getter.NewRegistry()
	sr := setter.NewRegistry()
	require.NoError(t, plugin.DiscoverAll(dir, gr, sr))
	tc.Run(t, gr, sr)
}

func TestDiscoverAll(t *testing.T) {
	tests := []discoveryTest{
		{
			TestName: "discovers_getter_binary",
			Setup: func(t *testing.T, dir string) {
				writeFakeGetterScript(t, dir, "testcal")
			},
			Run: func(t *testing.T, gr *getter.Registry, sr *setter.Registry) {
				g, err := gr.Get("testcal")
				require.NoError(t, err)
				assert.Equal(t, "testcal", g.Name())
				assert.Empty(t, sr.All())
			},
		},
		{
			TestName: "discovers_setter_binary",
			Setup: func(t *testing.T, dir string) {
				writeFakeSetterScript(t, dir, "testslack")
			},
			Run: func(t *testing.T, gr *getter.Registry, sr *setter.Registry) {
				s, err := sr.Get("testslack")
				require.NoError(t, err)
				assert.Equal(t, "testslack", s.Name())
				assert.Empty(t, gr.All())
			},
		},
		{
			TestName: "discovers_both_getter_and_setter",
			Setup: func(t *testing.T, dir string) {
				writeFakeGetterScript(t, dir, "testcal")
				writeFakeSetterScript(t, dir, "testslack")
			},
			Run: func(t *testing.T, gr *getter.Registry, sr *setter.Registry) {
				require.Len(t, gr.All(), 1)
				require.Len(t, sr.All(), 1)
			},
		},
		{
			TestName: "ignores_non_plugin_files",
			Setup: func(t *testing.T, dir string) {
				// a file that doesn't match either prefix
				require.NoError(t, os.WriteFile(filepath.Join(dir, "something-else"), []byte(""), 0o755))
			},
			Run: func(t *testing.T, gr *getter.Registry, sr *setter.Registry) {
				assert.Empty(t, gr.All())
				assert.Empty(t, sr.All())
			},
		},
		{
			TestName: "empty_dir_registers_nothing",
			Setup:    func(t *testing.T, dir string) {},
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
