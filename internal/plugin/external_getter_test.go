package plugin_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/plugin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFakeGetterScript writes a shell script that behaves like a getter plugin.
// It responds to --metadata and --get-events.
func writeFakeGetterScript(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, "status-getter-"+name)
	script := `#!/usr/bin/env bash
set -e
case "$1" in
  --metadata)
    printf '{"name":"%s","type":"getter","min_interval":"1m","param_specs":[{"name":"cal","description":"Cal","required":true,"type":"string"}]}\n' "` + name + `"
    ;;
  --get-events)
    read -r input
    printf '{"events":[{"calendar":"test-cal","title":"Fake Event","starts_at":"2026-01-01T09:00:00Z","ends_at":"2026-01-01T09:30:00Z","all_day":false}]}\n'
    ;;
  *)
    exit 1
    ;;
esac
`
	require.NoError(t, os.WriteFile(path, []byte(script), 0o755))
	return path
}

type externalGetterTest struct {
	TestName string
	Setup    func(t *testing.T, binDir string) *plugin.ExternalGetter
	Run      func(t *testing.T, g *plugin.ExternalGetter)
}

func (tc externalGetterTest) RunTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	g := tc.Setup(t, dir)
	tc.Run(t, g)
}

func TestExternalGetter(t *testing.T) {
	tests := []externalGetterTest{
		{
			TestName: "name_from_binary",
			Setup: func(t *testing.T, dir string) *plugin.ExternalGetter {
				path := writeFakeGetterScript(t, dir, "mycal")
				g, err := plugin.NewExternalGetter(path)
				require.NoError(t, err)
				return g
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter) {
				assert.Equal(t, "mycal", g.Name())
			},
		},
		{
			TestName: "min_interval_parsed",
			Setup: func(t *testing.T, dir string) *plugin.ExternalGetter {
				path := writeFakeGetterScript(t, dir, "mycal")
				g, err := plugin.NewExternalGetter(path)
				require.NoError(t, err)
				return g
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter) {
				assert.Equal(t, time.Minute, g.MinInterval())
			},
		},
		{
			TestName: "param_specs_from_metadata",
			Setup: func(t *testing.T, dir string) *plugin.ExternalGetter {
				path := writeFakeGetterScript(t, dir, "mycal")
				g, err := plugin.NewExternalGetter(path)
				require.NoError(t, err)
				return g
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter) {
				specs := g.ParamSpecs()
				require.Len(t, specs, 1)
				assert.Equal(t, "cal", specs[0].Name)
			},
		},
		{
			TestName: "get_events_invokes_binary",
			Setup: func(t *testing.T, dir string) *plugin.ExternalGetter {
				path := writeFakeGetterScript(t, dir, "mycal")
				g, err := plugin.NewExternalGetter(path)
				require.NoError(t, err)
				return g
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter) {
				events, err := g.GetEvents(context.Background(), params.FromMap(map[string]any{"cal": "primary"}))
				require.NoError(t, err)
				require.Len(t, events, 1)
				assert.Equal(t, "test-cal", events[0].Calendar)
				assert.Equal(t, "Fake Event", events[0].Title)
			},
		},
		{
			TestName: "nonexistent_binary_returns_error",
			Setup: func(t *testing.T, dir string) *plugin.ExternalGetter {
				g, err := plugin.NewExternalGetter(filepath.Join(dir, "status-getter-nope"))
				require.Error(t, err)
				return g
			},
			Run: func(t *testing.T, g *plugin.ExternalGetter) {
				// setup already verified the error; g is nil
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
