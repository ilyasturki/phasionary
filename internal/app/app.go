package app

import (
	"fmt"

	tea "github.com/charmbracelet/bubbletea"

	"phasionary/internal/data"
	"phasionary/internal/domain"
	"phasionary/internal/ui"
)

type model struct {
	projectName string
	width       int
	height      int
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c":
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m model) View() string {
	header := ui.HeaderStyle.Render("Phasionary")
	project := ui.MutedStyle.Render(fmt.Sprintf("Project: %s", m.projectName))
	body := "\n\n" + ui.MutedStyle.Render("Ultra-minimal TUI placeholder. Press q to quit.")
	return header + "  " + project + body + "\n"
}

func Run(dataDir string) error {
	store := data.NewStore(dataDir)
	project, err := ensureProject(store)
	if err != nil {
		return err
	}
	program := tea.NewProgram(model{projectName: project.Name})
	_, err = program.Run()
	return err
}

func ensureProject(store *data.Store) (domain.Project, error) {
	if err := store.Ensure(); err != nil {
		return domain.Project{}, err
	}
	projects, err := store.ListProjects()
	if err != nil {
		return domain.Project{}, err
	}
	if len(projects) == 0 {
		return store.InitDefault()
	}
	return projects[0], nil
}
