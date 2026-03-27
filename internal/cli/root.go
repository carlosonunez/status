package cli

import (
	"strings"
	"time"

	"github.com/carlosonunez/status/internal/getter"
	"github.com/carlosonunez/status/internal/setter"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	logLevel  string
	logFormat string
	cfgPath   string
)

// NewRootCommand builds and returns the root cobra command.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "status",
		Short: "Keep your status in sync with your whereabouts",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			return configureLogging(logLevel, logFormat)
		},
	}

	root.PersistentFlags().StringVar(&logLevel, "log-level", "info",
		"Log verbosity: trace, debug, info, warn, error")
	root.PersistentFlags().StringVar(&logFormat, "log-format", "text",
		"Log output format: text or json")
	root.PersistentFlags().StringVar(&cfgPath, "config", "",
		"Path to config file (default: ~/.config/status/config.yaml)")

	root.AddCommand(newCheckCommand())
	root.AddCommand(newStartCommand())
	root.AddCommand(newIntegrationCommand(getter.DefaultRegistry(), setter.DefaultRegistry()))
	root.AddCommand(newAuthCommand(getter.DefaultRegistry(), setter.DefaultRegistry()))

	return root
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
