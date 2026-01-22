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
	editing    bool
	editValue  string
	editCursor int
	store      *data.Store
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
		if m.editing {
			m.handleEditKey(msg)
			break
		}
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "up", "k":
			m.moveSelection(-1)
		case "down", "j":
			m.moveSelection(1)
		case " ":
			m.toggleSelectedTask()
		case "enter":
			m.startEditing()
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
			if m.editing && isSelected {
				bodyBuilder.WriteString(m.renderEditTaskLine(task))
			} else {
				bodyBuilder.WriteString(renderTaskLine(task, isSelected))
			}
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
		store:      store,
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
	priorityIcon := ui.PriorityIcon(task.Priority)
	if !selected {
		status := formatStatus(task.Status)
		icon := ""
		if priorityIcon != "" {
			icon = ui.PriorityStyle(task.Priority).Render(priorityIcon) + " "
		}
		titleStyle := ui.PriorityStyle(task.Priority)
		title := titleStyle.Render(task.Title)
		return fmt.Sprintf("%s[%s] %s%s", prefix, status, icon, title)
	}
	statusText := statusLabel(task.Status)
	statusStyle := ui.StatusStyle(task.Status).Bold(true).Reverse(true)
	titleStyle := ui.PriorityStyle(task.Priority).Bold(true).Reverse(true)
	icon := ""
	if priorityIcon != "" {
		icon = titleStyle.Render(priorityIcon) + " "
	}
	title := titleStyle.Render(task.Title)
	return ui.SelectedStyle.Render(prefix+"[") +
		statusStyle.Render(statusText) +
		ui.SelectedStyle.Render("] ") +
		icon +
		title
}

func statusLabel(status string) string {
	switch status {
	case domain.StatusInProgress:
		return "in progress"
	case domain.StatusCompleted:
		return "completed"
	case domain.StatusCancelled:
		return "cancelled"
	default:
		return "todo"
	}
}

func formatStatus(status string) string {
	return ui.StatusStyle(status).Render(statusLabel(status))
}

func (m *model) moveSelection(delta int) {
	if m.editing || len(m.positions) == 0 {
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
	if m.editing {
		return ui.StatusLineStyle.Render("Shortcuts: enter save | esc cancel | arrows move cursor")
	}
	return ui.StatusLineStyle.Render("Shortcuts: up/down or j/k move | enter edit title | space toggle status | q/ctrl+c quit")
}

func (m *model) startEditing() {
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	category := m.categories[position.CategoryIndex]
	task := category.Tasks[position.TaskIndex]
	m.editing = true
	m.editValue = task.Title
	m.editCursor = len([]rune(m.editValue))
}

func (m *model) handleEditKey(msg tea.KeyMsg) {
	switch msg.String() {
	case "enter":
		m.finishEditing()
	case "esc":
		m.cancelEditing()
	case "left":
		m.moveEditCursor(-1)
	case "right":
		m.moveEditCursor(1)
	case "backspace":
		m.deleteEditRune(-1)
	case "delete":
		m.deleteEditRune(0)
	case " ", "space":
		m.insertEditRunes([]rune(" "))
	default:
		if msg.Type == tea.KeyRunes {
			m.insertEditRunes(msg.Runes)
		}
	}
}

func (m *model) finishEditing() {
	if !m.editing {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		m.cancelEditing()
		return
	}
	trimmed := strings.TrimSpace(m.editValue)
	if trimmed == "" {
		m.cancelEditing()
		return
	}
	category := &m.categories[position.CategoryIndex]
	task := &category.Tasks[position.TaskIndex]
	if task.Title != trimmed {
		task.Title = trimmed
		task.UpdatedAt = domain.NowTimestamp()
		m.syncTaskToProject(position, *task)
		m.storeTaskUpdate()
	}
	m.refreshTaskView(position)
	m.editing = false
	m.editValue = ""
	m.editCursor = 0
}

func (m *model) cancelEditing() {
	m.editing = false
	m.editValue = ""
	m.editCursor = 0
}

func (m *model) moveEditCursor(delta int) {
	runes := []rune(m.editValue)
	next := m.editCursor + delta
	if next < 0 {
		next = 0
	}
	if next > len(runes) {
		next = len(runes)
	}
	m.editCursor = next
}

func (m *model) insertEditRunes(runesToInsert []rune) {
	if len(runesToInsert) == 0 {
		return
	}
	runes := []rune(m.editValue)
	cursor := m.editCursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	updated := make([]rune, 0, len(runes)+len(runesToInsert))
	updated = append(updated, runes[:cursor]...)
	updated = append(updated, runesToInsert...)
	updated = append(updated, runes[cursor:]...)
	m.editValue = string(updated)
	m.editCursor = cursor + len(runesToInsert)
}

func (m *model) deleteEditRune(offset int) {
	runes := []rune(m.editValue)
	if len(runes) == 0 {
		return
	}
	index := m.editCursor + offset
	if offset < 0 {
		index = m.editCursor - 1
	}
	if index < 0 || index >= len(runes) {
		return
	}
	updated := append([]rune{}, runes[:index]...)
	updated = append(updated, runes[index+1:]...)
	m.editValue = string(updated)
	if offset < 0 {
		m.editCursor = index
	} else if m.editCursor > len(updated) {
		m.editCursor = len(updated)
	}
}

func (m model) renderEditTaskLine(task domain.Task) string {
	prefix := "> "
	statusText := formatStatus(task.Status)
	edited := m.editValue
	if edited == "" {
		edited = " "
	}
	runes := []rune(edited)
	cursor := m.editCursor
	if cursor < 0 {
		cursor = 0
	}
	if cursor > len(runes) {
		cursor = len(runes)
	}
	left := string(runes[:cursor])
	right := string(runes[cursor:])
	cursorChar := " "
	if cursor < len(runes) {
		cursorChar = string(runes[cursor])
		right = string(runes[cursor+1:])
	}
	cursorStyle := ui.SelectedStyle
	titleStyle := ui.PriorityStyle(task.Priority)
	icon := ui.PriorityIcon(task.Priority)
	iconPrefix := ""
	if icon != "" {
		iconPrefix = titleStyle.Render(icon) + " "
	}
	return fmt.Sprintf(
		"%s[%s] %s%s%s%s",
		prefix,
		statusText,
		iconPrefix,
		titleStyle.Render(left),
		cursorStyle.Render(cursorChar),
		titleStyle.Render(right),
	)
}

func (m model) selectedPosition() (focusPosition, bool) {
	if m.selected < 0 || m.selected >= len(m.positions) {
		return focusPosition{}, false
	}
	return m.positions[m.selected], true
}

func (m *model) toggleSelectedTask() {
	if m.editing {
		return
	}
	position, ok := m.selectedPosition()
	if !ok || position.Kind != focusTask {
		return
	}
	category := &m.categories[position.CategoryIndex]
	task := &category.Tasks[position.TaskIndex]
	nextStatus := nextTaskStatus(task.Status)
	if nextStatus == task.Status {
		return
	}
	updateTaskStatus(task, nextStatus)
	m.syncTaskToProject(position, *task)
	m.storeTaskUpdate()
}

func nextTaskStatus(current string) string {
	switch current {
	case domain.StatusTodo:
		return domain.StatusInProgress
	case domain.StatusInProgress:
		return domain.StatusCompleted
	case domain.StatusCompleted:
		return domain.StatusTodo
	case domain.StatusCancelled:
		return domain.StatusTodo
	default:
		return domain.StatusTodo
	}
}

func updateTaskStatus(task *domain.Task, status string) {
	task.Status = status
	task.UpdatedAt = domain.NowTimestamp()
	if status == domain.StatusCompleted {
		task.CompletionDate = domain.NowTimestamp()
		task.Section = domain.SectionPast
		return
	}
	if status == domain.StatusCancelled {
		task.Section = domain.SectionPast
		return
	}
	if task.Section == domain.SectionPast {
		task.Section = domain.SectionCurrent
	}
	task.CompletionDate = ""
}

func (m *model) syncTaskToProject(position focusPosition, task domain.Task) {
	if position.CategoryIndex < 0 || position.CategoryIndex >= len(m.project.Categories) {
		return
	}
	projectCategory := &m.project.Categories[position.CategoryIndex]
	for index := range projectCategory.Tasks {
		if projectCategory.Tasks[index].ID == task.ID {
			projectCategory.Tasks[index] = task
			return
		}
	}
	projectCategory.Tasks = append(projectCategory.Tasks, task)
}

func (m *model) storeTaskUpdate() {
	if m.store == nil {
		return
	}
	_ = m.store.SaveProject(m.project)
}

func (m *model) refreshTaskView(position focusPosition) {
	if position.CategoryIndex < 0 || position.CategoryIndex >= len(m.categories) {
		return
	}
	category := &m.categories[position.CategoryIndex]
	sorted := append([]domain.Task(nil), category.Tasks...)
	domain.SortTasks(sorted)
	category.Tasks = sorted
	m.positions = rebuildPositions(m.categories)
	m.selected = m.findPositionForTask(position, category.Tasks)
}

func rebuildPositions(categories []categoryView) []focusPosition {
	positions := make([]focusPosition, 0)
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
	return positions
}

func (m *model) findPositionForTask(previous focusPosition, tasks []domain.Task) int {
	if previous.CategoryIndex < 0 || previous.CategoryIndex >= len(m.categories) {
		return m.selected
	}
	if previous.TaskIndex < 0 || previous.TaskIndex >= len(m.categories[previous.CategoryIndex].Tasks) {
		return m.selected
	}
	taskID := m.categories[previous.CategoryIndex].Tasks[previous.TaskIndex].ID
	for index, position := range m.positions {
		if position.Kind == focusTask &&
			position.CategoryIndex == previous.CategoryIndex &&
			position.TaskIndex >= 0 &&
			position.TaskIndex < len(tasks) &&
			tasks[position.TaskIndex].ID == taskID {
			return index
		}
	}
	return m.selected
}
