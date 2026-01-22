package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/config"
	"phasionary/internal/data"
)

func newInitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the data directory with a default project",
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := config.ResolveDataDir(viper.GetString("data"))
			if err != nil {
				return err
			}
			store := data.NewStore(dataDir)
			project, err := store.InitDefault()
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Initialized data directory: %s\n", dataDir)
			fmt.Fprintf(cmd.OutOrStdout(), "Default project: %s (%s)\n", project.Name, project.ID)
			return nil
		},
	}
	return cmd
}
