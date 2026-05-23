package app

import (
	"fmt"
	"strings"

	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

func (m model) filterView() string {
	statusLabels := map[string]string{
		domain.StatusTodo:       "Todo",
		domain.StatusInProgress: "In Progress",
		domain.StatusCompleted:  "Completed",
		domain.StatusCancelled:  "Cancelled",
	}
	lines := []string{ui.DialogTitleStyle.Render("Filter by Status:"), ""}
	for i, status := range filterStatuses {
		prefix := "  "
		if i == m.ui.Filter.Selected() {
			prefix = "> "
		}
		checkbox := "[ ]"
		if m.ui.Filter.IsEnabled(status) {
			checkbox = "[x]"
		}
		line := fmt.Sprintf("%s%s %s", prefix, checkbox, statusLabels[status])
		if i == m.ui.Filter.Selected() {
			line = ui.SelectedStyle.Render(line)
		}
		lines = append(lines, line)
	}
	lines = append(lines, "", ui.DialogHintStyle.Render("j/k navigate | space toggle | q/esc/f close"))
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
