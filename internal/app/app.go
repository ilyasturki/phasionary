package app

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

	"phasionary/internal/app/components"
	"phasionary/internal/app/modes"
	"phasionary/internal/app/selection"
	"phasionary/internal/config"
	"phasionary/internal/data"
	"phasionary/internal/domain"
	"phasionary/internal/export"
	"phasionary/internal/ui"
)

type clipboardResultMsg struct {
	err error
	// label, when set, names what was copied (e.g. "UUID: 550e..."); empty
	// falls back to a generic "Copied!".
	label string
}

// saveErrMsg reports a failed background save from the async saver. Successful
// writes send nothing; a failure surfaces here as a status message.
type saveErrMsg struct{ err error }

type model struct {
	project domain.Project
	ui      *UIState
	deps    *Dependencies
}

func (m model) Init() tea.Cmd {
	return m.listenSaveErrors()
}

// listenSaveErrors subscribes to the async saver's error channel so a failed
// background write becomes a visible status message. It resolves one delivery
// then must be re-issued (see the saveErrMsg case) to keep listening.
func (m model) listenSaveErrors() tea.Cmd {
	if m.deps.Saver == nil {
		return nil
	}
	results := m.deps.Saver.Results()
	return func() tea.Msg {
		err, ok := <-results
		if !ok {
			return nil
		}
		return saveErrMsg{err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.ui.Screen.Width = msg.Width
		m.ui.Screen.Height = msg.Height
		m.ensureVisible()
		if m.ui.Modes.IsProjectPicker() {
			m.ui.Picker.ensureVisible(m.pickerVisibleCount())
		}
	case clipboardResultMsg:
		if msg.err != nil {
			m.ui.Screen.StatusMsg = fmt.Sprintf("Copy failed: %v", msg.err)
		} else if msg.label != "" {
			m.ui.Screen.StatusMsg = "Copied " + msg.label
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
	case tea.MouseWheelMsg:
		if m.ui.Modes.Current() == modes.ModeNormal {
			m.handleMouseWheel(msg)
		}
		return m, nil
	case saveErrMsg:
		if msg.err != nil {
			m.ui.Screen.StatusMsg = "Save failed: " + msg.err.Error()
		}
		return m, m.listenSaveErrors()
	case tea.KeyPressMsg:
		m.ui.Screen.StatusMsg = ""
		return m.handleKeyMsg(msg)
	default:
		return m.forwardToInput(msg)
	}
	return m, nil
}

func (m *model) handleMouseWheel(msg tea.MouseWheelMsg) {
	switch msg.Button {
	case tea.MouseWheelUp:
		m.scrollUp(wheelScrollStep)
	case tea.MouseWheelDown:
		m.scrollDown(wheelScrollStep)
	}
}

// normalizeKey rewrites Shift+Backspace to a plain Backspace so it deletes a
// character wherever the user types. Terminals speaking the enhanced keyboard
// protocol deliver Shift+Backspace as a distinct key, but bubbles' text inputs
// bind delete-backward to "backspace"/"ctrl+h" only, so the modified form is
// otherwise silently ignored. Normal mode binds nothing to backspace, so this
// is a no-op outside the text-entry modes.
func normalizeKey(msg tea.KeyPressMsg) tea.KeyPressMsg {
	if msg.Code == tea.KeyBackspace && msg.Mod&tea.ModShift != 0 {
		msg.Mod &^= tea.ModShift
	}
	return msg
}

func (m model) handleKeyMsg(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	msg = normalizeKey(msg)
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
	case modes.ModeYankPicker:
		return m.handleYankPickerKey(msg)
	case modes.ModeSearch:
		return m.handleSearchKey(msg)
	case modes.ModeVisual:
		return m.handleVisualKey(msg)
	case modes.ModeEdit:
		cmd := m.handleEditKey(msg)
		return m, cmd
	case modes.ModeDescriptionEdit:
		return m.handleDescriptionEditKey(msg)
	case modes.ModeTagEdit:
		return m.handleTagEditKey(msg)
	case modes.ModeExternalEdit:
		return m, nil
	default:
		return m.handleNormalKey(msg)
	}
}

func (m model) handleHelpKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if m.ui.Help.Filtering {
		return m.handleHelpFilterKey(msg)
	}
	switch msg.String() {
	case "q", "esc", "?":
		m.ui.Modes.ToNormal()
		return m, nil
	case "/":
		cmd := m.startHelpFilter()
		return m, cmd
	case "j", "down":
		m.moveHelpFocus(1)
	case "k", "up":
		m.moveHelpFocus(-1)
	case "ctrl+d":
		m.moveHelpFocus(m.helpViewportHeight() / 2)
	case "ctrl+u":
		m.moveHelpFocus(-m.helpViewportHeight() / 2)
	case "enter":
		return m.runFocusedHelpBinding()
	}
	return m, nil
}

// handleHelpFilterKey handles keys while the `/` filter is active: Esc clears the
// filter (Esc again closes the dialog), Enter runs the focused match, arrows and
// ctrl+d/u navigate, and everything else edits the query.
func (m model) handleHelpFilterKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.ui.Help.Filtering = false
		m.ui.Help.Filter = textinput.Model{}
		m.ui.Help.Focused = 0
		m.ui.Help.ScrollOffset = 0
		m.ensureHelpVisible()
		return m, nil
	case "enter":
		return m.runFocusedHelpBinding()
	case "ctrl+c":
		return m, tea.Quit
	case "down":
		m.moveHelpFocus(1)
		return m, nil
	case "up":
		m.moveHelpFocus(-1)
		return m, nil
	case "ctrl+d":
		m.moveHelpFocus(m.helpViewportHeight() / 2)
		return m, nil
	case "ctrl+u":
		m.moveHelpFocus(-m.helpViewportHeight() / 2)
		return m, nil
	}
	var cmd tea.Cmd
	m.ui.Help.Filter, cmd = m.ui.Help.Filter.Update(msg)
	sanitizeInput(&m.ui.Help.Filter)
	m.ui.Help.Focused = 0
	m.ui.Help.ScrollOffset = 0
	m.ensureHelpVisible()
	return m, cmd
}

// startHelpFilter enters the `/` filter sub-mode with a fresh, focused input.
func (m *model) startHelpFilter() tea.Cmd {
	ti := textinput.New()
	cmd := ti.Focus()
	m.ui.Help.Filter = ti
	m.ui.Help.Filtering = true
	m.ui.Help.Focused = 0
	m.ui.Help.ScrollOffset = 0
	m.ensureHelpVisible()
	return cmd
}

// runFocusedHelpBinding runs the binding under the focused row (if any) and
// closes the dialog. Display-only rows aren't focusable, so this only ever fires
// a runnable Navigation/Actions binding.
func (m model) runFocusedHelpBinding() (tea.Model, tea.Cmd) {
	rows, focusables := m.currentHelpRows()
	if len(focusables) == 0 {
		return m, nil
	}
	idx := m.ui.Help.Focused
	if idx < 0 || idx >= len(focusables) {
		return m, nil
	}
	row := rows[focusables[idx]]
	if row.disabled {
		return m, nil
	}
	b := normalBindings[row.bindingIndex]
	m.ui.Modes.ToNormal()
	return m, b.action(&m)
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
	const optionCount = 4
	switch msg.String() {
	case "q", "esc", "enter":
		m.ui.Modes.ToNormal()
		// Re-clamp scroll now that the shortcut bar (which is hidden while
		// Options was open) may have toggled visible/hidden — its row count
		// changes available content height.
		m.ensureVisible()
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
	case FilterHubTag:
		m.ui.Filter.SetView(FilterViewTag)
	case FilterHubClearAll:
		if m.ui.Filter.HasActiveFilter() {
			m.ui.Filter.ClearAll()
			m.rebuildPositions()
		}
	}
}

func (m model) handleInfoKey(msg tea.KeyPressMsg) model {
	switch msg.String() {
	case "q", "esc":
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

func (m model) handleYankPickerKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc":
		m.ui.Modes.ToNormal()
		return m, nil
	case "j", "down":
		m.ui.YankPicker.MoveDown()
		return m, nil
	case "k", "up":
		m.ui.YankPicker.MoveUp()
		return m, nil
	case "enter", "y":
		it, ok := m.ui.YankPicker.SelectedItem()
		m.ui.Modes.ToNormal()
		if !ok {
			return m, nil
		}
		return m, copyYankItem(it)
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
		// Icons vs. text change the status column width, so cached row heights
		// are stale. (PriorityColor below only recolors — layout is unaffected.)
		m.invalidateLayout()
	case 1: // PriorityColor
		newValue := nextPriorityColor(m.deps.CfgManager.Get().PriorityColor)
		_ = m.deps.CfgManager.Update(func(cfg *config.Config) {
			cfg.PriorityColor = newValue
		})
	case 2: // ShowShortcutBar
		newValue := !m.deps.CfgManager.Get().ShowShortcutBar
		_ = m.deps.CfgManager.Update(func(cfg *config.Config) {
			cfg.ShowShortcutBar = newValue
		})
		// The bar is hidden while Options is open, so the layout under us
		// hasn't actually changed yet. handleOptionsKey's exit branch calls
		// ensureVisible against the post-toggle layout once Options closes.
	case 3: // ExpandDescriptionsByDefault
		newValue := !m.deps.CfgManager.Get().ExpandDescriptionsByDefault
		_ = m.deps.CfgManager.Update(func(cfg *config.Config) {
			cfg.ExpandDescriptionsByDefault = newValue
		})
		m.ui.Screen.ExpandDescriptions = newValue
		m.rebuildPositions()
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
		m.ui.TagCopiedLast = false
	case selection.FocusSeparator:
		// Copy the label text to the system clipboard, but don't stash the
		// separator as a paste-able item (it isn't part of the cut/paste model).
		text = m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex].Title
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
	v.MouseMode = tea.MouseModeCellMotion
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
	if bar := m.renderBottomBar(); bar != "" {
		// Push the bar to the bottom row even when content is shorter than the
		// screen. The viewport already reserved this row via FooterHeight, so
		// the gap calculation below accounts for the bar itself.
		rendered := strings.Count(content, "\n") + 1
		if content == "" {
			rendered = 0
		}
		gap := m.ui.Screen.Height - rendered - 1
		if gap > 0 {
			content += strings.Repeat("\n", gap)
		}
		if content != "" {
			content += "\n"
		}
		content += bar
	}
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
	case modes.ModeYankPicker:
		return modal.Render(content, m.yankPickerView())
	case modes.ModeDescriptionEdit:
		return modal.Render(content, m.descriptionEditView())
	case modes.ModeTagEdit:
		return modal.Render(content, m.tagEditView())
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
		return renderCategoryLine(category.Name, category.EstimateMinutes, category.AggregateStatus(), isSelected, folded, m.ui.Screen.Width, focused, inVisualRange, isCursor, cut, m.searchQuery(), m.searchMatchStyle(isCursor))

	case LayoutTask:
		task := m.project.Categories[item.CategoryIndex].Tasks[item.TaskIndex]
		if m.ui.Modes.IsEdit() && isCursor {
			return m.renderEditTaskLine(task)
		}
		cut := m.isTaskCut(task.ID)
		return m.renderTaskLine(task, isSelected, m.ui.Screen.Width, focused, inVisualRange, isCursor, cut)

	case LayoutSeparator:
		if m.ui.Modes.IsEdit() && isCursor {
			return m.renderEditSeparatorLine()
		}
		label := m.project.Categories[item.CategoryIndex].Tasks[item.TaskIndex].Title
		return m.renderSeparatorLine(label, isSelected, focused, visualMode, isCursor, m.ui.Screen.Width, m.searchQuery(), m.searchMatchStyle(isCursor))

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

// linkDirIfUnset establishes the directory→project link only when the directory
// has none yet, so a session-only project switch (--project, picker select)
// never overrides an existing link. Deliberate relinking (`project link`/`add`)
// calls SetProjectForDir directly.
func linkDirIfUnset(sm data.StateRepository, projectID string) {
	if sm.GetProjectForDir() == "" {
		_ = sm.SetProjectForDir(projectID)
	}
}

func Run(dataDir string, projectSelector string, cfgManager config.Reader, workingDir string, forcePicker bool) error {
	store := data.NewStore(dataDir)
	if err := store.Ensure(); err != nil {
		return err
	}

	// Persist off the event loop so no keystroke blocks on fsync. Deferring
	// Close here means the final edit is flushed on exit regardless of which
	// quit path the program took — program.Run only returns once the event loop
	// has stopped, so no Enqueue can race this Close.
	saver := data.NewSaver(store)
	defer saver.Close()

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
		_ = stateManager.SetProjectForDir(project.ID)
		projects = []domain.Project{project}
	} else if projectSelector != "" {
		project, err = store.LoadProject(projectSelector)
		if err != nil {
			if errors.Is(err, data.ErrProjectNotFound) {
				return fmt.Errorf("project %q not found", projectSelector)
			}
			return err
		}
		// --project opens a project for this session only: link the directory
		// when it has no link yet, but never override an existing one.
		linkDirIfUnset(stateManager, project.ID)
	} else if linkedID := stateManager.GetProjectForDir(); linkedID != "" {
		project, err = store.LoadProject(linkedID)
		if err != nil {
			if errors.Is(err, data.ErrProjectNotFound) {
				startMode = modes.ModeProjectPicker
			} else {
				return err
			}
		} else {
			_ = stateManager.SetProjectForDir(project.ID)
		}
	} else {
		startMode = modes.ModeProjectPicker
	}

	if forcePicker {
		startMode = modes.ModeProjectPicker
	}

	foldState := NewFoldStateFrom(stateManager.GetFoldedCategories(project.ID))
	expandDescriptions := cfgManager.Get().ExpandDescriptionsByDefault
	positions := rebuildPositions(project.Categories, nil, &foldState, expandDescriptions)
	initialSelection := findFirstTaskIndex(positions)
	selMgr := selection.NewManager(positions, initialSelection)
	modeMachine := modes.NewMachine(startMode)

	deps := NewDependencies(store, cfgManager, stateManager)
	deps.Saver = saver
	m := model{
		project: project,
		ui:      NewUIState(selMgr, modeMachine),
		deps:    deps,
	}
	m.ui.Fold = foldState
	m.ui.Screen.ExpandDescriptions = expandDescriptions

	if startMode == modes.ModeProjectPicker {
		ordered := orderProjects(projects, stateManager.GetProjectOrder())
		selected := 0
		for i, p := range ordered {
			if p.ID == project.ID {
				selected = i
				break
			}
		}
		m.ui.Picker = ProjectPickerState{
			projects:     ordered,
			selected:     selected,
			scrollOffset: 0,
		}
		m.ui.Picker.ensureVisible(m.pickerVisibleCount())
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
