package app

import (
	"errors"
	"fmt"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"phasionary/internal/app/components"
	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
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
	project domain.Project
	ui      *UIState
	deps    *Dependencies
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ui.Width = msg.Width
		m.ui.Height = msg.Height
		m.ensureVisible()
	case clipboardResultMsg:
		if msg.err != nil {
			m.ui.StatusMsg = fmt.Sprintf("Copy failed: %v", msg.err)
		} else {
			m.ui.StatusMsg = "Copied!"
		}
	case editorFinishedMsg:
		m.handleEditorFinished(msg)
		return m, nil
	case tea.FocusMsg:
		m.ui.WindowFocused = true
	case tea.BlurMsg:
		m.ui.WindowFocused = false
	case tea.MouseMsg:
		if !m.ui.Modes.IsNormal() {
			break
		}
		if msg.Button != tea.MouseButtonLeft || msg.Action != tea.MouseActionPress {
			break
		}
		rowMap := m.computeRowMap()
		if msg.Y >= 0 && msg.Y < len(rowMap) {
			pos := rowMap[msg.Y]
			if pos >= 0 && pos < m.ui.Selection.Count() {
				m.ui.Selection.SetSelected(pos)
				m.ensureVisible()
			}
		}
	case tea.KeyMsg:
		m.ui.StatusMsg = ""
		return m.handleKeyMsg(msg)
	}
	return m, nil
}

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "?" {
		m.ui.Modes.ToggleHelp()
		return m, nil
	}
	switch m.ui.Modes.Current() {
	case modes.ModeHelp:
		return m.handleHelpKey(msg), nil
	case modes.ModeConfirmDelete:
		return m.handleConfirmDeleteKey(msg), nil
	case modes.ModeOptions:
		return m.handleOptionsKey(msg), nil
	case modes.ModeProjectPicker:
		return m.handleProjectPickerKey(msg)
	case modes.ModeFilter:
		return m.handleFilterKey(msg), nil
	case modes.ModeInfo:
		return m.handleInfoKey(msg), nil
	case modes.ModeEstimatePicker:
		return m.handleEstimatePickerKey(msg), nil
	case modes.ModeEdit:
		cmd := m.handleEditKey(msg)
		return m, cmd
	case modes.ModeExternalEdit:
		return m, nil
	default:
		return m.handleNormalKey(msg)
	}
}

func (m model) handleHelpKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "q", "esc":
		m.ui.Modes.ToNormal()
	}
	return m
}

func (m model) handleConfirmDeleteKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "y", "enter":
		if m.ui.Picker.pendingDeleteID != "" {
			m.confirmDeleteProject()
		} else {
			m.confirmDeleteAction()
		}
	case "n", "esc":
		if m.ui.Picker.pendingDeleteID != "" {
			m.ui.Picker.pendingDeleteID = ""
			m.ui.Modes.ToProjectPicker()
		} else {
			m.ui.Modes.ToNormal()
		}
	}
	return m
}

func (m model) handleOptionsKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "q", "esc", "enter":
		m.ui.Modes.ToNormal()
	case "j", "down":
		// Ready for more options
	case "k", "up":
		// Ready for more options
	case " ", "tab", "h", "l":
		m.toggleSelectedOption()
	}
	return m
}

func (m model) handleFilterKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "q", "esc", "f":
		m.ui.Modes.ToNormal()
		m.rebuildPositions()
	case "j", "down":
		m.ui.Filter.MoveDown()
	case "k", "up":
		m.ui.Filter.MoveUp()
	case " ":
		m.ui.Filter.ToggleSelected()
	}
	return m
}

func (m model) handleInfoKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "i", "q", "esc":
		m.ui.Modes.ToNormal()
	}
	return m
}

func (m model) handleEstimatePickerKey(msg tea.KeyMsg) model {
	switch msg.String() {
	case "q", "esc":
		m.ui.Modes.ToNormal()
	case "j", "down":
		m.ui.EstimatePicker.MoveDown()
	case "k", "up":
		m.ui.EstimatePicker.MoveUp()
	case "enter":
		m.selectEstimate(m.ui.EstimatePicker.SelectedValue())
		m.ui.Modes.ToNormal()
	}
	return m
}

func (m *model) toggleSelectedOption() {
	switch m.ui.Options.selectedOption {
	case 0: // StatusDisplay
		newValue := config.StatusDisplayIcons
		if m.deps.CfgManager.Get().StatusDisplay == config.StatusDisplayIcons {
			newValue = config.StatusDisplayText
		}
		_ = m.deps.CfgManager.Update(func(cfg *config.Config) {
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
		m.ui.PendingKey = 0
	case "down", "j":
		m.moveSelection(1)
		m.ui.PendingKey = 0
	case " ":
		m.toggleSelectedTask()
		m.ui.PendingKey = 0
	case "enter":
		m.startEditing()
		m.ui.PendingKey = 0
	case "a":
		m.startAddingTask()
		m.ui.PendingKey = 0
	case "A":
		m.startAddingCategory()
		m.ui.PendingKey = 0
	case "h":
		m.decreasePriority()
		m.ui.PendingKey = 0
	case "l":
		m.increasePriority()
		m.ui.PendingKey = 0
	case "y":
		m.ui.PendingKey = 0
		return m, m.copySelected()
	case "d":
		m.deleteSelected()
		m.ui.PendingKey = 0
	case "J":
		if pos, ok := m.selectedPosition(); ok && pos.Kind == focusCategory {
			m.moveCategoryDown()
		} else {
			m.moveTaskDown()
		}
		m.ui.PendingKey = 0
	case "K":
		if pos, ok := m.selectedPosition(); ok && pos.Kind == focusCategory {
			m.moveCategoryUp()
		} else {
			m.moveTaskUp()
		}
		m.ui.PendingKey = 0
	case "ctrl+d":
		m.moveSelectionByPage(0.5)
		m.ui.PendingKey = 0
	case "ctrl+u":
		m.moveSelectionByPage(-0.5)
		m.ui.PendingKey = 0
	case "ctrl+f":
		m.moveSelectionByPage(1.0)
		m.ui.PendingKey = 0
	case "ctrl+b":
		m.moveSelectionByPage(-1.0)
		m.ui.PendingKey = 0
	case "g":
		if m.ui.PendingKey == 'g' {
			m.jumpToFirst()
			m.ui.PendingKey = 0
		} else {
			m.ui.PendingKey = 'g'
		}
	case "G":
		m.jumpToLast()
		m.ui.PendingKey = 0
	case "o":
		m.ui.Modes.ToOptions()
		m.ui.Options = OptionsState{selectedOption: 0}
		m.ui.PendingKey = 0
	case "p":
		m.openProjectPicker()
		m.ui.PendingKey = 0
	case "z":
		if m.ui.PendingKey == 'z' {
			m.centerOnSelected()
			m.ui.PendingKey = 0
		} else {
			m.ui.PendingKey = 'z'
		}
	case "s":
		m.sortTasksByStatus()
		m.ui.PendingKey = 0
	case "S":
		m.sortTasksByStatusReverse()
		m.ui.PendingKey = 0
	case "f":
		m.ui.Modes.ToFilter()
		m.ui.PendingKey = 0
	case "e":
		m.ui.PendingKey = 0
		return m, m.startExternalEdit()
	case "i":
		m.ui.Modes.ToInfo()
		m.ui.PendingKey = 0
	case "t":
		m.openEstimatePicker()
		m.ui.PendingKey = 0
	default:
		m.ui.PendingKey = 0
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
	if m.ui.Height == 0 {
		return ""
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Height, DefaultLayoutConfig())
	viewport.ComputeVisibility(m.ui.ScrollOffset)

	var lines []string

	if viewport.HasMoreAbove {
		lines = append(lines, ui.MutedStyle.Render("  ↑ more above"))
	}

	for i := viewport.VisibleStart; i < viewport.VisibleEnd; i++ {
		lines = append(lines, m.renderLayoutItem(layout.Items[i]))
	}

	if viewport.HasMoreBelow {
		lines = append(lines, ui.MutedStyle.Render("  ↓ more below"))
	}

	body := strings.Join(lines, "\n")

	statusLine := m.statusLine()
	shortcuts := m.shortcutsLine()
	content := body + "\n\n" + statusLine + "\n" + shortcuts
	modal := components.NewModal(m.ui.Width, m.ui.Height)
	switch m.ui.Modes.Current() {
	case modes.ModeHelp:
		return modal.Render(content, m.helpView())
	case modes.ModeConfirmDelete:
		return modal.Render(content, m.confirmDeleteView())
	case modes.ModeOptions:
		return modal.Render(content, m.optionsView())
	case modes.ModeProjectPicker:
		return modal.Render(content, m.projectPickerView())
	case modes.ModeFilter:
		return modal.Render(content, m.filterView())
	case modes.ModeInfo:
		return modal.Render(content, m.infoView())
	case modes.ModeEstimatePicker:
		return modal.Render(content, m.estimatePickerView())
	}
	return content
}

func (m model) renderLayoutItem(item LayoutItem) string {
	isSelected := item.PositionIndex >= 0 && item.PositionIndex == m.selected()
	focused := m.ui.WindowFocused

	switch item.Kind {
	case LayoutProject:
		if m.ui.Modes.IsEdit() && isSelected {
			return m.renderEditProjectLine()
		}
		return renderProjectLine(m.project.Name, isSelected, focused)

	case LayoutCategory:
		category := m.project.Categories[item.CategoryIndex]
		if m.ui.Modes.IsEdit() && isSelected {
			return m.renderEditCategoryLine()
		}
		return renderCategoryLine(category.Name, category.EstimateMinutes, isSelected, m.ui.Width, focused)

	case LayoutTask:
		task := m.project.Categories[item.CategoryIndex].Tasks[item.TaskIndex]
		if m.ui.Modes.IsEdit() && isSelected {
			return m.renderEditTaskLine(task)
		}
		return m.renderTaskLine(task, isSelected, m.ui.Width, focused)

	case LayoutEmptyCategory:
		return ui.MutedStyle.Render("  (no tasks)")

	case LayoutSpacing:
		return strings.Repeat("\n", item.Height-1)
	}

	return ""
}

func Run(dataDir string, projectSelector string, cfgManager *config.Manager, workingDir string) error {
	store := data.NewStore(dataDir)
	if err := store.Ensure(); err != nil {
		return err
	}

	stateManager := data.NewStateManager(dataDir, workingDir)
	if err := stateManager.Load(); err != nil {
		return err
	}

	projects, err := store.ListProjects()
	if err != nil {
		return err
	}

	var project domain.Project
	startMode := modes.ModeNormal

	if len(projects) == 0 {
		project, err = store.InitDefault()
		if err != nil {
			return err
		}
		_ = stateManager.SetLastProjectID(project.ID)
	} else if projectSelector != "" {
		project, err = store.LoadProject(projectSelector)
		if err != nil {
			if errors.Is(err, data.ErrProjectNotFound) {
				return fmt.Errorf("project %q not found", projectSelector)
			}
			return err
		}
		_ = stateManager.SetLastProjectID(project.ID)
	} else if lastID := stateManager.GetLastProjectID(); lastID != "" {
		project, err = store.LoadProject(lastID)
		if err != nil {
			if errors.Is(err, data.ErrProjectNotFound) {
				startMode = modes.ModeProjectPicker
			} else {
				return err
			}
		} else {
			_ = stateManager.SetLastProjectID(project.ID)
		}
	} else {
		startMode = modes.ModeProjectPicker
	}

	positions := rebuildPositions(project.Categories, nil)
	initialSelection := findFirstTaskIndex(positions)
	selMgr := selection.NewManager(toSelectionPositions(positions), initialSelection)
	modeMachine := modes.NewMachine(startMode)

	m := model{
		project: project,
		ui:      NewUIState(selMgr, modeMachine),
		deps:    NewDependencies(store, cfgManager, stateManager),
	}

	if startMode == modes.ModeProjectPicker {
		m.ui.Picker = ProjectPickerState{
			projects:     projects,
			selected:     0,
			scrollOffset: 0,
		}
	}

	program := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion(), tea.WithReportFocus())
	_, err = program.Run()
	return err
}

func findFirstTaskIndex(positions []focusPosition) int {
	for i, pos := range positions {
		if pos.Kind == focusTask {
			return i
		}
	}
	if len(positions) > 0 {
		return 0
	}
	return -1
}
