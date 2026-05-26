package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/atotto/clipboard"
	tea "charm.land/bubbletea/v2"

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
	case openURLResultMsg:
		m.handleOpenURLResult(msg)
	case editorFinishedMsg:
		m.handleEditorFinished(msg)
		return m, nil
	case tea.FocusMsg:
		m.ui.Screen.WindowFocused = true
	case tea.BlurMsg:
		m.ui.Screen.WindowFocused = false
	case tea.KeyPressMsg:
		m.ui.Screen.StatusMsg = ""
		return m.handleKeyMsg(msg)
	default:
		return m.forwardToInput(msg)
	}
	return m, nil
}

func (m model) handleKeyMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch m.ui.Modes.Current() {
	case modes.ModeHelp:
		return m.handleHelpKey(msg)
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
	case modes.ModeURLPicker:
		return m.handleURLPickerKey(msg)
	case modes.ModeVisual:
		return m.handleVisualKey(msg)
	case modes.ModeEdit:
		cmd := m.handleEditKey(msg)
		return m, cmd
	case modes.ModeDescriptionEdit:
		return m.handleDescriptionEditKey(msg)
	case modes.ModeExternalEdit:
		return m, nil
	default:
		return m.handleNormalKey(msg)
	}
}

func (m model) handleHelpKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "?":
		m.ui.Modes.ToNormal()
		return m, nil
	case "j", "down":
		m.moveHelpFocus(1)
	case "k", "up":
		m.moveHelpFocus(-1)
	case "ctrl+d":
		m.moveHelpFocus(m.helpViewportHeight() / 2)
	case "ctrl+u":
		m.moveHelpFocus(-m.helpViewportHeight() / 2)
	case "enter":
		if len(helpFocusables) == 0 {
			return m, nil
		}
		idx := m.ui.Help.Focused
		if idx < 0 || idx >= len(helpFocusables) {
			return m, nil
		}
		row := helpRows[helpFocusables[idx]]
		if row.disabled {
			return m, nil
		}
		b := normalBindings[row.bindingIndex]
		m.ui.Modes.ToNormal()
		return m, b.action(&m)
	}
	return m, nil
}

func (m model) handleConfirmDeleteKey(msg tea.KeyPressMsg) model {
	switch msg.String() {
	case "y", "enter":
		switch m.ui.ConfirmDelete.Kind {
		case ConfirmDeleteProject:
			m.confirmDeleteProject()
		case ConfirmDeleteVisualRange:
			m.confirmDeleteVisualRange()
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

func (m model) handleOptionsKey(msg tea.KeyPressMsg) model {
	const optionCount = 2
	switch msg.String() {
	case "q", "esc", "enter":
		m.ui.Modes.ToNormal()
	case "j", "down":
		if m.ui.Options.selectedOption < optionCount-1 {
			m.ui.Options.selectedOption++
		}
	case "k", "up":
		if m.ui.Options.selectedOption > 0 {
			m.ui.Options.selectedOption--
		}
	case "space", "tab", "h", "l":
		m.toggleSelectedOption()
	}
	return m
}

func (m model) handleFilterKey(msg tea.KeyPressMsg) model {
	catCount := len(m.project.Categories)
	switch msg.String() {
	case "q", "f":
		m.ui.Filter.ResetToHub()
		m.ui.Modes.ToNormal()
		m.rebuildPositions()
	case "esc":
		if m.ui.Filter.View() == FilterViewHub {
			m.ui.Modes.ToNormal()
			m.rebuildPositions()
		} else {
			m.ui.Filter.SetView(FilterViewHub)
		}
	case "j", "down":
		m.ui.Filter.MoveDown(catCount)
	case "k", "up":
		m.ui.Filter.MoveUp()
	case "enter", "space":
		if m.ui.Filter.View() == FilterViewHub {
			m.openFilterHubSelection()
		} else {
			m.ui.Filter.ToggleSelected(m.project.Categories)
			m.rebuildPositions()
		}
	}
	return m
}

func (m *model) openFilterHubSelection() {
	switch m.ui.Filter.HubSelected() {
	case FilterHubStatus:
		m.ui.Filter.SetView(FilterViewStatus)
	case FilterHubPriority:
		m.ui.Filter.SetView(FilterViewPriority)
	case FilterHubCategory:
		m.ui.Filter.SetView(FilterViewCategory)
	case FilterHubClearAll:
		if m.ui.Filter.HasActiveFilter() {
			m.ui.Filter.ClearAll()
			m.rebuildPositions()
		}
	}
}

func (m model) handleInfoKey(msg tea.KeyPressMsg) model {
	switch msg.String() {
	case "i", "q", "esc":
		m.ui.Modes.ToNormal()
	}
	return m
}

func (m model) handleEstimatePickerKey(msg tea.KeyPressMsg) model {
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

func (m model) handleURLPickerKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.ui.Modes.ToNormal()
		return m, nil
	case "j", "down":
		m.ui.URLPicker.MoveDown()
		return m, nil
	case "k", "up":
		m.ui.URLPicker.MoveUp()
		return m, nil
	case "enter":
		url := m.ui.URLPicker.SelectedURL()
		m.ui.Modes.ToNormal()
		if url == "" {
			return m, nil
		}
		return m, openURL(url)
	}
	return m, nil
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
	case 1: // PriorityColor
		newValue := nextPriorityColor(m.deps.CfgManager.Get().PriorityColor)
		_ = m.deps.CfgManager.Update(func(cfg *config.Config) {
			cfg.PriorityColor = newValue
		})
	}
}

func nextPriorityColor(current string) string {
	switch current {
	case config.PriorityColorFull:
		return config.PriorityColorIcon
	case config.PriorityColorIcon:
		return config.PriorityColorNone
	default:
		return config.PriorityColorFull
	}
}

func (m model) handleNormalKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
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

func (m model) View() tea.View {
	v := tea.NewView(m.renderView())
	v.AltScreen = true
	v.MouseMode = tea.MouseModeNone
	v.ReportFocus = true
	return v
}

func (m model) renderView() string {
	if m.ui.Screen.Height == 0 {
		return ""
	}

	layout := m.buildLayout()
	viewport := NewViewport(layout, m.ui.Screen.Height, m.layoutConfig())
	viewport.ComputeVisibility(m.ui.Screen.ScrollOffset)

	var lines []string

	if viewport.HasMoreAbove {
		lines = append(lines, ui.MutedStyle.Render(scrollMoreAbove))
	}

	for i := viewport.VisibleStart; i < viewport.VisibleEnd; i++ {
		lines = append(lines, m.renderLayoutItem(layout.Items[i]))
	}

	if viewport.HasMoreBelow && viewport.VisibleEnd < len(layout.Items) {
		if partial := m.renderLayoutItemTruncated(layout.Items[viewport.VisibleEnd], viewport.RemainingContentHeight()); partial != "" {
			lines = append(lines, partial)
		}
	}

	if viewport.HasMoreBelow {
		lines = append(lines, ui.MutedStyle.Render(scrollMoreBelow))
	}

	content := strings.Join(lines, "\n")
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
	case modes.ModeURLPicker:
		return modal.Render(content, m.urlPickerView())
	case modes.ModeDescriptionEdit:
		return modal.Render(content, m.descriptionEditView())
	}
	return content
}

func (m model) renderLayoutItem(item LayoutItem) string {
	isCursor := item.PositionIndex >= 0 && item.PositionIndex == m.selected()
	inVisualRange := item.PositionIndex >= 0 && m.isInVisualRange(item.PositionIndex)
	// A row participates in the selection-style band if it is the cursor OR
	// inside the visual range. The cursor is rendered distinctly within the
	// range so the user can see which end will extend on j/k.
	isSelected := isCursor || inVisualRange
	visualMode := m.ui.Modes.IsVisual()
	focused := m.ui.Screen.WindowFocused

	switch item.Kind {
	case LayoutProject:
		if m.ui.Modes.IsEdit() && isCursor {
			return m.renderEditProjectLine()
		}
		return renderProjectLine(m.project.Name, isSelected, focused, m.ui.Filter.HasActiveFilter(), visualMode, m.statusText(), m.ui.Screen.Width)

	case LayoutCategory:
		category := m.project.Categories[item.CategoryIndex]
		if m.ui.Modes.IsEdit() && isCursor {
			return m.renderEditCategoryLine()
		}
		folded := m.ui.Fold.IsFolded(category.ID)
		cut := m.isCategoryCut(category.ID)
		return renderCategoryLine(category.Name, category.EstimateMinutes, category.AggregateStatus(), isSelected, folded, m.ui.Screen.Width, focused, inVisualRange, isCursor, cut)

	case LayoutTask:
		task := m.project.Categories[item.CategoryIndex].Tasks[item.TaskIndex]
		if m.ui.Modes.IsEdit() && isCursor {
			return m.renderEditTaskLine(task)
		}
		cut := m.isTaskCut(task.ID)
		return m.renderTaskLine(task, isSelected, m.ui.Screen.Width, focused, inVisualRange, isCursor, cut)

	case LayoutDescription:
		task := m.project.Categories[item.CategoryIndex].Tasks[item.TaskIndex]
		cut := m.isTaskCut(task.ID)
		return m.renderTaskDescription(task, isCursor, m.ui.Screen.Width, focused, cut)

	case LayoutEmptyCategory:
		return ui.MutedStyle.Render("    (no tasks)")

	case LayoutFolded:
		return ui.MutedStyle.Render("    (folded)")

	case LayoutSpacing:
		return strings.Repeat("\n", item.Height-1)
	}

	return ""
}

func (m model) renderLayoutItemTruncated(item LayoutItem, maxRows int) string {
	if maxRows <= 0 {
		return ""
	}
	full := m.renderLayoutItem(item)
	if full == "" {
		return ""
	}
	rendered := strings.Split(full, "\n")
	if len(rendered) <= maxRows {
		return full
	}
	return strings.Join(rendered[:maxRows], "\n")
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
	positions := rebuildPositions(project.Categories, nil, &foldState, false)
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

	program := tea.NewProgram(m)
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
