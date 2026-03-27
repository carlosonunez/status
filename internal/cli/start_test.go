package cli_test

import (
	"bytes"
	"testing"

	"github.com/carlosonunez/status/internal/cli"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStart_MissingConfigReturnsHelpfulError(t *testing.T) {
	root := cli.NewRootCommand()
	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetErr(&buf)
	root.SetArgs([]string{"start", "--config=/nonexistent/status/config.yaml"})
	err := root.Execute()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "Status is not configured")
	assert.Contains(t, err.Error(), "status auth login")
}
