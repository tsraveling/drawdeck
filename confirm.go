package main

import tea "charm.land/bubbletea/v2"

// @region ui:confirm -- YES NO CONFIRM
type confirmKind int

const (
	confirmNone confirmKind = iota
	confirmDeleteDeck
	confirmResetDeck
	confirmStartTournament
	confirmLeaveTournament
	confirmQuitTournament
)

type confirmResult int

const (
	confirmPending confirmResult = iota
	confirmYes
	confirmNo
)

// reusable yes/no. Owns the prompt text and key interpretation only;
// placement stays with the caller.
type confirm struct {
	kind   confirmKind
	prompt string
}

func newConfirm(kind confirmKind, prompt string) confirm {
	return confirm{kind: kind, prompt: prompt}
}

func (c confirm) active() bool { return c.kind != confirmNone }

func (c confirm) handle(msg tea.Msg) confirmResult {
	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return confirmPending
	}
	switch km.String() {
	case "y", "Y":
		return confirmYes
	case "n", "N", "esc", "q":
		return confirmNo
	}
	return confirmPending
}

func (c confirm) View() string {
	return ErrorStyle.Render(c.prompt + " y/n")
}
