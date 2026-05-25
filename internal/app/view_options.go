package app

import (
	"fmt"
	"strings"

	"phasionary/internal/config"
	"phasionary/internal/ui"
)

func (m model) optionsView() string {
	cfg := m.deps.CfgManager.Get()

	statusValue := "Text Labels"
	if cfg.StatusDisplay == config.StatusDisplayIcons {
		statusValue = "Icons"
	}

	priorityValue := "Full Text"
	switch cfg.PriorityColor {
	case config.PriorityColorIcon:
		priorityValue = "Icon Only"
	case config.PriorityColorNone:
		priorityValue = "None"
	}

	rows := []string{
		fmt.Sprintf("Status Display:  [%s]", statusValue),
		fmt.Sprintf("Priority Color:  [%s]", priorityValue),
	}

	lines := []string{
		ui.DialogTitleStyle.Render("Options"),
		"",
	}
	for i, row := range rows {
		if i == m.ui.Options.selectedOption {
			lines = append(lines, "> "+row)
		} else {
			lines = append(lines, "  "+row)
		}
	}
	lines = append(lines,
		"",
		ui.DialogHintStyle.Render("space/tab cycle | q/esc/enter close"),
	)
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
