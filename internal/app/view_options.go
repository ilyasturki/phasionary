package app

import (
	"strings"

	"charm.land/lipgloss/v2"

	"phasionary/internal/config"
	"phasionary/internal/ui"
)

type optionValue struct {
	key   string
	label string
}

// optionSpec is one row of the Options dialog. The values list drives both the
// rendering and what h/l cycles through, so the two cannot drift.
type optionSpec struct {
	name    string
	detail  string
	values  []optionValue
	current func(cfg config.Config) string
	apply   func(m *model, key string)
}

var optionSpecs = []optionSpec{
	{
		name:   "Status Display",
		detail: "the status column as [x] glyphs or as words",
		values: []optionValue{
			{key: config.StatusDisplayIcons, label: "Icons"},
			{key: config.StatusDisplayText, label: "Text"},
		},
		// An empty field is a hand-edited config; it renders as the default, so
		// report the default here too rather than highlighting nothing.
		current: func(cfg config.Config) string {
			return orDefault(cfg.StatusDisplay, config.StatusDisplayText)
		},
		apply: func(m *model, key string) {
			_ = m.deps.CfgManager.Update(func(cfg *config.Config) { cfg.StatusDisplay = key })
			// Icons vs. text change the status column width, so cached row
			// heights are stale. (PriorityColor below only recolors.)
			m.invalidateLayout()
		},
	},
	{
		name:   "Priority Color",
		detail: "color the whole title, the priority arrow alone, or nothing",
		values: []optionValue{
			{key: config.PriorityColorFull, label: "Full"},
			{key: config.PriorityColorIcon, label: "Icon"},
			{key: config.PriorityColorNone, label: "None"},
		},
		current: func(cfg config.Config) string {
			return orDefault(cfg.PriorityColor, config.PriorityColorFull)
		},
		apply: func(m *model, key string) {
			_ = m.deps.CfgManager.Update(func(cfg *config.Config) { cfg.PriorityColor = key })
		},
	},
	{
		name:   "Shortcut Bar",
		detail: "the key hints along the bottom of the screen",
		values: onOff,
		current: func(cfg config.Config) string {
			return onOffKey(cfg.ShowShortcutBar)
		},
		apply: func(m *model, key string) {
			_ = m.deps.CfgManager.Update(func(cfg *config.Config) { cfg.ShowShortcutBar = key == "on" })
			// The bar is hidden while Options is open, so the layout under us
			// hasn't actually changed yet. handleOptionsKey's exit branch calls
			// ensureVisible against the post-toggle layout once Options closes.
		},
	},
	{
		name:   "Descriptions",
		detail: "whether descriptions start expanded; zd toggles them",
		values: onOff,
		current: func(cfg config.Config) string {
			return onOffKey(cfg.ExpandDescriptionsByDefault)
		},
		apply: func(m *model, key string) {
			on := key == "on"
			_ = m.deps.CfgManager.Update(func(cfg *config.Config) { cfg.ExpandDescriptionsByDefault = on })
			m.ui.Screen.ExpandDescriptions = on
			m.rebuildPositions()
		},
	},
}

var onOff = []optionValue{{key: "on", label: "On"}, {key: "off", label: "Off"}}

func orDefault(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

func onOffKey(on bool) string {
	if on {
		return "on"
	}
	return "off"
}

// cycleSelectedOption steps the selected option delta places through its values,
// wrapping at both ends so h and l are inverses.
func (m *model) cycleSelectedOption(delta int) {
	spec := optionSpecs[m.ui.Options.selectedOption]
	cur := spec.current(m.deps.CfgManager.Get())
	i := 0
	for j, v := range spec.values {
		if v.key == cur {
			i = j
		}
	}
	n := len(spec.values)
	spec.apply(m, spec.values[((i+delta)%n+n)%n].key)
}

func (m model) optionsView() string {
	cfg := m.deps.CfgManager.Get()
	width := m.dialogWidth()

	lines := []string{ui.DialogTitleStyle.Render("Options"), ""}
	for i, spec := range optionSpecs {
		if i > 0 {
			lines = append(lines, "")
		}
		lines = append(lines,
			m.optionRow(spec, cfg, i == m.ui.Options.selectedOption, width),
			ui.DialogDetailStyle.Render("    "+spec.detail),
		)
	}
	lines = append(lines,
		"",
		ui.RenderHintsToWidth([]ui.Hint{
			{Key: "h/l", Label: "change"},
			{Key: "j/k", Label: "move"},
			{Key: "q/esc/enter", Label: "close"},
		}, width),
	)
	return m.dialogStyle().Render(strings.Join(lines, "\n"))
}

// The value in effect reads bold against its faint alternatives, so which one
// is live survives the reverse band the cursor row is drawn in.
func (m model) optionRow(spec optionSpec, cfg config.Config, focused bool, width int) string {
	base := lipgloss.NewStyle()
	if focused {
		base = ui.GetSelectedStyle(m.ui.Screen.WindowFocused)
	}

	cur := spec.current(cfg)
	values := make([]string, len(spec.values))
	for i, v := range spec.values {
		style := base.Faint(true)
		if v.key == cur {
			style = base.Bold(true)
		}
		values[i] = style.Render(v.label)
	}

	left := base.Render("  ") + ui.DialogKey(base, focused).Render(spec.name)
	right := strings.Join(values, base.Render("  ")) + base.Render("  ")
	// Values sit flush right, so the row — and the band behind it when it is
	// selected — spans the dialog instead of trailing off into empty space.
	gap := max(width-lipgloss.Width(left)-lipgloss.Width(right), 1)
	return left + base.Render(strings.Repeat(" ", gap)) + right
}
