package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newVersionCommand(version, commit, date string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the status version",
		Run: func(cmd *cobra.Command, _ []string) {
			fmt.Fprintf(cmd.OutOrStdout(),
				"status version %s (Commit: %s, Last Updated: %s)\n",
				version, commit, date)
		},
	}
}
