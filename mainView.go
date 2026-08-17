package main

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

type mainViewModel struct {
	width  int
	height int
	quit   bool
}

func makeMainViewModel() (mainViewModel, tea.Cmd) {
	return mainViewModel{}, nil
}

func (m mainViewModel) Init() tea.Cmd {
	return nil
}

func (m mainViewModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			m.quit = true
			return m, tea.Quit
		}
	}
	return m, nil
}

func (m mainViewModel) View() string {
	if m.quit {
		return ""
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render("drawdeck"))
	b.WriteString("\n\n")
	b.WriteString(UnselectedItemStyle.Render(fmt.Sprintf("v%s", Version)))
	b.WriteString("\n\n")
	b.WriteString(HelpStyle.Render("q: quit"))

	return ViewStyle.Render(b.String())
}
