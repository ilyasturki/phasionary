package app

import (
	"errors"
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"

	"phasionary/internal/data"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

type categoryView struct {
	Name  string
	Tasks []domain.Task
}

type focusKind int

const (
	focusCategory focusKind = iota
	focusTask
)

type focusPosition struct {
	Kind          focusKind
	CategoryIndex int
	TaskIndex     int
}

type model struct {
	project    domain.Project
	categories []categoryView
	positions  []focusPosition
	selected   int
	width      int
	height     int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.moveSelection(-1)
		case "down", "j":
			m.moveSelection(1)
		}
	}
	return m, nil
}

func (m model) View() string {
	header := ui.HeaderStyle.Render("Phasionary")
	project := ui.MutedStyle.Render(fmt.Sprintf("Project: %s", m.project.Name))

	var bodyBuilder strings.Builder
	cursor := 0
	for i, category := range m.categories {
		if i > 0 {
			bodyBuilder.WriteString("\n")
		}
		isSelected := cursor == m.selected
		bodyBuilder.WriteString(renderCategoryLine(category.Name, isSelected))
		cursor++
		if len(category.Tasks) == 0 {
			bodyBuilder.WriteString("\n")
			bodyBuilder.WriteString(ui.MutedStyle.Render("  (no tasks)"))
			bodyBuilder.WriteString("\n")
			continue
		}
		bodyBuilder.WriteString("\n")
		for _, task := range category.Tasks {
			isSelected = cursor == m.selected
			bodyBuilder.WriteString(renderTaskLine(task, isSelected))
			bodyBuilder.WriteString("\n")
			cursor++
		}
	}

	body := strings.TrimRight(bodyBuilder.String(), "\n")
	statusLine := m.statusLine()
	shortcuts := m.shortcutsLine()

	return header + "  " + project + "\n\n" + body + "\n\n" + statusLine + "\n" + shortcuts + "\n"
}

func Run(dataDir string, projectSelector string) error {
	store := data.NewStore(dataDir)
	if err := store.Ensure(); err != nil {
		return err
	}
	projects, err := store.ListProjects()
	if err != nil {
		return err
	}
	if len(projects) == 0 {
		if _, err := store.InitDefault(); err != nil {
			return err
		}
	}

	project, err := store.LoadProject(projectSelector)
	if err != nil {
		if errors.Is(err, data.ErrProjectNotFound) {
			if projectSelector == "" {
				return fmt.Errorf("no projects available")
			}
			return fmt.Errorf("project %q not found", projectSelector)
		}
		return err
	}

	categories, positions := buildViews(project)
	selected := -1
	if len(positions) > 0 {
		selected = 0
		for i, position := range positions {
			if position.Kind == focusTask {
				selected = i
				break
			}
		}
	}

	program := tea.NewProgram(model{
		project:    project,
		categories: categories,
		positions:  positions,
		selected:   selected,
	}, tea.WithAltScreen())
	_, err = program.Run()
	return err
}

func buildViews(project domain.Project) ([]categoryView, []focusPosition) {
	categories := make([]categoryView, 0, len(project.Categories))
	positions := make([]focusPosition, 0)
	for _, category := range project.Categories {
		tasks := append([]domain.Task(nil), category.Tasks...)
		domain.SortTasks(tasks)
		categories = append(categories, categoryView{
			Name:  category.Name,
			Tasks: tasks,
		})
	}
	for cIndex, category := range categories {
		positions = append(positions, focusPosition{
			Kind:          focusCategory,
			CategoryIndex: cIndex,
			TaskIndex:     -1,
		})
		for tIndex := range category.Tasks {
			positions = append(positions, focusPosition{
				Kind:          focusTask,
				CategoryIndex: cIndex,
				TaskIndex:     tIndex,
			})
		}
	}
	return categories, positions
}

func renderCategoryLine(name string, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	line := fmt.Sprintf("%s%s", prefix, name)
	if selected {
		return ui.SelectedStyle.Render(line)
	}
	return ui.CategoryStyle.Render(line)
}

func renderTaskLine(task domain.Task, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	status := formatStatus(task.Status)
	line := fmt.Sprintf("%s[%s] %s", prefix, status, task.Title)
	if selected {
		return ui.SelectedStyle.Render(line)
	}
	return line
}

func formatStatus(status string) string {
	switch status {
	case domain.StatusInProgress:
		return ui.StatusStyle(status).Render("in progress")
	case domain.StatusCompleted:
		return ui.StatusStyle(status).Render("completed")
	case domain.StatusCancelled:
		return ui.StatusStyle(status).Render("cancelled")
	default:
		return ui.StatusStyle(status).Render("todo")
	}
}

func (m *model) moveSelection(delta int) {
	if len(m.positions) == 0 {
		return
	}
	next := m.selected + delta
	if next < 0 {
		next = 0
	}
	if next >= len(m.positions) {
		next = len(m.positions) - 1
	}
	m.selected = next
}

func (m model) statusLine() string {
	position, ok := m.selectedPosition()
	if !ok {
		return ui.StatusLineStyle.Render("No items to display.")
	}
	category := m.categories[position.CategoryIndex]
	if position.Kind == focusCategory {
		summary := fmt.Sprintf("Category: %s (%d tasks)", category.Name, len(category.Tasks))
		return ui.StatusLineStyle.Render(summary)
	}
	task := category.Tasks[position.TaskIndex]
	summary := fmt.Sprintf("Selected: %s / %s (%s - %s)", category.Name, task.Title, task.Status, task.Section)
	return ui.StatusLineStyle.Render(summary)
}

func (m model) shortcutsLine() string {
	return ui.StatusLineStyle.Render("Shortcuts: up/down or j/k move | q/ctrl+c quit")
}

func (m model) selectedPosition() (focusPosition, bool) {
	if m.selected < 0 || m.selected >= len(m.positions) {
		return focusPosition{}, false
	}
	return m.positions[m.selected], true
}
