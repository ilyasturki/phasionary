package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/config"
	"phasionary/internal/data"
)

func newProjectsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "projects",
		Aliases: []string{"ps"},
		Short:   "List all projects",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			projects, err := store.ListProjects()
			if err != nil {
				return err
			}
			return writeProjects(cmd.OutOrStdout(), projects)
		},
	}
	return cmd
}

func newProjectCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "project",
		Short: "Manage projects",
	}

	cmd.AddCommand(newProjectShowCmd())
	cmd.AddCommand(newProjectAddCmd())
	cmd.AddCommand(newProjectEditCmd())
	cmd.AddCommand(newProjectDeleteCmd())
	cmd.AddCommand(newProjectUseCmd())

	return cmd
}

func newProjectShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "show [name-or-id]",
		Aliases: []string{"p"},
		Short:   "Show project details",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			selector := viper.GetString("project")
			if len(args) > 0 {
				selector = args[0]
			}
			project, err := store.LoadProject(selector)
			if err != nil {
				return err
			}
			return writeProjectDetail(cmd.OutOrStdout(), project)
		},
	}
	return cmd
}

func newProjectAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name>",
		Aliases: []string{"pa"},
		Short:   "Add a new project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.CreateProject(args[0])
			if err != nil {
				return err
			}
			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Created project: %s (%s)", project.Name, project.ID))
			return nil
		},
	}
	return cmd
}

func newProjectEditCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:     "edit [name-or-id]",
		Aliases: []string{"pe"},
		Short:   "Edit project (rename)",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			store, err := storeFromViper()
			if err != nil {
				return err
			}
			selector := viper.GetString("project")
			if len(args) > 0 {
				selector = args[0]
			}
			project, err := store.LoadProject(selector)
			if err != nil {
				return err
			}

			project.Name = name
			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Renamed project to: %s", project.Name))
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "new project name")

	return cmd
}

func newProjectDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:     "delete [name-or-id]",
		Aliases: []string{"pd"},
		Short:   "Delete a project",
		Args:    cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			selector := viper.GetString("project")
			if len(args) > 0 {
				selector = args[0]
			}
			project, err := store.LoadProject(selector)
			if err != nil {
				return err
			}

			if !force {
				fmt.Fprintf(cmd.OutOrStdout(), "Delete project %q? [y/N]: ", project.Name)
				var response string
				if _, err := fmt.Fscanln(cmd.InOrStdin(), &response); err != nil {
					return nil
				}
				if response != "y" && response != "Y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
					return nil
				}
			}

			if err := store.DeleteProject(project.ID); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Deleted project: %s", project.Name))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")

	return cmd
}

func newProjectUseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "use <name-or-id>",
		Aliases: []string{"pu"},
		Short:   "Set default project",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(args[0])
			if err != nil {
				return err
			}

			configPath, err := config.ResolveConfigPath(viper.GetString("config"))
			if err != nil {
				return err
			}
			cfgManager := config.NewManager(configPath)
			if err := cfgManager.Load(); err != nil {
				return err
			}
			cfgManager.SetDefaultProject(project.ID)
			if err := cfgManager.Save(); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Set default project to: %s", project.Name))
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
