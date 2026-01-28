package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"phasionary/internal/data"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

type clipboardResultMsg struct{ err error }

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
	showHelp   bool
	editValue  string
	editCursor int
	store      *data.Store
	addingTask     bool   // true when adding a new task (vs editing existing)
	newTaskID      string // ID of task being added (for removal on cancel)
	addingCategory bool   // true when adding a new category
	newCategoryID  string // ID of category being added (for removal on cancel)
	statusMsg  string // temporary status message (e.g., "Copied!")
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case clipboardResultMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Copy failed: %v", msg.err)
		} else {
			m.statusMsg = "Copied!"
		}
	case tea.KeyMsg:
		m.statusMsg = ""
		if msg.String() == "?" {
			m.showHelp = !m.showHelp
			break
		}
		if m.showHelp {
			switch msg.String() {
			case "q", "esc":
				m.showHelp = false
			}
			break
		}
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
		case "a":
			m.startAddingTask()
		case "A":
			m.startAddingCategory()
		case "h":
			m.decreasePriority()
		case "l":
			m.increasePriority()
		case "y":
			return m, m.copySelected()
		}
	}
	return m, nil
}

func (m model) copySelected() tea.Cmd {
	pos, ok := m.selectedPosition()
	if !ok {
		return nil
	}
	var text string
	if pos.Kind == focusCategory {
		text = m.categories[pos.CategoryIndex].Name
	} else {
		text = m.categories[pos.CategoryIndex].Tasks[pos.TaskIndex].Title
	}
	return func() tea.Msg {
		return clipboardResultMsg{err: clipboard.WriteAll(text)}
	}
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
		if m.editing && isSelected {
			bodyBuilder.WriteString(m.renderEditCategoryLine())
		} else {
			bodyBuilder.WriteString(renderCategoryLine(category.Name, isSelected))
		}
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
	content := header + "  " + project + "\n\n" + body + "\n\n" + statusLine + "\n" + shortcuts + "\n"
	if m.showHelp {
		help := m.helpView()
		if m.width > 0 && m.height > 0 {
			return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, help)
		}
		return help
	}
	return content
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
