package cli

import (
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/plugin"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logLevel   string
	logFormat  string
	cfgPath    string
	pluginDirs []string
)

// pluginDiscoverer is the function used to discover and register external
// plugins. It is a variable so tests can override it without touching the
// filesystem or spawning subprocesses.
var pluginDiscoverer = defaultDiscoverPlugins

// NewRootCommand builds and returns the root cobra command.
// version, commit, and date are injected at build time via ldflags; pass
// "dev"/"unknown"/"unknown" when building outside of a release pipeline.
func NewRootCommand(version, commit, date string) *cobra.Command {
	root := &cobra.Command{
		Use:   "status",
		Short: "Keep your status in sync with your whereabouts",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := configureLogging(logLevel, logFormat); err != nil {
				return err
			}
			return pluginDiscoverer(pluginDirs)
		},
	}

	root.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Log verbosity: trace, debug, info, warn, error")
	root.PersistentFlags().StringVar(&logFormat, "log-format", "text",
		"Log output format: text or json")
	root.PersistentFlags().StringVar(&cfgPath, "config", "",
		"Path to config file (default: ~/.config/status/config.yaml)")
	root.PersistentFlags().StringArrayVar(&pluginDirs, "plugin-dir", nil,
		"Additional directory to scan for plugin binaries (may be given more than once)")

	root.SilenceErrors = true
	root.SilenceUsage = true

	root.AddCommand(newVersionCommand(version, commit, date))
	root.AddCommand(newCheckCommand())
	root.AddCommand(newStartCommand())
	root.AddCommand(newIntegrationCommand(getter.DefaultRegistry(), setter.DefaultRegistry()))
	root.AddCommand(newAuthCommand(getter.DefaultRegistry(), setter.DefaultRegistry()))

	return root
}

// defaultDiscoverPlugins scans the default plugin directory plus any extra
// directories provided via --plugin-dir and registers found binaries.
func defaultDiscoverPlugins(extraDirs []string) error {
	defaultDir := filepath.Join(xdg.ConfigHome, "status", "bin")
	dirs := append([]string{defaultDir}, extraDirs...)
	for _, dir := range dirs {
		if err := plugin.DiscoverAllDefault(dir); err != nil {
			logrus.WithError(err).Warnf("plugin discovery failed for %q", dir)
		}
	}
	return nil
}

func configureLogging(level, format string) error {
	lvl, err := logrus.ParseLevel(strings.ToLower(level))
	if err != nil {
		return err
	}
	logrus.SetLevel(lvl)

	if strings.ToLower(format) == "json" {
		logrus.SetFormatter(&logrus.JSONFormatter{TimestampFormat: time.RFC3339Nano})
	} else {
		logrus.SetFormatter(&logrus.TextFormatter{
			FullTimestamp:   true,
			TimestampFormat: time.RFC3339,
		})
	}

	return nil
}
