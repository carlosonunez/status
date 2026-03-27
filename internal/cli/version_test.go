package cli_test

import (
	"bytes"
	"testing"

	"github.com/carlosonunez/status/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type versionTest struct {
	TestName string
	Version  string
	WantOut  string
}

func (tc versionTest) RunTest(t *testing.T) {
	t.Helper()
	root := cli.NewRootCommand(tc.Version)
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"version"})
	err := root.Execute()
	require.NoError(t, err)
	assert.Contains(t, buf.String(), tc.WantOut)
}

func TestVersion(t *testing.T) {
	tests := []versionTest{
		{
			TestName: "shows_injected_version",
			Version:  "v1.2.3",
			WantOut:  "v1.2.3",
		},
		{
			TestName: "shows_dev_when_unset",
			Version:  "dev",
			WantOut:  "dev",
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}
