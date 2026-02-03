package cli

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/domain"
)

func newCategoriesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "categories",
		Aliases: []string{"cs"},
		Short:   "List categories",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}
			return writeCategories(cmd.OutOrStdout(), project.Categories)
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
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			cat, _, err := resolveCategory(project, args[0])
			if err != nil {
				return fmt.Errorf("category %q not found", args[0])
			}

			return writeCategoryDetail(cmd.OutOrStdout(), *cat)
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
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			name := args[0]
			for _, cat := range project.Categories {
				if domain.NormalizeName(cat.Name) == domain.NormalizeName(name) {
					return fmt.Errorf("category %q already exists", name)
				}
			}

			cat, err := domain.NewCategory(name)
			if err != nil {
				return err
			}

			project.AddCategory(cat)
			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Created category: %s (%s)", cat.Name, cat.ID))
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

			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			cat, catIdx, err := resolveCategory(project, args[0])
			if err != nil {
				return fmt.Errorf("category %q not found", args[0])
			}

			for i, c := range project.Categories {
				if i != catIdx && domain.NormalizeName(c.Name) == domain.NormalizeName(name) {
					return fmt.Errorf("category %q already exists", name)
				}
			}

			cat.Name = name
			cat.UpdatedAt = domain.NowTimestamp()
			project.Categories[catIdx] = *cat

			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Renamed category to: %s", cat.Name))
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
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
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
