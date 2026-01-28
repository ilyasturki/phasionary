package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/domain"
)

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}

	cmd.AddCommand(newTaskListCmd())
	cmd.AddCommand(newTaskAddCmd())
	cmd.AddCommand(newTaskStatusCmd())

	return cmd
}

func newTaskListCmd() *cobra.Command {
	var status string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tasks",
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}
			if status != "" {
				if err := domain.ValidateStatus(status); err != nil {
					return err
				}
			}

			found := false
			for _, category := range project.Categories {
				for _, task := range category.Tasks {
					if status != "" && task.Status != status {
						continue
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s (%s)\n", category.Name, task.Status, task.Title, task.ID)
					found = true
				}
			}
			if !found {
				fmt.Fprintln(cmd.OutOrStdout(), "No tasks found.")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "filter by status")
	return cmd
}

func newTaskAddCmd() *cobra.Command {
	var categoryName string
	var priority string

	cmd := &cobra.Command{
		Use:   "add <title>",
		Short: "Add a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(categoryName) == "" {
				return errors.New("--category is required")
			}
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			task, err := domain.NewTask(args[0])
			if err != nil {
				return err
			}

			if priority != "" {
				if err := domain.ValidatePriority(priority); err != nil {
					return err
				}
				task.Priority = priority
			}

			categoryIndex := -1
			for i, category := range project.Categories {
				if domain.NormalizeName(category.Name) == domain.NormalizeName(categoryName) {
					categoryIndex = i
					break
				}
			}
			if categoryIndex == -1 {
				return fmt.Errorf("category %q not found", categoryName)
			}

			project.Categories[categoryIndex].Tasks = append(project.Categories[categoryIndex].Tasks, task)
			if err := store.SaveProject(project); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Created task: %s (%s)\n", task.Title, task.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&categoryName, "category", "", "category name")
	cmd.Flags().StringVar(&priority, "priority", "", "priority: high|medium|low")
	return cmd
}

func newTaskStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <id> <status>",
		Short: "Update task status",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			id := args[0]
			status := args[1]

			if err := domain.ValidateStatus(status); err != nil {
				return err
			}

			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			updated := false
			for cIndex := range project.Categories {
				for tIndex := range project.Categories[cIndex].Tasks {
					task := &project.Categories[cIndex].Tasks[tIndex]
					if task.ID != id {
						continue
					}
					task.Status = status
					task.UpdatedAt = domain.NowTimestamp()
					if status == domain.StatusCompleted {
						task.CompletionDate = domain.NowTimestamp()
					} else {
						task.CompletionDate = ""
					}
					updated = true
					break
				}
				if updated {
					break
				}
			}

			if !updated {
				return fmt.Errorf("task %s not found", id)
			}

			if err := store.SaveProject(project); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Updated task %s to %s\n", id, status)
			return nil
		},
	}
	return cmd
}
