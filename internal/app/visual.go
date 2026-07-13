package app

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/atotto/clipboard"

	"phasionary/internal/app/selection"
	"phasionary/internal/domain"
)

// visualTaskLevel reports whether kind is a row that belongs to the unified
// task-level visual domain — a real task, a separator divider, or an expanded
// description row. Task-level selections cover all three, so any of them
// caught in the range (or chosen as an endpoint) highlights and travels with
// the block.
func visualTaskLevel(kind selection.FocusKind) bool {
	return kind == selection.FocusTask || kind == selection.FocusSeparator || kind == selection.FocusDescription
}

// matchesVisualDomain reports whether a row of kind k belongs to the same
// visual-selection domain as an anchor whose normalized kind is visualKind
// (either FocusCategory or FocusTask). Category anchors select only category
// rows; task-level anchors select the whole task+separator+description domain.
func matchesVisualDomain(k, visualKind selection.FocusKind) bool {
	if visualKind == selection.FocusCategory {
		return k == selection.FocusCategory
	}
	return visualTaskLevel(k)
}

// visualRoundUpPositions maps each description row in a visual selection to
// its parent task and drops the duplicate that appears when a task and its
// description are both in range — structural operations (delete, cut, move,
// tags) act on whole tasks, so the pair counts once. A description whose
// parent sits above the range still pulls that task in ("round up").
// Separators pass through untouched.
func visualRoundUpPositions(positions []selection.Position) []selection.Position {
	out := make([]selection.Position, 0, len(positions))
	seen := make(map[[2]int]struct{}, len(positions))
	for _, p := range positions {
		if p.Kind == selection.FocusDescription {
			p.Kind = selection.FocusTask
		}
		key := [2]int{p.CategoryIndex, p.TaskIndex}
		if _, dup := seen[key]; dup {
			continue
		}
		seen[key] = struct{}{}
		out = append(out, p)
	}
	return out
}

// countRealTasks returns how many of the positions are real tasks, ignoring any
// separators swept into a task-level range — so status counts report the number
// of tasks the user acted on rather than the raw row span.
func countRealTasks(positions []selection.Position) int {
	n := 0
	for _, p := range positions {
		if p.Kind == selection.FocusTask {
			n++
		}
	}
	return n
}

// rowCountLabel phrases a mixed task/separator batch for a status message. It
// counts real tasks, only falling back to a separator count when the batch was
// dividers alone, so a separator caught in a range never inflates "N task(s)".
func rowCountLabel(realTasks, separators int) string {
	if realTasks > 0 || separators == 0 {
		return fmt.Sprintf("%d task(s)", realTasks)
	}
	return fmt.Sprintf("%d separator(s)", separators)
}

func (m *model) enterVisualMode() {
	if !m.ui.Modes.IsNormal() {
		return
	}
	pos, ok := m.selectedPosition()
	if !ok || pos.Kind == selection.FocusProject {
		m.ui.Screen.StatusMsg = "Visual mode doesn't work on the project row"
		return
	}
	if !m.ui.Modes.ToVisual() {
		return
	}
	cat := m.project.Categories[pos.CategoryIndex]
	// Normalize a separator anchor to the task-level kind so the whole selection
	// (tasks and separators alike) shares one navigable domain.
	kind := selection.FocusCategory
	if visualTaskLevel(pos.Kind) {
		kind = selection.FocusTask
	}
	anchor := VisualState{
		Active:           true,
		Kind:             kind,
		AnchorCategoryID: cat.ID,
	}
	if visualTaskLevel(pos.Kind) {
		anchor.AnchorTaskID = cat.Tasks[pos.TaskIndex].ID
		anchor.AnchorOnDescription = pos.Kind == selection.FocusDescription
	}
	m.ui.Visual = anchor
}

func (m *model) exitVisualMode() {
	m.ui.Visual = VisualState{}
	m.ui.Modes.ToNormal()
}

// visualSwap exchanges the anchor and cursor positions, leaving the covered
// range unchanged. Useful for extending the selection from the other end.
func (m *model) visualSwap() {
	if !m.ui.Visual.Active {
		return
	}
	anchorIdx := m.resolveVisualAnchor()
	if anchorIdx < 0 {
		return
	}
	cursorIdx := m.ui.Selection.Selected()
	if anchorIdx == cursorIdx {
		return
	}
	positions := m.ui.Selection.Positions()
	if cursorIdx < 0 || cursorIdx >= len(positions) {
		return
	}
	curPos := positions[cursorIdx]
	if curPos.CategoryIndex < 0 || curPos.CategoryIndex >= len(m.project.Categories) {
		return
	}
	cat := m.project.Categories[curPos.CategoryIndex]
	m.ui.Visual.AnchorCategoryID = cat.ID
	m.ui.Visual.AnchorTaskID = ""
	m.ui.Visual.AnchorOnDescription = false
	if visualTaskLevel(curPos.Kind) {
		if curPos.TaskIndex < 0 || curPos.TaskIndex >= len(cat.Tasks) {
			return
		}
		m.ui.Visual.AnchorTaskID = cat.Tasks[curPos.TaskIndex].ID
		m.ui.Visual.AnchorOnDescription = curPos.Kind == selection.FocusDescription
	}
	m.ui.Selection.MoveTo(anchorIdx)
	m.ensureVisible()
}

func (m *model) visualMoveCursor(delta int) {
	if !m.ui.Visual.Active {
		return
	}
	positions := m.ui.Selection.Positions()
	if len(positions) == 0 {
		return
	}
	cur := m.ui.Selection.Selected()
	step := 1
	if delta < 0 {
		step = -1
	}
	for i := 0; i < abs(delta); i++ {
		next := nextVisualPosition(positions, cur, step, m.ui.Visual.Kind)
		if next == cur {
			break
		}
		cur = next
	}
	m.ui.Selection.MoveTo(cur)
	m.ensureVisible()
}

func nextVisualPosition(positions []selection.Position, from, step int, kind selection.FocusKind) int {
	i := from + step
	for i >= 0 && i < len(positions) {
		if matchesVisualDomain(positions[i].Kind, kind) {
			return i
		}
		i += step
	}
	return from
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (m model) visualRange() (lo, hi int, active bool) {
	if !m.ui.Visual.Active {
		return 0, 0, false
	}
	a := m.resolveVisualAnchor()
	if a < 0 {
		return 0, 0, false
	}
	b := m.ui.Selection.Selected()
	if a > b {
		a, b = b, a
	}
	return a, b, true
}

// resolveVisualAnchor finds the current position index of the visual anchor
// by matching the stable ID pair. Returns -1 if the anchor item no longer
// exists in the position list (e.g., it was deleted or filtered out).
func (m model) resolveVisualAnchor() int {
	positions := m.ui.Selection.Positions()
	for i, p := range positions {
		if !matchesVisualDomain(p.Kind, m.ui.Visual.Kind) {
			continue
		}
		if p.CategoryIndex < 0 || p.CategoryIndex >= len(m.project.Categories) {
			continue
		}
		cat := m.project.Categories[p.CategoryIndex]
		if cat.ID != m.ui.Visual.AnchorCategoryID {
			continue
		}
		if m.ui.Visual.Kind == selection.FocusCategory {
			return i
		}
		// Task-level: match the anchor by its stable task/separator ID. A task
		// row and its description row share that ID, so the description flag
		// disambiguates which of the two the anchor sits on.
		if (p.Kind == selection.FocusDescription) != m.ui.Visual.AnchorOnDescription {
			continue
		}
		if p.TaskIndex < 0 || p.TaskIndex >= len(cat.Tasks) {
			continue
		}
		if cat.Tasks[p.TaskIndex].ID == m.ui.Visual.AnchorTaskID {
			return i
		}
	}
	return -1
}

func (m model) visualSelectedPositions() []selection.Position {
	lo, hi, active := m.visualRange()
	if !active {
		return nil
	}
	positions := m.ui.Selection.Positions()
	var result []selection.Position
	for i := lo; i <= hi && i < len(positions); i++ {
		if matchesVisualDomain(positions[i].Kind, m.ui.Visual.Kind) {
			result = append(result, positions[i])
		}
	}
	return result
}

func (m model) isInVisualRange(index int) bool {
	lo, hi, active := m.visualRange()
	if !active {
		return false
	}
	if index < lo || index > hi {
		return false
	}
	positions := m.ui.Selection.Positions()
	if index < 0 || index >= len(positions) {
		return false
	}
	return matchesVisualDomain(positions[index].Kind, m.ui.Visual.Kind)
}

func (m *model) visualCopyTitles() tea.Cmd {
	selPositions := m.visualSelectedPositions()
	if len(selPositions) == 0 {
		m.exitVisualMode()
		return nil
	}
	text := m.visualTitleText(selPositions)
	m.stashVisualClipboard(selPositions, false)
	m.exitVisualMode()
	return func() tea.Msg {
		return clipboardResultMsg{err: clipboard.WriteAll(text)}
	}
}

// visualTitleText joins the plain titles of the selection. The copied text
// mirrors what is visibly selected: a description row in range contributes its
// body after the title (or alone, if only the description is selected).
// Separators carry no task content, so they are omitted from the copied text
// (the internal clipboard still keeps them for a faithful paste).
func (m model) visualTitleText(selPositions []selection.Position) string {
	titles := make([]string, 0, len(selPositions))
	for _, pos := range selPositions {
		switch pos.Kind {
		case selection.FocusTask:
			titles = append(titles, m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex].Title)
		case selection.FocusDescription:
			titles = append(titles, m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex].Description)
		case selection.FocusCategory:
			titles = append(titles, m.project.Categories[pos.CategoryIndex].Name)
		}
	}
	return strings.Join(titles, "\n")
}

func (m *model) visualCopyChecklist() tea.Cmd {
	selPositions := m.visualSelectedPositions()
	if len(selPositions) == 0 {
		m.exitVisualMode()
		return nil
	}
	text := m.visualChecklistText(selPositions)
	m.stashVisualClipboard(selPositions, false)
	m.exitVisualMode()
	return func() tea.Msg {
		return clipboardResultMsg{err: clipboard.WriteAll(text)}
	}
}

// visualChecklistText renders the selection as a markdown checklist. As with
// visualTitleText, separators are omitted from the copied text, and a selected
// description row nests its body under the parent's checklist item.
func (m model) visualChecklistText(selPositions []selection.Position) string {
	lines := make([]string, 0, len(selPositions))
	for _, pos := range selPositions {
		switch pos.Kind {
		case selection.FocusTask:
			task := m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex]
			mark := " "
			if task.Status == domain.StatusCompleted {
				mark = "x"
			}
			lines = append(lines, fmt.Sprintf("- [%s] %s", mark, task.Title))
		case selection.FocusDescription:
			desc := m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex].Description
			for _, l := range strings.Split(desc, "\n") {
				lines = append(lines, "  "+l)
			}
		case selection.FocusCategory:
			lines = append(lines, "## "+m.project.Categories[pos.CategoryIndex].Name)
		}
	}
	return strings.Join(lines, "\n")
}

func (m *model) stashVisualClipboard(selPositions []selection.Position, isCut bool) {
	if len(selPositions) == 0 {
		return
	}
	m.ui.TagCopiedLast = false
	switch m.ui.Visual.Kind {
	case selection.FocusCategory:
		cats := make([]domain.Category, 0, len(selPositions))
		ids := make([]string, 0, len(selPositions))
		for _, pos := range selPositions {
			c := m.project.Categories[pos.CategoryIndex]
			cats = append(cats, c)
			ids = append(ids, c.ID)
		}
		m.ui.Clipboard = ClipboardState{Categories: cats, CategoryIDs: ids, IsCut: isCut}
	default:
		// Task-level: stash the tasks and any interleaved separators so a paste
		// reproduces the block, dividers and all. Description rows round up to
		// their parent task (deduped) — a task is stashed once even when both
		// its row and its description are in range.
		selPositions = visualRoundUpPositions(selPositions)
		tasks := make([]domain.Task, 0, len(selPositions))
		ids := make([]string, 0, len(selPositions))
		for _, pos := range selPositions {
			t := m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex]
			tasks = append(tasks, t)
			ids = append(ids, t.ID)
		}
		m.ui.Clipboard = ClipboardState{Tasks: tasks, TaskIDs: ids, IsCut: isCut}
	}
}

func (m *model) visualDelete() {
	selPositions := m.visualSelectedPositions()
	if len(selPositions) == 0 {
		m.exitVisualMode()
		return
	}
	state := ConfirmDeleteState{Kind: ConfirmDeleteVisualRange}
	switch m.ui.Visual.Kind {
	case selection.FocusTask:
		for _, pos := range visualRoundUpPositions(selPositions) {
			state.TaskIDs = append(state.TaskIDs, m.project.Categories[pos.CategoryIndex].Tasks[pos.TaskIndex].ID)
		}
	case selection.FocusCategory:
		for _, pos := range selPositions {
			state.CategoryIDs = append(state.CategoryIDs, m.project.Categories[pos.CategoryIndex].ID)
		}
	}
	m.exitVisualMode()
	m.ui.ConfirmDelete = state
	m.ui.Modes.ToConfirmDelete()
}

func (m *model) confirmDeleteVisualRange() {
	taskIDs := m.ui.ConfirmDelete.TaskIDs
	catIDs := m.ui.ConfirmDelete.CategoryIDs
	m.ui.ConfirmDelete.reset()
	m.ui.Modes.ToNormal()

	m.recordHistory()
	// Track real tasks and separators separately so the message reflects tasks,
	// while `removed` still gates whether the edit is worth keeping in history.
	deletedTasks, deletedSeps, removed := 0, 0, 0
	for _, id := range taskIDs {
		isSep := false
		if t := m.taskByID(id); t != nil {
			isSep = t.IsSeparator()
		}
		if m.removeTaskByIDIfExists(id) {
			removed++
			if isSep {
				deletedSeps++
			} else {
				deletedTasks++
			}
		}
	}
	deletedCats := 0
	if len(catIDs) > 0 {
		idSet := make(map[string]struct{}, len(catIDs))
		for _, id := range catIDs {
			idSet[id] = struct{}{}
		}
		kept := m.project.Categories[:0]
		for _, c := range m.project.Categories {
			if _, drop := idSet[c.ID]; drop {
				deletedCats++
				removed++
				continue
			}
			kept = append(kept, c)
		}
		m.project.Categories = kept
		m.project.UpdatedAt = domain.NowTimestamp()
	}

	if removed == 0 {
		m.discardLastHistory()
		return
	}
	m.rebuildAndClamp()
	m.storeTaskUpdate()
	if len(catIDs) > 0 {
		m.ui.Screen.StatusMsg = fmt.Sprintf("Deleted %d categor%s", deletedCats, plural(deletedCats, "y", "ies"))
	} else {
		m.ui.Screen.StatusMsg = "Deleted " + rowCountLabel(deletedTasks, deletedSeps)
	}
}

func (m *model) removeTaskByIDIfExists(id string) bool {
	for catIdx := range m.project.Categories {
		for taskIdx, task := range m.project.Categories[catIdx].Tasks {
			if task.ID == id {
				_ = m.project.Categories[catIdx].RemoveTask(taskIdx)
				return true
			}
		}
	}
	return false
}

// visualMoveDown shifts the visual selection block one position down,
// preserving the block's contents and the anchor/cursor orientation.
// Categories shift across the whole project; tasks shift within their
// category, and at a category boundary the whole block crosses into the
// adjacent category (top when moving down, bottom when moving up), mirroring
// the single-task moveTaskDown/moveTaskUp behavior. No-op if the selection has
// no room to move, the block already spans multiple categories (tasks only),
// or it is not source-contiguous in task indices (e.g. a filter excludes
// interleaving tasks).
func (m *model) visualMoveDown() {
	m.visualShift(+1)
}

func (m *model) visualMoveUp() {
	m.visualShift(-1)
}

func (m *model) visualShift(dir int) {
	selPositions := m.visualSelectedPositions()
	if len(selPositions) == 0 {
		return
	}
	switch m.ui.Visual.Kind {
	case selection.FocusCategory:
		m.visualShiftCategories(selPositions, dir)
	case selection.FocusTask:
		m.visualShiftTasks(visualRoundUpPositions(selPositions), dir)
	}
}

func (m *model) visualShiftCategories(selPositions []selection.Position, dir int) {
	lo := selPositions[0].CategoryIndex
	hi := selPositions[len(selPositions)-1].CategoryIndex
	var neighbor, dst int
	if dir > 0 {
		if hi >= len(m.project.Categories)-1 {
			return
		}
		neighbor = hi + 1
		dst = lo
	} else {
		if lo <= 0 {
			return
		}
		neighbor = lo - 1
		dst = hi
	}
	m.recordHistory()
	moved := m.project.Categories[neighbor]
	_ = m.project.RemoveCategory(neighbor)
	m.project.InsertCategory(dst, moved)
	m.rebuildPositions()
	m.ui.Selection.MoveBy(dir)
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) visualShiftTasks(selPositions []selection.Position, dir int) {
	catIdx := selPositions[0].CategoryIndex
	for _, p := range selPositions[1:] {
		if p.CategoryIndex != catIdx {
			return
		}
	}
	taskLo := selPositions[0].TaskIndex
	taskHi := selPositions[len(selPositions)-1].TaskIndex
	if taskHi-taskLo+1 != len(selPositions) {
		return
	}
	cat := &m.project.Categories[catIdx]
	var neighbor, dst int
	if dir > 0 {
		if taskHi >= len(cat.Tasks)-1 {
			m.visualShiftTasksAcross(catIdx, taskLo, taskHi, dir)
			return
		}
		neighbor = taskHi + 1
		dst = taskLo
	} else {
		if taskLo <= 0 {
			m.visualShiftTasksAcross(catIdx, taskLo, taskHi, dir)
			return
		}
		neighbor = taskLo - 1
		dst = taskHi
	}
	cursorID, cursorOnDesc := m.visualCursorRowID()
	m.recordHistory()
	moved := cat.Tasks[neighbor]
	_ = cat.RemoveTask(neighbor)
	cat.InsertTask(dst, moved)
	m.project.UpdatedAt = domain.NowTimestamp()
	m.rebuildPositions()
	// The hopped neighbor may span one row or two (task + expanded
	// description), so the cursor re-attaches by ID instead of shifting by a
	// fixed row count.
	m.selectTaskLevelRow(cursorID, cursorOnDesc)
	m.ensureVisible()
	m.storeTaskUpdate()
}

// visualCursorRowID captures the cursor's task ID and whether the cursor sits
// on the task's description row, so the cursor can re-attach to the same row
// after a structural change re-indexes the position list.
func (m *model) visualCursorRowID() (string, bool) {
	cur, ok := m.ui.Selection.SelectedPosition()
	if !ok || !visualTaskLevel(cur.Kind) {
		return "", false
	}
	if cur.CategoryIndex < 0 || cur.CategoryIndex >= len(m.project.Categories) {
		return "", false
	}
	cat := m.project.Categories[cur.CategoryIndex]
	if cur.TaskIndex < 0 || cur.TaskIndex >= len(cat.Tasks) {
		return "", false
	}
	return cat.Tasks[cur.TaskIndex].ID, cur.Kind == selection.FocusDescription
}

// selectTaskLevelRow moves the selection to the task-level row with the given
// task ID — the description row when onDescription is set, the task row
// otherwise. No-op when id is empty or the row no longer exists.
func (m *model) selectTaskLevelRow(id string, onDescription bool) {
	if id == "" {
		return
	}
	m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
		if !visualTaskLevel(p.Kind) || (p.Kind == selection.FocusDescription) != onDescription {
			return false
		}
		return m.project.Categories[p.CategoryIndex].Tasks[p.TaskIndex].ID == id
	})
}

// visualShiftTasksAcross moves the contiguous task block [taskLo, taskHi] out
// of category catIdx and into the adjacent category in direction dir: moving
// down drops the block at the top of the next category, moving up at the
// bottom of the previous one. The cursor follows its task by ID, and the
// ID-based anchor re-resolves so the range still covers the moved block. No-op
// when there is no adjacent category.
func (m *model) visualShiftTasksAcross(catIdx, taskLo, taskHi, dir int) {
	var dstCatIdx int
	if dir > 0 {
		dstCatIdx = catIdx + 1
		if dstCatIdx >= len(m.project.Categories) {
			return
		}
	} else {
		dstCatIdx = catIdx - 1
		if dstCatIdx < 0 {
			return
		}
	}

	cursorID, cursorOnDesc := m.visualCursorRowID()

	m.recordHistory()
	src := &m.project.Categories[catIdx]
	moved := make([]domain.Task, taskHi-taskLo+1)
	copy(moved, src.Tasks[taskLo:taskHi+1])
	// Remove from the bottom up so earlier indices stay valid.
	for i := taskHi; i >= taskLo; i-- {
		_ = src.RemoveTask(i)
	}

	dst := &m.project.Categories[dstCatIdx]
	insertAt := 0
	if dir < 0 {
		insertAt = len(dst.Tasks)
	}
	for i, t := range moved {
		dst.InsertTask(insertAt+i, t)
	}

	m.project.UpdatedAt = domain.NowTimestamp()
	// The anchor task crossed with the block, so its recorded category ID must
	// follow or resolveVisualAnchor can no longer find it.
	m.ui.Visual.AnchorCategoryID = m.project.Categories[dstCatIdx].ID
	m.rebuildPositions()
	m.selectTaskLevelRow(cursorID, cursorOnDesc)
	m.ensureVisible()
	m.storeTaskUpdate()
}

func (m *model) visualCut() {
	selPositions := visualRoundUpPositions(m.visualSelectedPositions())
	if len(selPositions) == 0 {
		m.exitVisualMode()
		return
	}
	m.stashVisualClipboard(selPositions, true)
	if m.ui.Visual.Kind == selection.FocusCategory {
		n := len(selPositions)
		m.ui.Screen.StatusMsg = fmt.Sprintf("Marked %d categor%s for cut", n, plural(n, "y", "ies"))
	} else {
		tasks := countRealTasks(selPositions)
		m.ui.Screen.StatusMsg = "Marked " + rowCountLabel(tasks, len(selPositions)-tasks) + " for cut"
	}
	m.exitVisualMode()
}

func plural(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

func halfPageStep(availHeight int) int {
	step := availHeight / 2
	if step < 1 {
		return 1
	}
	return step
}

func (m *model) pasteMulti() bool {
	if len(m.ui.Clipboard.Tasks) == 0 && len(m.ui.Clipboard.Categories) == 0 {
		return false
	}
	if len(m.ui.Clipboard.Tasks) > 0 {
		m.pasteMultiTasks()
		return true
	}
	m.pasteMultiCategories()
	return true
}

func (m *model) pasteMultiTasks() {
	pos, ok := m.selectedPosition()
	if !ok || len(m.project.Categories) == 0 {
		m.ui.Screen.StatusMsg = "No category to paste into"
		return
	}
	var dstCatIndex, dstInsertIndex int
	switch pos.Kind {
	case selection.FocusProject:
		dstCatIndex = 0
		dstInsertIndex = 0
	case selection.FocusCategory:
		dstCatIndex = pos.CategoryIndex
		dstInsertIndex = 0
	case selection.FocusTask, selection.FocusDescription, selection.FocusSeparator:
		dstCatIndex = pos.CategoryIndex
		dstInsertIndex = pos.TaskIndex + 1
	}

	// Pre-allocate every new ID before mutating the project so a mid-loop
	// NewID() failure cannot leave a half-pasted/half-cut project (sources
	// removed, only some destinations inserted) — pasteMulti is otherwise
	// non-atomic across the source removal and the insertion loop.
	newIDs := make([]string, len(m.ui.Clipboard.Tasks))
	for i := range newIDs {
		id, err := domain.NewID()
		if err != nil {
			m.ui.Screen.StatusMsg = "Failed to create task ID"
			return
		}
		newIDs[i] = id
	}

	m.recordHistory()

	if m.ui.Clipboard.IsCut {
		if dstCatIndex >= 0 && dstCatIndex < len(m.project.Categories) {
			cutIDs := make(map[string]struct{}, len(m.ui.Clipboard.TaskIDs))
			for _, id := range m.ui.Clipboard.TaskIDs {
				cutIDs[id] = struct{}{}
			}
			before := 0
			for idx, t := range m.project.Categories[dstCatIndex].Tasks {
				if idx >= dstInsertIndex {
					break
				}
				if _, ok := cutIDs[t.ID]; ok {
					before++
				}
			}
			dstInsertIndex -= before
		}
		for _, id := range m.ui.Clipboard.TaskIDs {
			m.removeTaskByID(id)
		}
		if dstInsertIndex < 0 {
			dstInsertIndex = 0
		}
		if dstCatIndex < len(m.project.Categories) && dstInsertIndex > len(m.project.Categories[dstCatIndex].Tasks) {
			dstInsertIndex = len(m.project.Categories[dstCatIndex].Tasks)
		}
	}

	var firstNewID string
	for i, src := range m.ui.Clipboard.Tasks {
		newTask := domain.Task{
			ID:              newIDs[i],
			Title:           src.Title,
			Status:          src.Status,
			CreatedAt:       src.CreatedAt,
			UpdatedAt:       domain.NowTimestamp(),
			Priority:        src.Priority,
			CompletionDate:  src.CompletionDate,
			EstimateMinutes: src.EstimateMinutes,
			Description:     src.Description,
			Kind:            src.Kind,
			TagColor:        src.TagColor,
			TagLabel:        src.TagLabel,
		}
		m.project.Categories[dstCatIndex].InsertTask(dstInsertIndex+i, newTask)
		if i == 0 {
			firstNewID = newIDs[i]
		}
	}

	pastedTasks := 0
	for _, t := range m.ui.Clipboard.Tasks {
		if !t.IsSeparator() {
			pastedTasks++
		}
	}
	pastedSeps := len(m.ui.Clipboard.Tasks) - pastedTasks
	wasCut := m.ui.Clipboard.IsCut
	m.ui.Clipboard = ClipboardState{}
	m.rebuildPositions()
	if firstNewID != "" {
		m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
			if !visualTaskLevel(p.Kind) {
				return false
			}
			return m.project.Categories[p.CategoryIndex].Tasks[p.TaskIndex].ID == firstNewID
		})
	}
	m.ensureVisible()
	m.storeTaskUpdate()
	verb := "Pasted"
	if wasCut {
		verb = "Moved"
	}
	m.ui.Screen.StatusMsg = verb + " " + rowCountLabel(pastedTasks, pastedSeps)
}

func (m *model) pasteMultiCategories() {
	pos, ok := m.selectedPosition()
	if !ok {
		return
	}
	dstIndex := 0
	switch pos.Kind {
	case selection.FocusProject:
		dstIndex = 0
	case selection.FocusCategory:
		dstIndex = pos.CategoryIndex
	case selection.FocusTask, selection.FocusDescription, selection.FocusSeparator:
		dstIndex = pos.CategoryIndex
	}

	// Pre-allocate every new ID (category + tasks for the yank path, only
	// missing-category IDs for the cut path) before mutating the project, so
	// a mid-loop NewID failure cannot leave a half-pasted state.
	type catIDs struct {
		cat   string
		tasks []string
	}
	preIDs := make([]catIDs, len(m.ui.Clipboard.Categories))
	cutIDs := make(map[string]struct{}, len(m.ui.Clipboard.CategoryIDs))
	for _, id := range m.ui.Clipboard.CategoryIDs {
		cutIDs[id] = struct{}{}
	}
	for i, src := range m.ui.Clipboard.Categories {
		if m.ui.Clipboard.IsCut {
			if _, present := cutIDs[src.ID]; !present {
				id, err := domain.NewID()
				if err != nil {
					m.ui.Screen.StatusMsg = "Failed to create category ID"
					return
				}
				preIDs[i].cat = id
			}
		} else {
			id, err := domain.NewID()
			if err != nil {
				m.ui.Screen.StatusMsg = "Failed to create category ID"
				return
			}
			preIDs[i].cat = id
			preIDs[i].tasks = make([]string, len(src.Tasks))
			for j := range src.Tasks {
				tid, err := domain.NewID()
				if err != nil {
					m.ui.Screen.StatusMsg = "Failed to create task ID"
					return
				}
				preIDs[i].tasks[j] = tid
			}
		}
	}

	m.recordHistory()

	if m.ui.Clipboard.IsCut {
		// Anchor on the first surviving neighbor at or after the cursor; if
		// none, fall back to the last surviving neighbor before the cursor.
		anchorID := ""
		for i := dstIndex; i < len(m.project.Categories); i++ {
			if _, drop := cutIDs[m.project.Categories[i].ID]; !drop {
				anchorID = m.project.Categories[i].ID
				break
			}
		}
		if anchorID == "" {
			for i := dstIndex - 1; i >= 0; i-- {
				if _, drop := cutIDs[m.project.Categories[i].ID]; !drop {
					anchorID = m.project.Categories[i].ID
					break
				}
			}
		}
		kept := m.project.Categories[:0]
		for _, c := range m.project.Categories {
			if _, removed := cutIDs[c.ID]; removed {
				continue
			}
			kept = append(kept, c)
		}
		m.project.Categories = kept
		dstIndex = 0
		if anchorID != "" {
			for i, c := range m.project.Categories {
				if c.ID == anchorID {
					dstIndex = i
					break
				}
			}
		}
	}

	if dstIndex > len(m.project.Categories) {
		dstIndex = len(m.project.Categories)
	}
	if dstIndex < 0 {
		dstIndex = 0
	}

	var firstNewID string
	for i, src := range m.ui.Clipboard.Categories {
		newCat := src
		newCat.UpdatedAt = domain.NowTimestamp()
		// Cut preserves IDs (the source was already removed by the kept-filter
		// above, so there is no collision). Yank/copy mints fresh IDs for the
		// category and every task so the project never holds duplicate IDs.
		newTasks := make([]domain.Task, len(src.Tasks))
		copy(newTasks, src.Tasks)
		if preIDs[i].cat != "" {
			newCat.ID = preIDs[i].cat
		}
		if !m.ui.Clipboard.IsCut {
			for j := range newTasks {
				newTasks[j].ID = preIDs[i].tasks[j]
			}
		}
		newCat.Tasks = newTasks
		m.project.InsertCategory(dstIndex+i, newCat)
		if i == 0 {
			firstNewID = newCat.ID
		}
	}

	count := len(m.ui.Clipboard.Categories)
	wasCut := m.ui.Clipboard.IsCut
	m.ui.Clipboard = ClipboardState{}
	m.rebuildPositions()
	if firstNewID != "" {
		m.ui.Selection.SelectByPredicate(func(p selection.Position) bool {
			return p.Kind == selection.FocusCategory && m.project.Categories[p.CategoryIndex].ID == firstNewID
		})
	}
	m.ensureVisible()
	m.storeTaskUpdate()
	verb := "Pasted"
	if wasCut {
		verb = "Moved"
	}
	m.ui.Screen.StatusMsg = fmt.Sprintf("%s %d categor%s", verb, count, plural(count, "y", "ies"))
}

func (m model) handleVisualKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc", "v", "q":
		m.exitVisualMode()
		return m, nil
	case "j", "down":
		m.visualMoveCursor(1)
		return m, nil
	case "k", "up":
		m.visualMoveCursor(-1)
		return m, nil
	case "J":
		m.visualMoveDown()
		return m, nil
	case "K":
		m.visualMoveUp()
		return m, nil
	case "ctrl+d":
		m.visualMoveCursor(halfPageStep(m.availableHeight()))
		return m, nil
	case "ctrl+u":
		m.visualMoveCursor(-halfPageStep(m.availableHeight()))
		return m, nil
	case "ctrl+f":
		m.visualMoveCursor(m.availableHeight())
		return m, nil
	case "ctrl+b":
		m.visualMoveCursor(-m.availableHeight())
		return m, nil
	case "G":
		m.visualMoveCursor(len(m.ui.Selection.Positions()))
		return m, nil
	case "g":
		m.visualMoveCursor(-len(m.ui.Selection.Positions()))
		return m, nil
	case "o", "O":
		m.visualSwap()
		return m, nil
	case "y":
		return m, m.visualCopyTitles()
	case "Y":
		return m, m.visualCopyChecklist()
	case "x":
		m.visualCut()
		return m, nil
	case "d":
		m.visualDelete()
		return m, nil
	case "t":
		m.visualCycleTag()
		return m, nil
	case "T":
		return m, m.visualStartTagEdit()
	case "p":
		m.visualPasteTag()
		return m, nil
	}
	return m, nil
}
