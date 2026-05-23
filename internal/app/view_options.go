package app

import (
	"fmt"
	"strings"

	"phasionary/internal/config"
	"phasionary/internal/ui"
)

func (m model) optionsView() string {
	statusValue := "Text Labels"
	if m.deps.CfgManager.Get().StatusDisplay == config.StatusDisplayIcons {
		statusValue = "Icons"
	}
	lines := []string{
		ui.DialogTitleStyle.Render("Options"),
		"",
		fmt.Sprintf("> Status Display: [%s]", statusValue),
		"",
		ui.DialogHintStyle.Render("space/tab toggle | q/esc/enter close"),
	}
	return ui.HelpDialogStyle.Render(strings.Join(lines, "\n"))
}
