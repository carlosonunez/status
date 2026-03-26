package plugin_test

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/carlosonunez/status/internal/params"
	"github.com/carlosonunez/status/internal/plugin"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeFakeSetterScript writes a shell script that behaves like a setter plugin.
func writeFakeSetterScript(t *testing.T, dir, name string) string {
	t.Helper()
	path := filepath.Join(dir, "status-setter-"+name)
	script := `#!/usr/bin/env bash
set -e
case "$1" in
  --metadata)
    printf '{"name":"%s","type":"setter","param_specs":[{"name":"status_message","description":"Message","required":true,"type":"string"}]}\n' "` + name + `"
    ;;
  --set-status)
    read -r input
    printf '{"skipped":false,"response":"ok"}\n'
    ;;
  *)
    exit 1
    ;;
esac
`
	require.NoError(t, os.WriteFile(path, []byte(script), 0o755))
	return path
}

type externalSetterTest struct {
	TestName string
	Setup    func(t *testing.T, binDir string) *plugin.ExternalSetter
	Run      func(t *testing.T, s *plugin.ExternalSetter)
}

func (tc externalSetterTest) RunTest(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	s := tc.Setup(t, dir)
	tc.Run(t, s)
}

func TestExternalSetter(t *testing.T) {
	tests := []externalSetterTest{
		{
			TestName: "name_from_binary",
			Setup: func(t *testing.T, dir string) *plugin.ExternalSetter {
				path := writeFakeSetterScript(t, dir, "myslack")
				s, err := plugin.NewExternalSetter(path)
				require.NoError(t, err)
				return s
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter) {
				assert.Equal(t, "myslack", s.Name())
			},
		},
		{
			TestName: "param_specs_from_metadata",
			Setup: func(t *testing.T, dir string) *plugin.ExternalSetter {
				path := writeFakeSetterScript(t, dir, "myslack")
				s, err := plugin.NewExternalSetter(path)
				require.NoError(t, err)
				return s
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter) {
				specs := s.ParamSpecs()
				require.Len(t, specs, 1)
				assert.Equal(t, "status_message", specs[0].Name)
			},
		},
		{
			TestName: "set_status_invokes_binary",
			Setup: func(t *testing.T, dir string) *plugin.ExternalSetter {
				path := writeFakeSetterScript(t, dir, "myslack")
				s, err := plugin.NewExternalSetter(path)
				require.NoError(t, err)
				return s
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter) {
				st := setter.Status{Params: params.FromMap(map[string]any{"status_message": "In a meeting"})}
				result, err := s.SetStatus(context.Background(), st)
				require.NoError(t, err)
				assert.False(t, result.Skipped)
				assert.Equal(t, "ok", result.Response)
			},
		},
		{
			TestName: "nonexistent_binary_returns_error",
			Setup: func(t *testing.T, dir string) *plugin.ExternalSetter {
				s, err := plugin.NewExternalSetter(filepath.Join(dir, "status-setter-nope"))
				require.Error(t, err)
				return s
			},
			Run: func(t *testing.T, s *plugin.ExternalSetter) {
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
