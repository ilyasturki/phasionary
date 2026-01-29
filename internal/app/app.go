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
	scrollOffset  int    // position index at top of visible area
	pendingKey    rune   // for multi-key sequences like gg and zz
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
		// Ignore mouse events when in overlays or editing mode
		if m.editing || m.showHelp || m.confirmDelete {
			break
		}
		// Only handle left mouse button press
		if msg.Button != tea.MouseButtonLeft {
			break
		}
		// Ignore wheel events
		if msg.Action != tea.MouseActionPress {
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
			m.moveTaskDown()
			m.pendingKey = 0
		case "K":
			m.moveTaskUp()
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
	availHeight := m.availableHeight()
	usedHeight := 0
	cursor := 0
	hasMoreAbove := m.scrollOffset > 0
	hasMoreBelow := false

	// Reserve space for scroll indicators when needed
	if hasMoreAbove {
		availHeight-- // Space for "more above"
	}
	// Reserve space for potential "more below" indicator
	availHeight--
	if availHeight < 1 {
		availHeight = 1
	}

	// Helper to render an element if visible.
	// Each element occupies elementHeight visual rows. A separator \n is added
	// between elements but NOT counted toward usedHeight.
	renderIfVisible := func(renderFn func() string, elementHeight int) bool {
		if cursor < m.scrollOffset {
			cursor++
			return true // Skip, continue
		}
		if usedHeight+elementHeight > availHeight {
			hasMoreBelow = true
			return false // Stop rendering
		}
		if usedHeight > 0 {
			bodyBuilder.WriteString("\n") // Separator (not counted)
		}
		bodyBuilder.WriteString(renderFn())
		usedHeight += elementHeight
		cursor++
		return true
	}

	// Helper to add blank visual lines between elements.
	// Each blank line is a real visual row counted toward usedHeight.
	addBlankLines := func(count int) {
		if cursor <= m.scrollOffset || usedHeight == 0 {
			return // Don't add blank lines before we start rendering
		}
		for i := 0; i < count; i++ {
			if usedHeight+1 > availHeight {
				return
			}
			bodyBuilder.WriteString("\n") // Each \n = 1 blank visual line
			usedHeight++
		}
	}

	// Project line (first focusable item)
	isProjectSelected := cursor == m.selected
	projectHeight := m.countProjectLines()
	if !renderIfVisible(func() string {
		if m.editing && isProjectSelected {
			return m.renderEditProjectLine()
		}
		return renderProjectLine(m.project.Name, isProjectSelected)
	}, projectHeight) {
		goto footer
	}
	addBlankLines(2) // 2 blank lines after project

	for i, category := range m.categories {
		if i > 0 {
			addBlankLines(1) // 1 blank line between categories
		}

		// Category header
		isSelected := cursor == m.selected
		catHeight := m.countCategoryLines(category.Name)
		if !renderIfVisible(func() string {
			if m.editing && isSelected {
				return m.renderEditCategoryLine()
			}
			return renderCategoryLine(category.Name, isSelected, m.width)
		}, catHeight) {
			goto footer
		}

		if len(category.Tasks) == 0 {
			// "(no tasks)" placeholder - not a position, just visual
			if cursor > m.scrollOffset && usedHeight+1 <= availHeight {
				bodyBuilder.WriteString("\n")
				bodyBuilder.WriteString(ui.MutedStyle.Render("  (no tasks)"))
				usedHeight++
			}
			continue
		}

		addBlankLines(1) // 1 blank line after category header

		// Tasks (consecutive tasks have no blank lines between them)
		for _, task := range category.Tasks {
			isTaskSelected := cursor == m.selected
			taskHeight := m.countTaskLines(task)
			taskCopy := task // capture for closure
			if !renderIfVisible(func() string {
				if m.editing && isTaskSelected {
					return m.renderEditTaskLine(taskCopy)
				}
				return renderTaskLine(taskCopy, isTaskSelected, m.width)
			}, taskHeight) {
				goto footer
			}
		}
	}

	// Check if there's more content below
	if cursor < len(m.positions) {
		hasMoreBelow = true
	}

footer:
	body := strings.TrimRight(bodyBuilder.String(), "\n")

	// Add scroll indicators
	if hasMoreAbove {
		body = ui.MutedStyle.Render("  ↑ more above") + "\n" + body
	}
	if hasMoreBelow {
		body = body + "\n" + ui.MutedStyle.Render("  ↓ more below")
	}

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
	}, tea.WithAltScreen(), tea.WithMouseCellMotion())
	_, err = program.Run()
	return err
}

func buildViews(project domain.Project) ([]categoryView, []focusPosition) {
	categories := make([]categoryView, 0, len(project.Categories))
	for _, category := range project.Categories {
		tasks := append([]domain.Task(nil), category.Tasks...)
		categories = append(categories, categoryView{
			Name:  category.Name,
			Tasks: tasks,
		})
	}
	positions := rebuildPositions(categories)
	return categories, positions
}
