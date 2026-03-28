package cli

import (
	"fmt"
	"os"

	"github.com/carlosonunez/status/internal/config"
	"github.com/spf13/cobra"
)

func newCheckCommand() *cobra.Command {
	check := &cobra.Command{
		Use:   "check",
		Short: "Run checks against Status configuration and state",
	}
	check.AddCommand(newCheckConfigCommand())
	return check
}

func newCheckConfigCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "config",
		Short: "Validate the Status config file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := cfgPath
			if path == "" {
				path = config.DefaultPath()
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Checking config at %s...\n", path)

			cfg, err := config.Load(path)
			if err != nil {
				fmt.Fprintf(cmd.OutOrStdout(), "❌ %s\n", err)
				os.Exit(1)
			}

			errs := config.Validate(cfg)
			if len(errs) > 0 {
				fmt.Fprintf(cmd.OutOrStdout(), "❌ Config has %d error(s):\n", len(errs))
				for _, e := range errs {
					fmt.Fprintf(cmd.OutOrStdout(), "   • %s\n", e)
				}
				os.Exit(1)
			}

			fmt.Fprintf(cmd.OutOrStdout(), "✅ Config looks good!\n")
			return nil
		},
	}
}
