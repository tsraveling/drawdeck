package main

import (
	tea "charm.land/bubbletea/v2"
	"charm.land/bubbles/v2/textinput"
	"charm.land/lipgloss/v2"
)

type promptResult int

const (
	promptPending promptResult = iota
	promptCommit
	promptCancel
)

// reusable single-field input modal. Owns text, validation, and esc/enter;
// the caller decides what a commit means.
type textPrompt struct {
	title    string
	input    textinput.Model
	validate func(string) error
	err      error
}

func newTextPrompt(title, placeholder string, validate func(string) error) textPrompt {
	ti := textinput.New()
	ti.Placeholder = placeholder
	ti.SetWidth(modalWidth() - 8)
	ti.Focus()
	return textPrompt{title: title, input: ti, validate: validate}
}

// narrower than the view beneath it so the overlay reads as a modal
func modalWidth() int {
	return max(cfg.contentWidth()-8, minViewWidth-10)
}

func (p textPrompt) value() string { return p.input.Value() }

func (p textPrompt) Update(msg tea.Msg) (textPrompt, tea.Cmd, promptResult) {
	if km, ok := msg.(tea.KeyPressMsg); ok {
		switch km.String() {
		case "esc":
			return p, nil, promptCancel
		case "enter":
			if p.validate != nil {
				if err := p.validate(p.input.Value()); err != nil {
					p.err = err
					return p, nil, promptPending
				}
			}
			return p, nil, promptCommit
		}
		// any edit clears a stale validation error
		p.err = nil
	}

	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	return p, cmd, promptPending
}

func (p textPrompt) View() string {
	parts := []string{
		TitleStyle.Render(p.title),
		"",
		p.input.View(),
	}
	if p.err != nil {
		parts = append(parts, "", ErrorStyle.Render(p.err.Error()))
	}
	parts = append(parts, "", HelpStyle.Render("enter confirm  esc cancel"))

	return ModalStyle.Width(modalWidth()).
		Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}
