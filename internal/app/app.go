package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	tea "github.com/charmbracelet/bubbletea"

	"phasionary/internal/app/components"
	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
	"phasionary/internal/config"
	"phasionary/internal/data"
	"phasionary/internal/domain"
	"phasionary/internal/export"
	"phasionary/internal/ui"
)

type clipboardResultMsg struct{ err error }

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
		m.ui.Screen.Width = msg.Width
		m.ui.Screen.Height = msg.Height
		m.ensureVisible()
	case clipboardResultMsg:
		if msg.err != nil {
			m.ui.Screen.StatusMsg = fmt.Sprintf("Copy failed: %v", msg.err)
		} else {
			m.ui.Screen.StatusMsg = "Copied!"
		}
	case editorFinishedMsg:
		m.handleEditorFinished(msg)
		return m, nil
	case tea.FocusMsg:
		m.ui.Screen.WindowFocused = true
	case tea.BlurMsg:
		m.ui.Screen.WindowFocused = false
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
		m.ui.Screen.StatusMsg = ""
		return m.handleKeyMsg(msg)
	default:
		return m.forwardToInput(msg)
	}
	return m, nil
}

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
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
		switch m.ui.ConfirmDelete.Kind {
		case ConfirmDeleteProject:
			m.confirmDeleteProject()
		default:
			m.confirmDeleteAction()
		}
	case "n", "esc":
		switch m.ui.ConfirmDelete.Kind {
		case ConfirmDeleteProject:
			m.ui.ConfirmDelete.reset()
			m.ui.Modes.ToProjectPicker()
		default:
			m.ui.ConfirmDelete.reset()
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
	cmd := m.dispatchNormalKey(msg.String())
	return m, cmd
}

func (m *model) copySelected() tea.Cmd {
	pos, ok := m.selectedPosition()
	if !ok {
		return nil
	}
	var text string
	switch pos.Kind {
	case selection.FocusProject:
		text = m.project.Name
	case selection.FocusCategory:
		text = m.project.Categories[pos.CategoryIndex].Name
	case selection.FocusTask:
		task := m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex]
		text = task.Title
		taskCopy := task
		m.ui.Clipboard = ClipboardState{
			Task:     &taskCopy,
			IsCut:    false,
			SourceID: "",
		}
	}
	return func() tea.Msg {
		return clipboardResultMsg{err: clipboard.WriteAll(text)}
	}
}

func (m *model) copyCategoryContent() tea.Cmd {
	pos, ok := m.selectedPosition()
	if !ok || pos.Kind == selection.FocusProject {
		return nil
	}
	text := export.ExportCategoryMarkdown(m.project.Categories[pos.CategoryIndex])
	return func() tea.Msg {
		return clipboardResultMsg{err: clipboard.WriteAll(text)}
	}
}

func (m model) View() string {
	if m.ui.Screen.Height == 0 {
		return ""
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Screen.Height, DefaultLayoutConfig())
	viewport.ComputeVisibility(m.ui.Screen.ScrollOffset)

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
	content := body + "\n\n" + statusLine
	modal := components.NewModal(m.ui.Screen.Width, m.ui.Screen.Height)
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
	focused := m.ui.Screen.WindowFocused

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
		folded := m.ui.Fold.IsFolded(category.ID)
		return renderCategoryLine(category.Name, category.EstimateMinutes, category.AggregateStatus(), isSelected, folded, m.ui.Screen.Width, focused)

	case LayoutTask:
		task := m.project.Categories[item.CategoryIndex].Tasks[item.TaskIndex]
		if m.ui.Modes.IsEdit() && isSelected {
			return m.renderEditTaskLine(task)
		}
		return m.renderTaskLine(task, isSelected, m.ui.Screen.Width, focused)

	case LayoutEmptyCategory:
		return ui.MutedStyle.Render("    (no tasks)")

	case LayoutFolded:
		return ui.MutedStyle.Render("    (folded)")

	case LayoutSpacing:
		return strings.Repeat("\n", item.Height-1)
	}

	return ""
}

func Run(dataDir string, projectSelector string, cfgManager config.Reader, workingDir string) error {
	store := data.NewStore(dataDir)
	if err := store.Ensure(); err != nil {
		return err
	}

	stateManager := data.NewStateManager(filepath.Dir(dataDir), workingDir)
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

	foldState := NewFoldStateFrom(stateManager.GetFoldedCategories(project.ID))
	positions := rebuildPositions(project.Categories, nil, &foldState)
	initialSelection := findFirstTaskIndex(positions)
	selMgr := selection.NewManager(positions, initialSelection)
	modeMachine := modes.NewMachine(startMode)

	m := model{
		project: project,
		ui:      NewUIState(selMgr, modeMachine),
		deps:    NewDependencies(store, cfgManager, stateManager),
	}
	m.ui.Fold = foldState

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

func findFirstTaskIndex(positions []selection.Position) int {
	for i, pos := range positions {
		if pos.Kind == selection.FocusTask {
			return i
		}
	}
	if len(positions) > 0 {
		return 0
	}
	return -1
}
