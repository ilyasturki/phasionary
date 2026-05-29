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

	shortcutBarValue := "Off"
	if cfg.ShowShortcutBar {
		shortcutBarValue = "On"
	}

	expandDescValue := "Off"
	if cfg.ExpandDescriptionsByDefault {
		expandDescValue = "On"
	}

	rows := []string{
		fmt.Sprintf("Status Display:  [%s]", statusValue),
		fmt.Sprintf("Priority Color:  [%s]", priorityValue),
		fmt.Sprintf("Shortcut Bar:    [%s]", shortcutBarValue),
		fmt.Sprintf("Descriptions:    [%s]", expandDescValue),
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
		ui.RenderHints([]ui.Hint{{Key: "space/tab", Label: "cycle"}, {Key: "q/esc/enter", Label: "close"}}),
	)
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
