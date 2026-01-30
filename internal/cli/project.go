package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/config"
	"phasionary/internal/data"
)

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}

	cmd.AddCommand(newProjectListCmd())
	cmd.AddCommand(newProjectAddCmd())

	return cmd
}

func newProjectListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			projects, err := store.ListProjects()
			if err != nil {
				return err
			}
			if len(projects) == 0 {
				fmt.Fprintln(cmd.OutOrStdout(), "No projects found.")
				return nil
			}
			for _, project := range projects {
				fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\n", project.Name, project.ID)
			}
			return nil
		},
	}
	return cmd
}

func newProjectAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add <name>",
		Short: "Add a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.CreateProject(args[0])
			if err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "Created project: %s (%s)\n", project.Name, project.ID)
			return nil
		},
	}
	return cmd
}

func storeFromViper() (*data.Store, error) {
	dataDir, err := config.ResolveDataDir(viper.GetString("data"))
	if err != nil {
		return nil, err
	}
	return data.NewStore(dataDir), nil
}
