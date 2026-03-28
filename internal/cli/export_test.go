package cli

import (
	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/pluginspec"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/spf13/cobra"
)

// NewIntegrationCommandForTest exposes the integration command with injected
// registries for use in tests.
func NewIntegrationCommandForTest(gr *getter.Registry, sr *setter.Registry) *cobra.Command {
	return newIntegrationCommand(gr, sr)
}

// NewAuthCommandForTest exposes the auth command with injected registries and
// token store for use in tests.
func NewAuthCommandForTest(gr *getter.Registry, sr *setter.Registry, store pluginspec.TokenStore) *cobra.Command {
	return newAuthCommandWithStore(gr, sr, store)
}
