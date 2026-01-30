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
			configPath, err := config.ResolveConfigPath(viper.GetString("config"))
			if err != nil {
				return err
			}
			cfgManager := config.NewManager(configPath)
			if err := cfgManager.Load(); err != nil {
				return err
			}

			dataDir, err := config.ResolveDataDir(viper.GetString("data"))
			if err != nil {
				return err
			}
			return app.Run(dataDir, viper.GetString("project"), cfgManager)
		},
	}

	cmd.SetOut(os.Stdout)
	cmd.SetErr(os.Stderr)

	cmd.PersistentFlags().String("config", "", "override config directory path")
	cmd.PersistentFlags().String("data", "", "override data directory path")
	cmd.PersistentFlags().String("project", "", "target project for commands")

	viper.AutomaticEnv()
	if err := viper.BindEnv("config", config.EnvConfigPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := viper.BindEnv("data", config.EnvDataPath); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	if err := viper.BindPFlag("config", cmd.PersistentFlags().Lookup("config")); err != nil {
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
