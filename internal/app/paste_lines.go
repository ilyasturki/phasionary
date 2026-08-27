package app

import (
	"fmt"
	"regexp"
	"strings"

	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

// pastedTaskLine is one task parsed out of a multi-line paste: the title with
// any leading list marker stripped, and the status its checkbox implies.
type pastedTaskLine struct {
	title  string
	status string
}

// listMarkerPattern matches a leading bullet ("-", "*", "+", "•") or numbered
// ("1.", "1)") list marker, optionally followed by a "[ ]"/"[x]" checkbox.
var listMarkerPattern = regexp.MustCompile(`^\s*(?:[-*+•]|\d+[.)])\s+(?:\[([ xX])\]\s*)?`)

// parsePastedTaskLines splits pasted text into one task per non-empty line.
func parsePastedTaskLines(text string) []pastedTaskLine {
	text = strings.ReplaceAll(text, "\r", "\n")
	var out []pastedTaskLine
	for _, line := range strings.Split(text, "\n") {
		title := line
		status := domain.StatusTodo
		if m := listMarkerPattern.FindStringSubmatch(line); m != nil {
			title = line[len(m[0]):]
			if m[1] == "x" || m[1] == "X" {
				status = domain.StatusCompleted
			}
		}
		title = strings.TrimSpace(title)
		if title == "" {
			continue
		}
		out = append(out, pastedTaskLine{title: title, status: status})
	}
	return out
}

// pasteLinesWhileAdding turns a multi-line paste during the add-task flow into
// a batch add, one task per line, then closes the edit. Reports false —
// leaving the paste to the plain single-line input path — when not adding a
// task or when the paste holds fewer than two usable lines. The whole batch
// undoes in one step via the history entry startAddingTask already recorded.
func (m *model) pasteLinesWhileAdding(text string) bool {
	if !m.ui.Edit.isAdding || m.ui.Edit.itemType != selection.FocusTask {
		return false
	}
	lines := parsePastedTaskLines(text)
	if len(lines) < 2 {
		return false
	}
	pos, ok := m.selectedPosition()
	if !ok || pos.Kind != selection.FocusTask {
		return false
	}

	if typed := strings.TrimSpace(m.ui.Edit.input.Value()); typed != "" {
		lines[0].title = typed + " " + lines[0].title
	}

	// Mint every task up front so a mid-loop NewID failure cannot leave a
	// half-inserted batch.
	newTasks := make([]domain.Task, 0, len(lines)-1)
	for _, ln := range lines[1:] {
		task, err := domain.NewTask(ln.title)
		if err != nil {
			m.ui.Screen.StatusMsg = "Failed to create task ID"
			return true
		}
		_ = task.SetStatus(ln.status)
		newTasks = append(newTasks, task)
	}

	cat := &m.project.Categories[pos.CategoryIndex]
	pending := &cat.Tasks[pos.TaskIndex]
	pending.Title = lines[0].title
	_ = pending.SetStatus(lines[0].status)
	for i, task := range newTasks {
		cat.InsertTask(pos.TaskIndex+1+i, task)
	}

	m.ui.Modes.ToNormal()
	m.ui.Edit.reset()
	m.rebuildPositions()
	m.selectTaskByID(newTasks[len(newTasks)-1].ID)
	m.ensureVisible()
	m.storeTaskUpdate()
	m.ui.Screen.StatusMsg = fmt.Sprintf("Added %d tasks", len(lines))
	return true
}
