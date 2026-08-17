package main

import (
	_ "embed"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/viewport"
	"charm.land/lipgloss/v2"
)

//go:embed HELP.md
var helpSource string

var (
	helpH1     = lipgloss.NewStyle().Bold(true).Foreground(ColorPrimary)
	helpH2     = lipgloss.NewStyle().Bold(true).Foreground(ColorSelection)
	helpCode   = lipgloss.NewStyle().Foreground(ColorActive)
	helpBullet = lipgloss.NewStyle().Foreground(ColorDimBright)
	helpText   = lipgloss.NewStyle().Foreground(ColorBasic)
)

// minimal markdown renderer: headers, bullets, fenced blocks, and code spans
func renderHelp(src string, width int) string {
	var out []string
	inFence := false

	for line := range strings.SplitSeq(src, "\n") {
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
			continue
		}
		if inFence {
			out = append(out, helpCode.Render("  "+line))
			continue
		}

		switch {
		case trimmed == "":
			out = append(out, "")
		case strings.HasPrefix(trimmed, "## "):
			out = append(out, helpH2.Render(strings.TrimPrefix(trimmed, "## ")))
		case strings.HasPrefix(trimmed, "# "):
			out = append(out, helpH1.Render(strings.TrimPrefix(trimmed, "# ")))
		case strings.HasPrefix(trimmed, "- "):
			out = append(out, helpBullet.Render("  • ")+inlineCode(strings.TrimPrefix(trimmed, "- ")))
		default:
			out = append(out, helpText.Render(wrapText(inlineCode(trimmed), width)))
		}
	}

	return strings.Join(out, "\n")
}

// styles `code` spans, leaving surrounding text alone
func inlineCode(s string) string {
	parts := strings.Split(s, "`")
	var b strings.Builder
	for i, p := range parts {
		if i%2 == 1 {
			b.WriteString(helpCode.Render(p))
		} else {
			b.WriteString(helpText.Render(p))
		}
	}
	return b.String()
}

type helpModel struct {
	vp viewport.Model
}

func makeHelpModel() helpModel {
	vp := viewport.New()
	vp.SetWidth(maxViewWidth)
	vp.SetHeight(24)
	vp.SetContent(renderHelp(helpSource, maxViewWidth-4))
	return helpModel{vp: vp}
}

func (m helpModel) Init() tea.Cmd { return nil }

func (m helpModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		w := max(minViewWidth, min(msg.Width, maxViewWidth))
		m.vp.SetWidth(w)
		m.vp.SetHeight(max(msg.Height-4, 5))
		m.vp.SetContent(renderHelp(helpSource, w-4))
		return m, nil
	case tea.KeyPressMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

func (m helpModel) View() tea.View {
	body := m.vp.View() + "\n" + HelpStyle.Render("↑↓ scroll  q quit")
	v := tea.NewView(lipgloss.NewStyle().Padding(1, 2).Render(body))
	v.AltScreen = true
	return v
}
