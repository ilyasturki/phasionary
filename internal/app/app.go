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
	focusProject focusKind = iota
	focusCategory
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
	statusMsg     string // temporary status message (e.g., "Copied!")
	confirmDelete bool   // true when delete confirmation dialog is shown
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
		if m.confirmDelete {
			switch msg.String() {
			case "y", "enter":
				m.confirmDeleteAction()
			case "n", "esc":
				m.confirmDelete = false
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
		case "d":
			m.deleteSelected()
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
	switch pos.Kind {
	case focusProject:
		text = m.project.Name
	case focusCategory:
		text = m.categories[pos.CategoryIndex].Name
	default:
		text = m.categories[pos.CategoryIndex].Tasks[pos.TaskIndex].Title
	}
	return func() tea.Msg {
		return clipboardResultMsg{err: clipboard.WriteAll(text)}
	}
}

func (m model) View() string {
	var bodyBuilder strings.Builder
	cursor := 0

	// Project line (first focusable item)
	isProjectSelected := cursor == m.selected
	if m.editing && isProjectSelected {
		bodyBuilder.WriteString(m.renderEditProjectLine())
	} else {
		bodyBuilder.WriteString(renderProjectLine(m.project.Name, isProjectSelected))
	}
	cursor++
	bodyBuilder.WriteString("\n\n")

	for i, category := range m.categories {
		if i > 0 {
			bodyBuilder.WriteString("\n")
		}
		isSelected := cursor == m.selected
		if m.editing && isSelected {
			bodyBuilder.WriteString(m.renderEditCategoryLine())
		} else {
			bodyBuilder.WriteString(renderCategoryLine(category.Name, isSelected, m.width))
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
				bodyBuilder.WriteString(renderTaskLine(task, isSelected, m.width))
			}
			bodyBuilder.WriteString("\n")
			cursor++
		}

	}

	body := strings.TrimRight(bodyBuilder.String(), "\n")
	statusLine := m.statusLine()
	shortcuts := m.shortcutsLine()
	content := body + "\n\n" + statusLine + "\n" + shortcuts + "\n"
	if m.showHelp {
		help := m.helpView()
		if m.width > 0 && m.height > 0 {
			bg := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
			return placeOverlay(bg, help, m.width, m.height)
		}
		return help
	}
	if m.confirmDelete {
		dialog := m.confirmDeleteView()
		if m.width > 0 && m.height > 0 {
			bg := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
			return placeOverlay(bg, dialog, m.width, m.height)
		}
		return dialog
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
	for _, category := range project.Categories {
		tasks := append([]domain.Task(nil), category.Tasks...)
		domain.SortTasks(tasks)
		categories = append(categories, categoryView{
			Name:  category.Name,
			Tasks: tasks,
		})
	}
	positions := rebuildPositions(categories)
	return categories, positions
}
