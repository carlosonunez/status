package main

import (
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/carlosonunez/status/internal/cli"
	"github.com/carlosonunez/status/internal/plugin"

	// Register built-in event getters.
	_ "github.com/carlosonunez/status/internal/getter/dummy"

	// Register built-in status setters.
	_ "github.com/carlosonunez/status/internal/setter/dummy"
)

func main() {
	binDir := filepath.Join(xdg.ConfigHome, "status", "bin")
	if err := plugin.DiscoverAllDefault(binDir); err != nil {
		// Non-fatal: log and continue so built-in plugins still work.
		_, _ = os.Stderr.Write([]byte("status: plugin discovery: " + err.Error() + "\n"))
	}

	root := cli.NewRootCommand()
	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
