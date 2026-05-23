package cli

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/data"
	"phasionary/internal/domain"
)

func newCategoriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "categories",
		Aliases: []string{"cs"},
		Short:   "List categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			return viewProject(viper.GetString("project"), func(project domain.Project) error {
				return writeCategories(cmd.OutOrStdout(), project.Categories)
			})
		},
	}
	return cmd
}

func newCategoryCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "category",
		Short: "Manage categories",
	}

	cmd.AddCommand(newCategoryShowCmd())
	cmd.AddCommand(newCategoryAddCmd())
	cmd.AddCommand(newCategoryEditCmd())
	cmd.AddCommand(newCategoryDeleteCmd())

	return cmd
}

func newCategoryShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "show <name-or-id>",
		Aliases:           []string{"c"},
		Short:             "Show category details",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeCategories,
		RunE: func(cmd *cobra.Command, args []string) error {
			return viewProject(viper.GetString("project"), func(project domain.Project) error {
				cat, _, err := resolveCategory(project, args[0])
				if err != nil {
					return fmt.Errorf("category %q not found", args[0])
				}
				return writeCategoryDetail(cmd.OutOrStdout(), *cat)
			})
		},
	}
	return cmd
}

func newCategoryAddCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "add <name>",
		Aliases: []string{"ca"},
		Short:   "Add a category",
		Args:    cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			name := args[0]
			var created *domain.Category
			err := withProject(viper.GetString("project"), func(_ *data.Store, project *domain.Project) error {
				cat, err := project.AddCategoryNamed(name)
				if err != nil {
					if errors.Is(err, domain.ErrDuplicateCategoryName) {
						return fmt.Errorf("category %q already exists", name)
					}
					return err
				}
				created = cat
				return nil
			})
			if err != nil {
				return err
			}
			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Created category: %s (%s)", created.Name, created.ID))
			return nil
		},
	}
	return cmd
}

func newCategoryEditCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:               "edit <name-or-id>",
		Aliases:           []string{"ce"},
		Short:             "Edit category (rename)",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeCategories,
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("--name is required")
			}

			err := withProject(viper.GetString("project"), func(_ *data.Store, project *domain.Project) error {
				_, catIdx, err := resolveCategory(*project, args[0])
				if err != nil {
					return fmt.Errorf("category %q not found", args[0])
				}
				if err := project.RenameCategory(catIdx, name); err != nil {
					if errors.Is(err, domain.ErrDuplicateCategoryName) {
						return fmt.Errorf("category %q already exists", name)
					}
					return err
				}
				return nil
			})
			if err != nil {
				return err
			}
			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Renamed category to: %s", name))
			return nil
		},
	}

	cmd.Flags().StringVarP(&name, "name", "n", "", "new category name")

	return cmd
}

func newCategoryDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:               "delete <name-or-id>",
		Aliases:           []string{"cd"},
		Short:             "Delete a category",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeCategories,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, project, err := loadStoreAndProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			cat, catIdx, err := resolveCategory(project, args[0])
			if err != nil {
				return fmt.Errorf("category %q not found", args[0])
			}

			if len(cat.Tasks) > 0 && !force {
				fmt.Fprintf(cmd.OutOrStdout(), "Category %q has %d tasks. Delete anyway? [y/N]: ", cat.Name, len(cat.Tasks))
				var response string
				if _, err := fmt.Fscanln(cmd.InOrStdin(), &response); err != nil {
					return nil
				}
				if response != "y" && response != "Y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
					return nil
				}
			}

			if err := project.RemoveCategory(catIdx); err != nil {
				return err
			}
			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Deleted category: %s", cat.Name))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")

	return cmd
}
