package cli

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"phasionary/internal/domain"
)

func newTasksCmd() *cobra.Command {
	var (
		status   string
		category string
		priority string
	)

	cmd := &cobra.Command{
		Use:     "tasks",
		Aliases: []string{"ts"},
		Short:   "List tasks",
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
			if priority != "" {
				if err := domain.ValidatePriority(priority); err != nil {
					return err
				}
			}

			var tasks []TaskListItem
			for _, cat := range project.Categories {
				if category != "" && domain.NormalizeName(cat.Name) != domain.NormalizeName(category) {
					continue
				}
				for _, task := range cat.Tasks {
					if status != "" && task.Status != status {
						continue
					}
					if priority != "" && task.Priority != priority {
						continue
					}
					tasks = append(tasks, TaskListItem{
						ID:              task.ID,
						Title:           task.Title,
						Status:          task.Status,
						Priority:        task.Priority,
						Category:        cat.Name,
						EstimateMinutes: task.EstimateMinutes,
					})
				}
			}

			return writeTaskList(cmd.OutOrStdout(), tasks)
		},
	}

	cmd.Flags().StringVarP(&status, "status", "s", "", "filter by status (todo, in_progress, completed, cancelled)")
	cmd.Flags().StringVarP(&category, "category", "C", "", "filter by category name")
	cmd.Flags().StringVar(&priority, "priority", "", "filter by priority (high, medium, low)")

	_ = cmd.RegisterFlagCompletionFunc("status", completeStatuses)
	_ = cmd.RegisterFlagCompletionFunc("category", completeCategories)
	_ = cmd.RegisterFlagCompletionFunc("priority", completePriorities)

	return cmd
}

func newTaskCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "task",
		Short: "Manage tasks",
	}

	cmd.AddCommand(newTaskShowCmd())
	cmd.AddCommand(newTaskAddCmd())
	cmd.AddCommand(newTaskEditCmd())
	cmd.AddCommand(newTaskDeleteCmd())
	cmd.AddCommand(newTaskStatusCmd())
	cmd.AddCommand(newTaskPriorityCmd())
	cmd.AddCommand(newTaskMoveCmd())

	return cmd
}

func newTaskShowCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "show <id-or-title>",
		Aliases:           []string{"t"},
		Short:             "Show task details",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTasks,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			task, catName, _, _, err := resolveTask(project, args[0])
			if err != nil {
				return fmt.Errorf("task %q not found", args[0])
			}

			return writeTaskDetail(cmd.OutOrStdout(), *task, catName)
		},
	}
	return cmd
}

func newTaskAddCmd() *cobra.Command {
	var (
		categoryName string
		priority     string
		estimate     string
	)

	cmd := &cobra.Command{
		Use:     "add <title>",
		Aliases: []string{"ta"},
		Short:   "Add a task",
		Args:    cobra.ExactArgs(1),
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

			if estimate != "" {
				minutes, err := parseTimeEstimate(estimate)
				if err != nil {
					return err
				}
				task.EstimateMinutes = minutes
			}

			cat, catIdx, err := resolveCategory(project, categoryName)
			if err != nil {
				return fmt.Errorf("category %q not found", categoryName)
			}

			cat.Tasks = append(cat.Tasks, task)
			project.Categories[catIdx] = *cat
			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Created task: %s (%s)", task.Title, task.ID))
			return nil
		},
	}

	cmd.Flags().StringVarP(&categoryName, "category", "C", "", "category name (required)")
	cmd.Flags().StringVar(&priority, "priority", "", "priority: high|medium|low")
	cmd.Flags().StringVarP(&estimate, "estimate", "e", "", "time estimate: 30, 2h, 1.5h, 2h30m")

	_ = cmd.RegisterFlagCompletionFunc("category", completeCategories)
	_ = cmd.RegisterFlagCompletionFunc("priority", completePriorities)

	return cmd
}

func newTaskEditCmd() *cobra.Command {
	var (
		title    string
		priority string
		estimate string
	)

	cmd := &cobra.Command{
		Use:               "edit <id-or-title>",
		Aliases:           []string{"te"},
		Short:             "Edit task properties",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTasks,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			task, _, catIdx, taskIdx, err := resolveTask(project, args[0])
			if err != nil {
				return fmt.Errorf("task %q not found", args[0])
			}

			if title != "" {
				task.Title = title
			}
			if priority != "" {
				if err := task.SetPriority(priority); err != nil {
					return err
				}
			}
			if estimate != "" {
				minutes, err := parseTimeEstimate(estimate)
				if err != nil {
					return err
				}
				task.SetEstimate(minutes)
			}

			project.Categories[catIdx].Tasks[taskIdx] = *task
			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Updated task: %s", task.Title))
			return nil
		},
	}

	cmd.Flags().StringVarP(&title, "title", "t", "", "new title")
	cmd.Flags().StringVar(&priority, "priority", "", "priority: high|medium|low")
	cmd.Flags().StringVarP(&estimate, "estimate", "e", "", "time estimate: 30, 2h, 1.5h, 2h30m")

	_ = cmd.RegisterFlagCompletionFunc("priority", completePriorities)

	return cmd
}

func newTaskDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:               "delete <id-or-title>",
		Aliases:           []string{"td"},
		Short:             "Delete a task",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: completeTasks,
		RunE: func(cmd *cobra.Command, args []string) error {
			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			task, _, catIdx, taskIdx, err := resolveTask(project, args[0])
			if err != nil {
				return fmt.Errorf("task %q not found", args[0])
			}

			if !force {
				fmt.Fprintf(cmd.OutOrStdout(), "Delete task %q? [y/N]: ", task.Title)
				var response string
				if _, err := fmt.Fscanln(cmd.InOrStdin(), &response); err != nil {
					return nil
				}
				if response != "y" && response != "Y" {
					fmt.Fprintln(cmd.OutOrStdout(), "Cancelled.")
					return nil
				}
			}

			if err := project.Categories[catIdx].RemoveTask(taskIdx); err != nil {
				return err
			}
			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Deleted task: %s", task.Title))
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "skip confirmation prompt")

	return cmd
}

func newTaskStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "status <id-or-title> <status>",
		Aliases: []string{"tst"},
		Short:   "Update task status",
		Args:    cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeTasks(cmd, args, toComplete)
			}
			return completeStatuses(cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			selector := args[0]
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

			task, _, catIdx, taskIdx, err := resolveTask(project, selector)
			if err != nil {
				return fmt.Errorf("task %q not found", selector)
			}

			if err := task.SetStatus(status); err != nil {
				return err
			}

			project.Categories[catIdx].Tasks[taskIdx] = *task
			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Updated task %s to %s", task.Title, status))
			return nil
		},
	}
	return cmd
}

func newTaskPriorityCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "priority <id-or-title> <priority>",
		Aliases: []string{"tp"},
		Short:   "Update task priority",
		Args:    cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeTasks(cmd, args, toComplete)
			}
			return completePriorities(cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			selector := args[0]
			priority := args[1]

			if err := domain.ValidatePriority(priority); err != nil {
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

			task, _, catIdx, taskIdx, err := resolveTask(project, selector)
			if err != nil {
				return fmt.Errorf("task %q not found", selector)
			}

			if err := task.SetPriority(priority); err != nil {
				return err
			}

			project.Categories[catIdx].Tasks[taskIdx] = *task
			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Updated task %s priority to %s", task.Title, priority))
			return nil
		},
	}
	return cmd
}

func newTaskMoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "move <id-or-title> <category>",
		Aliases: []string{"tm"},
		Short:   "Move task to different category",
		Args:    cobra.ExactArgs(2),
		ValidArgsFunction: func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			if len(args) == 0 {
				return completeTasks(cmd, args, toComplete)
			}
			return completeCategories(cmd, args, toComplete)
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			selector := args[0]
			targetCategory := args[1]

			store, err := storeFromViper()
			if err != nil {
				return err
			}
			project, err := store.LoadProject(viper.GetString("project"))
			if err != nil {
				return err
			}

			task, _, srcCatIdx, taskIdx, err := resolveTask(project, selector)
			if err != nil {
				return fmt.Errorf("task %q not found", selector)
			}

			_, dstCatIdx, err := resolveCategory(project, targetCategory)
			if err != nil {
				return fmt.Errorf("category %q not found", targetCategory)
			}

			if srcCatIdx == dstCatIdx {
				return fmt.Errorf("task is already in category %q", targetCategory)
			}

			taskCopy := *task
			if err := project.Categories[srcCatIdx].RemoveTask(taskIdx); err != nil {
				return err
			}
			project.Categories[dstCatIdx].AddTask(taskCopy)

			if err := store.SaveProject(project); err != nil {
				return err
			}

			writeSuccess(cmd.OutOrStdout(), fmt.Sprintf("Moved task %s to %s", task.Title, project.Categories[dstCatIdx].Name))
			return nil
		},
	}
	return cmd
}
