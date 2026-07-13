package components

import (
	"fmt"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

type TaskLineRenderer struct {
	width               int
	statusDisplay       string
	priorityColor       string
	focused             bool
	visualMode          bool
	isCursor            bool
	cut                 bool
	descriptionExpanded bool
	searchQuery         string
	searchMatch         lipgloss.Style
}

func NewTaskLineRenderer(width int, statusDisplay, priorityColor string, focused bool) *TaskLineRenderer {
	return &TaskLineRenderer{
		width:         width,
		statusDisplay: statusDisplay,
		priorityColor: priorityColor,
		focused:       focused,
	}
}

func (r *TaskLineRenderer) WithVisualMode(v bool) *TaskLineRenderer {
	r.visualMode = v
	return r
}

func (r *TaskLineRenderer) WithCursor(c bool) *TaskLineRenderer {
	r.isCursor = c
	return r
}

func (r *TaskLineRenderer) WithCut(c bool) *TaskLineRenderer {
	r.cut = c
	return r
}

func (r *TaskLineRenderer) WithDescriptionExpanded(v bool) *TaskLineRenderer {
	r.descriptionExpanded = v
	return r
}

// WithSearch enables search-match highlighting: occurrences of query within the
// title/description are rendered with match instead of the base text style.
func (r *TaskLineRenderer) WithSearch(query string, match lipgloss.Style) *TaskLineRenderer {
	r.searchQuery = query
	r.searchMatch = match
	return r
}

// highlightLine styles one already-wrapped plain line with base, emphasizing any
// active search match within it. With no active query it is just
// base.Render(line).
func (r *TaskLineRenderer) highlightLine(line string, base lipgloss.Style) string {
	if r.searchQuery == "" {
		return base.Render(line)
	}
	return ui.HighlightMatches(line, r.searchQuery, base, r.searchMatch)
}

func (r *TaskLineRenderer) maybeCut(s lipgloss.Style) lipgloss.Style {
	if r.cut {
		return ui.ApplyCut(s)
	}
	return s
}

func (r *TaskLineRenderer) selectedStyle() lipgloss.Style {
	if r.visualMode {
		if r.isCursor {
			return r.maybeCut(ui.GetVisualCursorStyle(r.focused))
		}
		return r.maybeCut(ui.GetVisualSelectedStyle(r.focused))
	}
	return r.maybeCut(ui.GetSelectedStyle(r.focused))
}

func (r *TaskLineRenderer) Render(task domain.Task, selected bool) string {
	prefix := "  "
	if selected {
		prefix = "> "
	}
	priorityIcon := ui.PriorityIcon(task.Priority)

	if selected {
		return r.padToWidth(r.renderSelected(task, prefix, priorityIcon))
	}
	return r.renderUnselected(task, prefix, priorityIcon)
}

// padToWidth extends each rendered line to the full row width using the
// active selection style, so a visual-mode highlight reads as a solid band
// rather than stopping at the end of the text. Only active when the row
// participates in the visual range (r.visualMode); normal-mode selection
// keeps its current "text-width" highlight.
func (r *TaskLineRenderer) padToWidth(rendered string) string {
	if r.width <= 0 || !r.visualMode {
		return rendered
	}
	style := r.selectedStyle()
	lines := strings.Split(rendered, "\n")
	for i, l := range lines {
		gap := r.width - ansi.StringWidth(l)
		if gap <= 0 {
			continue
		}
		lines[i] = l + style.Render(strings.Repeat(" ", gap))
	}
	return strings.Join(lines, "\n")
}

const descriptionBar = "▎ "

// RenderDescription renders a task description as a standalone, focusable block
// with a blockquote-style left bar. The bar sits at `indent` columns (aligned
// with the title on the task row) and text follows after it. The text mirrors
// the title's color treatment (priority color, faint when completed/cancelled)
// with italic on top so descriptions still read as a distinct element.
func (r *TaskLineRenderer) RenderDescription(task domain.Task, indent int, selected bool) string {
	description := task.Description
	if description == "" {
		return ""
	}
	indentStr := strings.Repeat(" ", indent)
	prefixWidth := indent + ansi.StringWidth(descriptionBar)
	available := 0
	if r.width > 0 {
		available = r.width - prefixWidth
		if available < 1 {
			available = 0
		}
	}

	textStyle := ui.TaskTitleStyle(task.Priority, task.Status, r.priorityColor).Italic(true)
	barStyle := ui.DescriptionBarStyle
	if r.cut {
		textStyle = ui.ApplyCut(textStyle)
		barStyle = ui.ApplyCut(barStyle)
	}

	// When selected, every visual line gets padded to the full row width so the
	// highlight rectangle is uniform — otherwise blank-line indents and short
	// paragraph tails produce a ragged selection band.
	shouldPad := selected && r.width > 0

	var out []string
	for _, paragraph := range strings.Split(description, "\n") {
		if paragraph == "" {
			out = append(out, r.styleDescriptionLine(indentStr, "", textStyle, barStyle, selected, shouldPad))
			continue
		}
		text := paragraph
		if available > 0 {
			text = ansi.Wrap(paragraph, available, "")
		}
		for _, l := range strings.Split(text, "\n") {
			out = append(out, r.styleDescriptionLine(indentStr, l, textStyle, barStyle, selected, shouldPad))
		}
	}
	return strings.Join(out, "\n")
}

func (r *TaskLineRenderer) styleDescriptionLine(indentStr, text string, textStyle, barStyle lipgloss.Style, selected, shouldPad bool) string {
	if selected {
		sel := r.selectedStyle()
		prefix := indentStr + descriptionBar
		// Render the bar and text separately so a search match keeps its own
		// styling inside the selection band; pad with the band style to width.
		rendered := sel.Render(prefix) + r.highlightLine(text, sel)
		if shouldPad {
			if gap := r.width - ansi.StringWidth(prefix) - ansi.StringWidth(text); gap > 0 {
				rendered += sel.Render(strings.Repeat(" ", gap))
			}
		}
		return rendered
	}
	return indentStr + barStyle.Render(descriptionBar) + r.highlightLine(text, textStyle)
}

func (r *TaskLineRenderer) renderUnselected(task domain.Task, prefix, priorityIcon string) string {
	status := r.formatStatus(task.Status, false)
	finished := task.Status == domain.StatusCompleted || task.Status == domain.StatusCancelled
	iconStyled := ""
	iconWidth := 0
	if priorityIcon != "" {
		iconStyle := ui.PriorityIconStyle(task.Priority, r.priorityColor)
		if finished {
			iconStyle = iconStyle.Faint(true)
		}
		iconStyle = r.maybeCut(iconStyle)
		iconStyled = iconStyle.Render(priorityIcon) + " "
		iconWidth = ansi.StringWidth(priorityIcon + " ")
	}
	tagStyled, tagWidth := r.tagSegment(task, finished, false, priorityIcon != "")
	leading := tagStyled + iconStyled
	leadingWidth := tagWidth + iconWidth
	descMarker := r.formatDescriptionBadge(task.Description, false)
	estimate := r.formatEstimateBadge(task.EstimateMinutes, false)
	cutBadge, cutBadgeText := r.formatCutBadge(false)
	suffix := cutBadge + descMarker + estimate
	suffixText := cutBadgeText + r.descriptionBadgeText(task.Description) + r.estimateBadgeText(task.EstimateMinutes)
	titleStyle := r.maybeCut(ui.TaskTitleStyle(task.Priority, task.Status, r.priorityColor))
	prefixPart := fmt.Sprintf("%s[%s] ", prefix, status)

	if r.width <= 0 {
		return prefixPart + leading + r.highlightLine(task.Title, titleStyle) + suffix
	}

	return r.wrapTaskContentWithSuffix(task.Title, prefixPart, leading, leadingWidth, titleStyle, suffix, suffixText)
}

// tagSegment renders the leading tag (a colored dot plus its optional label and
// a trailing space) and its display width, or ("", 0) when untagged. On an
// unselected row the tag is its own colored foreground; on a highlighted row it
// becomes a filled color block (reverse) so it shares the selection bar's look
// instead of clashing as a colored foreground island. faint mirrors the priority
// icon's completed/cancelled dimming (unselected rows only). priorityFollows
// reports whether a priority icon block sits between the tag and the title.
func (r *TaskLineRenderer) tagSegment(task domain.Task, finished, selected, priorityFollows bool) (string, int) {
	seg := ui.TagSegmentText(task.TagColor, task.TagLabel)
	if seg == "" {
		return "", 0
	}
	width := ansi.StringWidth(seg)
	if selected {
		style, _ := ui.TagBlockStyle(task.TagColor)
		// A labeled tag's segment ends with a separator space before the next
		// element. Keep that space in the row's selection band when the title
		// follows directly, so the colored block hugs the label instead of
		// trailing one cell past it. But when a priority icon block follows,
		// color the separator with the tag so the two colored blocks connect
		// rather than being split by a monochrome gap. A color-only tag has no
		// such trailing space — its single space is the dot-to-next separator —
		// so it always stays a whole block.
		if task.TagLabel != "" && !priorityFollows {
			core := strings.TrimSuffix(seg, " ")
			return r.maybeCut(style).Render(core) + r.selectedStyle().Render(" "), width
		}
		return r.maybeCut(style).Render(seg), width
	}
	style, _ := ui.TagDotStyle(task.TagColor, finished)
	return r.maybeCut(style).Render(seg), width
}

func (r *TaskLineRenderer) renderSelected(task domain.Task, prefix, priorityIcon string) string {
	statusText := r.statusLabel(task.Status)
	selectedStyle := r.selectedStyle()
	priorityStyle := selectedStyle
	statusStyle := selectedStyle

	iconStyled := ""
	iconWidth := 0
	if priorityIcon != "" {
		// Keep the priority color visible on the selection bar by rendering the
		// icon as a reverse color block, the same treatment the tag dot gets in
		// tagSegment; fall back to the plain selection style when icon color is
		// off so the icon still sits in the band.
		iconStyle, ok := ui.PriorityIconBlockStyle(task.Priority, r.priorityColor)
		if !ok {
			iconStyle = selectedStyle
		}
		iconStyled = r.maybeCut(iconStyle).Render(priorityIcon + " ")
		iconWidth = ansi.StringWidth(priorityIcon + " ")
	}
	finished := task.Status == domain.StatusCompleted || task.Status == domain.StatusCancelled
	tagStyled, tagWidth := r.tagSegment(task, finished, true, priorityIcon != "")
	leading := tagStyled + iconStyled
	leadingWidth := tagWidth + iconWidth

	descMarker := r.formatDescriptionBadge(task.Description, true)
	estimate := r.formatEstimateBadge(task.EstimateMinutes, true)
	cutBadge, cutBadgeText := r.formatCutBadge(true)
	suffix := cutBadge + descMarker + estimate
	suffixText := cutBadgeText + r.descriptionBadgeText(task.Description) + r.estimateBadgeText(task.EstimateMinutes)

	prefixPart := selectedStyle.Render(prefix+"[") +
		statusStyle.Render(statusText) +
		selectedStyle.Render("] ")

	if r.width <= 0 {
		return prefixPart + leading + r.highlightLine(task.Title, priorityStyle) + suffix
	}

	overhead := ansi.StringWidth(prefix + "[" + statusText + "] ")
	return r.wrapSelectedContentWithSuffix(task.Title, prefixPart, leading, leadingWidth, overhead, priorityStyle, suffix, suffixText)
}

func (r *TaskLineRenderer) wrapTaskContentWithSuffix(title, prefixPart, leading string, leadingWidth int, titleStyle lipgloss.Style, suffix, suffixText string) string {
	overhead := ansi.StringWidth(prefixPart)
	suffixWidth := ansi.StringWidth(suffixText)
	available := safeWidth(r.width, overhead+leadingWidth+suffixWidth)
	wrapped := ansi.Wrap(title, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)
	last := len(wrapLines) - 1

	var result []string
	for i, line := range wrapLines {
		styledLine := r.highlightLine(line, titleStyle)
		var rendered string
		if i == 0 {
			rendered = prefixPart + leading + styledLine
		} else {
			rendered = indent + styledLine
		}
		if i == last {
			rendered += suffix
		}
		result = append(result, rendered)
	}
	return strings.Join(result, "\n")
}

func (r *TaskLineRenderer) wrapSelectedContentWithSuffix(title, prefixPart, leading string, leadingWidth, overhead int, titleStyle lipgloss.Style, suffix, suffixText string) string {
	suffixWidth := ansi.StringWidth(suffixText)
	available := safeWidth(r.width, overhead+leadingWidth+suffixWidth)
	wrapped := ansi.Wrap(title, available, "")
	wrapLines := strings.Split(wrapped, "\n")
	indent := strings.Repeat(" ", overhead)
	selectedStyle := r.selectedStyle()
	last := len(wrapLines) - 1

	var result []string
	for i, line := range wrapLines {
		styledTitle := r.highlightLine(line, titleStyle)
		var rendered string
		if i == 0 {
			rendered = prefixPart + leading + styledTitle
		} else {
			styledIndent := selectedStyle.Render(indent)
			rendered = styledIndent + styledTitle
		}
		if i == last {
			rendered += suffix
		}
		result = append(result, rendered)
	}
	return strings.Join(result, "\n")
}

func (r *TaskLineRenderer) formatStatus(status string, selected bool) string {
	label := r.statusLabel(status)
	if selected {
		return label
	}
	return ui.StatusStyle(status).Render(label)
}

func (r *TaskLineRenderer) statusLabel(status string) string {
	if r.statusDisplay == "icons" {
		return statusIcon(status)
	}
	switch status {
	case domain.StatusInProgress:
		return " progress"
	case domain.StatusCompleted:
		return "completed"
	case domain.StatusCancelled:
		return "cancelled"
	default:
		return "  todo   "
	}
}

func statusIcon(status string) string {
	switch status {
	case domain.StatusInProgress:
		return "/"
	case domain.StatusCompleted:
		return "x"
	case domain.StatusCancelled:
		return "-"
	default:
		return " "
	}
}

func safeWidth(totalWidth, overhead int) int {
	available := totalWidth - overhead
	if available < 1 {
		return 1
	}
	return available
}

func (r *TaskLineRenderer) formatEstimateBadge(minutes int, selected bool) string {
	text := r.estimateBadgeText(minutes)
	if text == "" {
		return ""
	}
	if selected {
		return r.selectedStyle().Render(text)
	}
	return ui.MutedStyle.Render(text)
}

func (r *TaskLineRenderer) estimateBadgeText(minutes int) string {
	if minutes == 0 {
		return ""
	}
	return " ~" + formatEstimateShort(minutes)
}

func (r *TaskLineRenderer) formatDescriptionBadge(description string, selected bool) string {
	text := r.descriptionBadgeText(description)
	if text == "" {
		return ""
	}
	if selected {
		return r.selectedStyle().Render(text)
	}
	return ui.MutedStyle.Render(text)
}

func (r *TaskLineRenderer) descriptionBadgeText(description string) string {
	if description == "" || r.descriptionExpanded {
		return ""
	}
	return " ¶"
}

func (r *TaskLineRenderer) formatCutBadge(selected bool) (string, string) {
	if !r.cut {
		return "", ""
	}
	text := ui.CutMark
	if selected {
		return r.selectedStyle().Render(text), text
	}
	return ui.CutBadgeStyle.Render(text), text
}

func formatEstimateShort(minutes int) string {
	if minutes < 60 {
		return fmt.Sprintf("%dm", minutes)
	}
	hours := minutes / 60
	if hours < 8 {
		return fmt.Sprintf("%dh", hours)
	}
	days := hours / 8
	return fmt.Sprintf("%dd", days)
}
