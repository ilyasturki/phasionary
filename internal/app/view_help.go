package app

import (
	"fmt"
	"strings"

	"phasionary/internal/ui"
)

func (m model) helpView() string {
	var lines []string
	for _, section := range []string{sectionNavigation, sectionActions} {
		lines = append(lines, ui.DialogTitleStyle.Render(section+":"))
		for _, b := range normalBindings {
			if b.section != section || b.desc == "" {
				continue
			}
			display := b.display
			if display == "" {
				display = strings.Join(b.keys, "/")
			}
			lines = append(lines, fmt.Sprintf("  %-14s%s", display, b.desc))
		}
		lines = append(lines, "")
	}
	lines = append(lines,
		ui.DialogTitleStyle.Render("Editing:"),
		"  enter         save changes",
		"  esc           cancel editing",
		"  ←/→           move cursor",
		"  ctrl+a/e      start/end of line",
		"  ctrl+w        delete word backward",
		"  ctrl+k/u      delete to end/start",
		"  ctrl+←/→      word navigation",
	)
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
