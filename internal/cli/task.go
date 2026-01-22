package cli

import (
	"errors"
	"fmt"
	"strings"
	"time"

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
	var section string

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
			if section != "" {
				if err := domain.ValidateSection(section); err != nil {
					return err
				}
			}

			found := false
			for _, category := range project.Categories {
				tasks := append([]domain.Task(nil), category.Tasks...)
				domain.SortTasks(tasks)
				for _, task := range tasks {
					if status != "" && task.Status != status {
						continue
					}
					if section != "" && task.Section != section {
						continue
					}
					fmt.Fprintf(cmd.OutOrStdout(), "%s\t%s\t%s\t%s (%s)\n", category.Name, task.Status, task.Section, task.Title, task.ID)
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
	cmd.Flags().StringVar(&section, "section", "", "filter by section")
	return cmd
}

func newTaskAddCmd() *cobra.Command {
	var categoryName string
	var priority string
	var deadline string
	var estimate string

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

			if deadline != "" {
				if _, err := time.Parse("2006-01-02", deadline); err != nil {
					return fmt.Errorf("invalid deadline format: %w", err)
				}
				task.Deadline = deadline
			}

			if estimate != "" {
				value, unit, err := parseEstimate(estimate)
				if err != nil {
					return err
				}
				task.TimeEstimateValue = value
				task.TimeEstimateUnit = unit
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
	cmd.Flags().StringVar(&deadline, "deadline", "", "deadline YYYY-MM-DD")
	cmd.Flags().StringVar(&estimate, "estimate", "", "estimate: 30m, 2h, 1d")
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
						task.Section = domain.SectionPast
					}
					if status == domain.StatusCancelled {
						task.Section = domain.SectionPast
					}
					if status == domain.StatusTodo || status == domain.StatusInProgress {
						if task.Section == domain.SectionPast {
							task.Section = domain.SectionCurrent
						}
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

func parseEstimate(raw string) (int, string, error) {
	if len(raw) < 2 {
		return 0, "", fmt.Errorf("invalid estimate %q", raw)
	}
	unitChar := raw[len(raw)-1:]
	valuePart := raw[:len(raw)-1]
	value, err := parsePositiveInt(valuePart)
	if err != nil {
		return 0, "", fmt.Errorf("invalid estimate %q", raw)
	}
	switch unitChar {
	case "m":
		return value, domain.EstimateMinutes, nil
	case "h":
		return value, domain.EstimateHours, nil
	case "d":
		return value, domain.EstimateDays, nil
	default:
		return 0, "", fmt.Errorf("invalid estimate unit %q", unitChar)
	}
}

func parsePositiveInt(raw string) (int, error) {
	if raw == "" {
		return 0, errors.New("missing number")
	}
	var value int
	for _, r := range raw {
		if r < '0' || r > '9' {
			return 0, fmt.Errorf("invalid number %q", raw)
		}
		value = value*10 + int(r-'0')
	}
	if value <= 0 {
		return 0, errors.New("number must be positive")
	}
	return value, nil
}
