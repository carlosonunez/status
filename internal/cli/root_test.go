package cli_test

import (
	"bytes"
	"testing"

	"github.com/carlosonunez/status/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type pluginDirTest struct {
	TestName    string
	Args        []string
	WantDirs    []string
}

func (tc pluginDirTest) RunTest(t *testing.T) {
	t.Helper()
	var gotDirs []string
	restore := cli.SetPluginDiscoverer(func(dirs []string) error {
		gotDirs = dirs
		return nil
	})
	defer restore()

	root := cli.NewRootCommand("dev", "unknown", "unknown")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs(tc.Args)
	// Use Execute (not ExecuteC) so errors don't print usage.
	_ = root.Execute()

	assert.Equal(t, tc.WantDirs, gotDirs)
}

func TestPluginDirFlag(t *testing.T) {
	tests := []pluginDirTest{
		{
			TestName: "no_plugin_dir_passes_empty_slice",
			Args:     []string{"integration", "list"},
			WantDirs: nil,
		},
		{
			TestName: "single_plugin_dir_is_passed",
			Args:     []string{"--plugin-dir", "/opt/status/plugins", "integration", "list"},
			WantDirs: []string{"/opt/status/plugins"},
		},
		{
			TestName: "multiple_plugin_dirs_are_all_passed",
			Args: []string{
				"--plugin-dir", "/opt/status/plugins",
				"--plugin-dir", "/usr/local/lib/status",
				"integration", "list",
			},
			WantDirs: []string{"/opt/status/plugins", "/usr/local/lib/status"},
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}

func TestPluginDirWithVersionCommand(t *testing.T) {
	// Verify that PersistentPreRunE (and therefore discovery) runs for
	// non-integration subcommands too.
	var called bool
	restore := cli.SetPluginDiscoverer(func(_ []string) error {
		called = true
		return nil
	})
	defer restore()

	root := cli.NewRootCommand("dev", "unknown", "unknown")
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"version"})
	require.NoError(t, root.Execute())
	assert.True(t, called, "plugin discoverer should run for every subcommand")
}
