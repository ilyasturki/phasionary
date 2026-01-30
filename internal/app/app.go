package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"phasionary/internal/config"
	"phasionary/internal/data"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

type clipboardResultMsg struct{ err error }


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
	project      domain.Project
	positions    []focusPosition
	selected     int
	width        int
	height       int
	mode         UIMode
	edit         EditState
	picker       ProjectPickerState
	store        data.ProjectRepository
	cfgManager   *config.Manager
	options      OptionsState
	statusMsg    string
	scrollOffset int
	pendingKey   rune
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ensureVisible()
	case clipboardResultMsg:
		if msg.err != nil {
			m.statusMsg = fmt.Sprintf("Copy failed: %v", msg.err)
		} else {
			m.statusMsg = "Copied!"
		}
	case tea.MouseMsg:
		if m.mode != ModeNormal {
			break
		}
		if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
			break
		}
		rowMap := m.computeRowMap()
		if msg.Y >= 0 && msg.Y < len(rowMap) {
			pos := rowMap[msg.Y]
			if pos >= 0 && pos < len(m.positions) {
				m.selected = pos
				m.ensureVisible()
			}
		}
	case tea.KeyMsg:
		m.statusMsg = ""
		return m.handleKeyMsg(msg)
	}
	return m, nil
}

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "?" {
		if m.mode == ModeHelp {
			m.mode = ModeNormal
		} else if m.mode == ModeNormal {
			m.mode = ModeHelp
		}
		return m, nil
	}
	switch m.mode {
	case ModeHelp:
		return m.handleHelpKey(msg), nil
	case ModeConfirmDelete:
		return m.handleConfirmDeleteKey(msg), nil
	case ModeOptions:
		return m.handleOptionsKey(msg), nil
	case ModeProjectPicker:
		return m.handleProjectPickerKey(msg), nil
	case ModeEdit:
		cmd := m.handleEditKey(msg)
		return m, cmd
	default:
		return m.handleNormalKey(msg)
	}
}

func (m model) handleHelpKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "q", "esc":
		m.mode = ModeNormal
	}
	return m
}

func (m model) handleConfirmDeleteKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "y", "enter":
		m.confirmDeleteAction()
	case "n", "esc":
		m.mode = ModeNormal
	}
	return m
}

func (m model) handleOptionsKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "q", "esc", "enter":
		m.mode = ModeNormal
	case "j", "down":
		// Ready for more options
	case "k", "up":
		// Ready for more options
	case " ", "tab", "h", "l":
		m.toggleSelectedOption()
	}
	return m
}

func (m *model) toggleSelectedOption() {
	switch m.options.selectedOption {
	case 0: // StatusDisplay
		newValue := config.StatusDisplayIcons
		if m.cfgManager.Get().StatusDisplay == config.StatusDisplayIcons {
			newValue = config.StatusDisplayText
		}
		_ = m.cfgManager.Update(func(cfg *config.Config) {
			cfg.StatusDisplay = newValue
		})
	}
}

func (m model) handleNormalKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "ctrl+c":
		return m, tea.Quit
	case "up", "k":
		m.moveSelection(-1)
		m.pendingKey = 0
	case "down", "j":
		m.moveSelection(1)
		m.pendingKey = 0
	case " ":
		m.toggleSelectedTask()
		m.pendingKey = 0
	case "enter":
		m.startEditing()
		m.pendingKey = 0
	case "a":
		m.startAddingTask()
		m.pendingKey = 0
	case "A":
		m.startAddingCategory()
		m.pendingKey = 0
	case "h":
		m.decreasePriority()
		m.pendingKey = 0
	case "l":
		m.increasePriority()
		m.pendingKey = 0
	case "y":
		m.pendingKey = 0
		return m, m.copySelected()
	case "d":
		m.deleteSelected()
		m.pendingKey = 0
	case "J":
		if pos, ok := m.selectedPosition(); ok && pos.Kind == focusCategory {
			m.moveCategoryDown()
		} else {
			m.moveTaskDown()
		}
		m.pendingKey = 0
	case "K":
		if pos, ok := m.selectedPosition(); ok && pos.Kind == focusCategory {
			m.moveCategoryUp()
		} else {
			m.moveTaskUp()
		}
		m.pendingKey = 0
	case "ctrl+d":
		m.moveSelectionByPage(0.5)
		m.pendingKey = 0
	case "ctrl+u":
		m.moveSelectionByPage(-0.5)
		m.pendingKey = 0
	case "ctrl+f":
		m.moveSelectionByPage(1.0)
		m.pendingKey = 0
	case "ctrl+b":
		m.moveSelectionByPage(-1.0)
		m.pendingKey = 0
	case "g":
		if m.pendingKey == 'g' {
			m.jumpToFirst()
			m.pendingKey = 0
		} else {
			m.pendingKey = 'g'
		}
	case "G":
		m.jumpToLast()
		m.pendingKey = 0
	case "o":
		m.mode = ModeOptions
		m.options = OptionsState{selectedOption: 0}
		m.pendingKey = 0
	case "p":
		m.openProjectPicker()
		m.pendingKey = 0
	case "z":
		if m.pendingKey == 'z' {
			m.centerOnSelected()
			m.pendingKey = 0
		} else {
			m.pendingKey = 'z'
		}
	default:
		m.pendingKey = 0
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
		text = m.project.Categories[pos.CategoryIndex].Name
	default:
		text = m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex].Title
	}
	return func() tea.Msg {
		return clipboardResultMsg{err: clipboard.WriteAll(text)}
	}
}

func (m model) View() string {
	layout := m.buildLayout()
	viewport := NewViewport(layout, m.height, DefaultLayoutConfig())
	viewport.ComputeVisibility(m.scrollOffset)

	var lines []string

	if viewport.HasMoreAbove {
		lines = append(lines, ui.MutedStyle.Render("  ↑ more above"))
	}

	for i := viewport.VisibleStart; i < viewport.VisibleEnd; i++ {
		rendered := m.renderLayoutItem(layout.Items[i])
		if rendered != "" {
			lines = append(lines, rendered)
		}
	}

	if viewport.HasMoreBelow {
		lines = append(lines, ui.MutedStyle.Render("  ↓ more below"))
	}

	body := strings.Join(lines, "\n")

	statusLine := m.statusLine()
	shortcuts := m.shortcutsLine()
	content := body + "\n\n" + statusLine + "\n" + shortcuts + "\n"
	switch m.mode {
	case ModeHelp:
		help := m.helpView()
		if m.width > 0 && m.height > 0 {
			bg := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
			return placeOverlay(bg, help, m.width, m.height)
		}
		return help
	case ModeConfirmDelete:
		dialog := m.confirmDeleteView()
		if m.width > 0 && m.height > 0 {
			bg := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
			return placeOverlay(bg, dialog, m.width, m.height)
		}
		return dialog
	case ModeOptions:
		dialog := m.optionsView()
		if m.width > 0 && m.height > 0 {
			bg := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
			return placeOverlay(bg, dialog, m.width, m.height)
		}
		return dialog
	case ModeProjectPicker:
		dialog := m.projectPickerView()
		if m.width > 0 && m.height > 0 {
			bg := lipgloss.Place(m.width, m.height, lipgloss.Left, lipgloss.Top, content)
			return placeOverlay(bg, dialog, m.width, m.height)
		}
		return dialog
	}
	return content
}

func (m model) renderLayoutItem(item LayoutItem) string {
	isSelected := item.PositionIndex >= 0 && item.PositionIndex == m.selected

	switch item.Kind {
	case LayoutProject:
		if m.mode == ModeEdit && isSelected {
			return m.renderEditProjectLine()
		}
		return renderProjectLine(m.project.Name, isSelected)

	case LayoutCategory:
		category := m.project.Categories[item.CategoryIndex]
		if m.mode == ModeEdit && isSelected {
			return m.renderEditCategoryLine()
		}
		return renderCategoryLine(category.Name, isSelected, m.width)

	case LayoutTask:
		task := m.project.Categories[item.CategoryIndex].Tasks[item.TaskIndex]
		if m.mode == ModeEdit && isSelected {
			return m.renderEditTaskLine(task)
		}
		return m.renderTaskLine(task, isSelected, m.width)

	case LayoutEmptyCategory:
		return ui.MutedStyle.Render("  (no tasks)")

	case LayoutSpacing:
		return strings.Repeat("\n", item.Height-1)
	}

	return ""
}

func Run(dataDir string, projectSelector string, cfgManager *config.Manager) error {
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

	positions := rebuildPositions(project.Categories)
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
		positions:  positions,
		selected:   selected,
		store:      store,
		cfgManager: cfgManager,
	}, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = program.Run()
	return err
}
