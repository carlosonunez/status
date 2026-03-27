package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/carlosonunez/status/internal/cli"
	"github.com/carlosonunez/status/internal/plugin"
	"github.com/sirupsen/logrus"

	// Register built-in event getters.
	_ "github.com/carlosonunez/status/internal/getter/dummy"

	// Register built-in status setters.
	_ "github.com/carlosonunez/status/internal/setter/dummy"
)

// Build metadata overridden at build time via:
//
//	-ldflags "-X main.version=v1.2.3 -X main.commit=abc1234 -X main.date=2026-03-27"
var (
	version = "NO_VERSION"
	commit  = "unknown"
	date    = "unknown"
)

func main() {
	binDir := filepath.Join(xdg.ConfigHome, "status", "bin")
	if err := plugin.DiscoverAllDefault(binDir); err != nil {
		// Non-fatal: log and continue so built-in plugins still work.
		logrus.WithError(err).Warn("plugin discovery failed")
	}

	root := cli.NewRootCommand(version, commit, date)
	cmd, err := root.ExecuteC()
	if err != nil {
		fmt.Fprintln(os.Stderr, cmd.UsageString())
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
