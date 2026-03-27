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
	Commit   string
	Date     string
	WantOut  string
}

func (tc versionTest) RunTest(t *testing.T) {
	t.Helper()
	root := cli.NewRootCommand(tc.Version, tc.Commit, tc.Date)
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
			TestName: "shows_full_build_info",
			Version:  "v1.2.3",
			Commit:   "abc1234",
			Date:     "2026-03-27",
			WantOut:  "status version v1.2.3 (Commit: abc1234, Last Updated: 2026-03-27)",
		},
		{
			TestName: "shows_dev_defaults",
			Version:  "dev",
			Commit:   "unknown",
			Date:     "unknown",
			WantOut:  "status version dev (Commit: unknown, Last Updated: unknown)",
		},
	}
	for _, tc := range tests {
		t.Run(tc.TestName, func(t *testing.T) {
			tc.RunTest(t)
		})
	}
}
