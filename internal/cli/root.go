package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/app"
	"phasionary/internal/config"
)

func Execute() error {
	return newRootCmd().Execute()
}

func newRootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "phasionary",
		Short: "Terminal-first project planning tool",
		RunE: func(cmd *cobra.Command, args []string) error {
			dataDir, err := config.ResolveDataDir(viper.GetString("data"))
			if err != nil {
				return err
			}
			return app.Run(dataDir, viper.GetString("project"))
		},
	}

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	cmd.PersistentFlags().String("data", "", "override data directory path")
	cmd.PersistentFlags().String("project", "", "target project for commands")

	viper.AutomaticEnv()
	if err := viper.BindEnv("data", config.EnvDataPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := viper.BindPFlag("data", cmd.PersistentFlags().Lookup("data")); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := viper.BindPFlag("project", cmd.PersistentFlags().Lookup("project")); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	cmd.AddCommand(newInitCmd())
	cmd.AddCommand(newProjectCmd())
	cmd.AddCommand(newTaskCmd())

	return cmd
}
