package cli

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
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

	return root
}

func configureLogging(level, format string) error {
	lvl, err := zerolog.ParseLevel(strings.ToLower(level))
	if err != nil {
		return err
	}
	zerolog.SetGlobalLevel(lvl)
	zerolog.TimeFieldFormat = time.RFC3339Nano

	if strings.ToLower(format) == "json" {
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
	} else {
		log.Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).With().Timestamp().Logger()
	}

	return nil
}
