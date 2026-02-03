package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"phasionary/internal/version"
)

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Aliases: []string{"v"},
		Short:   "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintf(cmd.OutOrStdout(), "phasionary %s (commit: %s, built: %s)\n",
				version.Version, version.Commit, version.BuildDate)
		},
	}
}
