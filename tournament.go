package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	// how long the chosen card stays highlighted before advancing
	pickHighlightDur = 450 * time.Millisecond

	// below this terminal width the two cards stack vertically
	stackBelowWidth = 70

	cardGutter    = 4
	maxNotesLines = 4
)

type pickTickMsg struct{}

func pickTick() tea.Cmd {
	return tea.Tick(pickHighlightDur, func(time.Time) tea.Msg { return pickTickMsg{} })
}

type tournamentPhase int

const (
	phaseFlash tournamentPhase = iota
	phaseChallenge
	phaseVictory
)

// what the pending exit confirm will do on yes
type exitIntent int

const (
	exitNone exitIntent = iota
	exitToList
	exitQuit
)

// emitted when a tournament ends or is abandoned
type leaveTournamentMsg struct{}

// @region tourney:model -- TOURNAMENT STATE
type tournament struct {
	deck *deck

	phase   tournamentPhase
	round   []card
	idx     int
	winners []card

	flash roundFlash

	// -1 when nothing is pending; otherwise the side awaiting its highlight
	highlight int

	confirm confirm
	intent  exitIntent
}

func makeTournament(d *deck) (tournament, tea.Cmd) {
	t := tournament{deck: d, highlight: -1}
	t.round = shuffled(d.cards)
	t.phase = phaseFlash
	return t, t.flash.start(flashText(len(t.round)))
}

func shuffled(in []card) []card {
	out := make([]card, len(in))
	copy(out, in)
	rand.Shuffle(len(out), func(i, j int) { out[i], out[j] = out[j], out[i] })
	return out
}

// @region tourney:bracket -- ROUND PAIRING AND ADVANCE
// challenges plus, when the round is odd, one keep/discard decision
func (t tournament) totalSteps() int {
	return len(t.round)/2 + len(t.round)%2
}

// the trailing card of an odd round has no opponent
func (t tournament) isOddStep() bool {
	return 2*t.idx+1 >= len(t.round)
}

func (t tournament) leftCard() card  { return t.round[2*t.idx] }
func (t tournament) rightCard() card { return t.round[2*t.idx+1] }

func roundName(n int) string {
	switch n {
	case 2:
		return "Final"
	case 4:
		return "Semifinals"
	case 8:
		return "Quarterfinals"
	}
	return fmt.Sprintf("Round of %d", n)
}

func flashText(n int) string {
	if n == 2 {
		return "FINAL ROUND!"
	}
	return strings.ToUpper(roundName(n)) + "!"
}

// the Final has only one challenge, so it drops the counter
func (t tournament) header() string {
	n := len(t.round)
	if n == 2 {
		return roundName(n)
	}
	return fmt.Sprintf("%s: %d / %d", roundName(n), t.idx+1, t.totalSteps())
}

// @region tourney:pick -- PICK COMMIT
// applies the highlighted pick and moves to the next step or round
func (t *tournament) commitPick() tea.Cmd {
	side := t.highlight
	t.highlight = -1

	if t.isOddStep() {
		// left keeps the odd card, right discards it
		if side == 0 {
			t.winners = append(t.winners, t.leftCard())
		}
	} else if side == 0 {
		t.winners = append(t.winners, t.leftCard())
	} else {
		t.winners = append(t.winners, t.rightCard())
	}

	t.idx++
	if t.idx < t.totalSteps() {
		return nil
	}
	return t.endRound()
}

func (t *tournament) endRound() tea.Cmd {
	t.round = shuffled(t.winners)
	t.winners = nil
	t.idx = 0

	// a lone survivor wins outright; no final is played
	if len(t.round) <= 1 {
		return t.declareWinner()
	}

	t.phase = phaseFlash
	return t.flash.start(flashText(len(t.round)))
}

// @region tourney:victory -- WINNER SCREEN
func (t *tournament) declareWinner() tea.Cmd {
	t.phase = phaseVictory
	if len(t.round) == 0 {
		return nil
	}

	w := t.round[0]
	t.deck.setWinner(w.title)
	t.deck.write()
	return t.flash.start(strings.ToUpper(w.title) + "!")
}

// @region tourney:keys -- TOURNAMENT INPUT
func (t tournament) Update(msg tea.Msg) (tournament, tea.Cmd) {
	if t.confirm.active() {
		kind := t.confirm.kind
		switch t.confirm.handle(msg) {
		case confirmYes:
			intent := t.intent
			t.confirm = confirm{}
			t.intent = exitNone

			if kind == confirmResetDeck {
				t.deck.resetAll()
				t.deck.write()
				return t, func() tea.Msg { return leaveTournamentMsg{} }
			}
			if intent == exitQuit {
				return t, tea.Quit
			}
			return t, func() tea.Msg { return leaveTournamentMsg{} }
		case confirmNo:
			t.confirm = confirm{}
			t.intent = exitNone
		}
		return t, nil
	}

	switch msg.(type) {
	case flashTickMsg:
		cmd, running := t.flash.advance()
		if running {
			return t, cmd
		}
		if t.phase == phaseFlash {
			t.phase = phaseChallenge
		}
		return t, nil

	case pickTickMsg:
		return t, t.commitPick()
	}

	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return t, nil
	}

	// the winner is already on disk, so there is no progress left to protect
	if t.phase == phaseVictory {
		switch km.String() {
		case "q", "ctrl+c":
			return t, tea.Quit
		case "esc":
			return t, func() tea.Msg { return leaveTournamentMsg{} }
		case "ctrl+r":
			t.confirm = newConfirm(confirmResetDeck, "Are you sure you want to reset this deck?")
		}
		return t, nil
	}

	switch km.String() {
	case "esc":
		t.intent = exitToList
		t.confirm = newConfirm(confirmLeaveTournament,
			"Leaving will lose your progress on the tournament. Continue?")
		return t, nil
	case "q", "ctrl+c":
		t.intent = exitQuit
		t.confirm = newConfirm(confirmQuitTournament,
			"Quitting will lose your progress on the tournament. Continue?")
		return t, nil
	}

	// picks only land during a challenge, and never while one is resolving
	if t.phase != phaseChallenge || t.highlight >= 0 {
		return t, nil
	}

	switch km.String() {
	case "left", "h":
		t.highlight = 0
		return t, pickTick()
	case "right", "l":
		t.highlight = 1
		return t, pickTick()
	}

	return t, nil
}

func (t tournament) View() string {
	w := cfg.tournamentWidth()
	inner := w - 4
	height := max(cfg.wh-6, 12)

	if t.phase == phaseVictory {
		if t.flash.active {
			return ViewStyle.Width(w).Render(t.flash.view(inner, height))
		}
		body := renderVictory(t.deck, inner)
		if t.confirm.active() {
			body += "\n\n" + lipgloss.PlaceHorizontal(inner, lipgloss.Center, t.confirm.View())
		}
		return ViewStyle.Width(w).Render(body)
	}

	if t.phase == phaseFlash {
		return ViewStyle.Width(w).Render(t.flash.view(inner, height))
	}

	var b strings.Builder
	b.WriteString(TitleStyle.Render(t.deck.title))
	b.WriteString("  ")
	b.WriteString(CountStyle.Render(t.header()))
	b.WriteString("\n\n")
	b.WriteString(t.renderChallenge(inner))
	b.WriteString("\n\n")

	if t.confirm.active() {
		b.WriteString(lipgloss.PlaceHorizontal(inner, lipgloss.Center, t.confirm.View()))
		b.WriteString("\n\n")
	}

	help := "←→/hl pick  esc leave  q quit"
	if t.isOddStep() {
		help = "← keep  → discard  esc leave  q quit"
	}
	b.WriteString(HelpStyle.Render(help))

	return ViewStyle.Width(w).Render(b.String())
}

// @region tourney:render -- CHALLENGE RENDER
func (t tournament) renderChallenge(inner int) string {
	stacked := cfg.ww < stackBelowWidth

	cardW := inner
	if !stacked {
		cardW = (inner - cardGutter) / 2
	}

	leftBody, leftStyle := cardBody(t.leftCard(), cardW, t.highlight == 0)

	var rightBody string
	var rightStyle lipgloss.Style
	if t.isOddStep() {
		rightBody, rightStyle = discardBody(cardW, t.highlight == 1)
	} else {
		rightBody, rightStyle = cardBody(t.rightCard(), cardW, t.highlight == 1)
	}

	// side by side, both boxes match the taller body so the pair reads as a pair;
	// stacked, that padding would only waste rows
	if !stacked {
		h := max(lipgloss.Height(leftBody), lipgloss.Height(rightBody))
		leftBody = padLines(leftBody, h)
		rightBody = padLines(rightBody, h)
	}

	left := leftStyle.Width(cardW).Render(leftBody)
	right := rightStyle.Width(cardW).Render(rightBody)

	if stacked {
		return lipgloss.JoinVertical(lipgloss.Left, left, "", right)
	}
	gutter := strings.Repeat(" ", cardGutter)
	return lipgloss.JoinHorizontal(lipgloss.Top, left, gutter, right)
}

// the inner text of a card plus the box style it should wear
func cardBody(c card, w int, highlighted bool) (string, lipgloss.Style) {
	style := CardStyle
	titleStyle := CardTitleStyle
	if highlighted {
		style = CardStyle.BorderForeground(ColorActive)
		titleStyle = CardTitleStyle.Foreground(ColorActive)
	}

	textW := w - 6
	parts := []string{titleStyle.Render(wrapText(c.title, textW))}
	if c.notes != "" {
		parts = append(parts, "", DimStyle.Render(truncateLines(wrapText(c.notes, textW), maxNotesLines)))
	}

	return lipgloss.JoinVertical(lipgloss.Left, parts...), style
}

func discardBody(w int, highlighted bool) (string, lipgloss.Style) {
	c := ColorMuted
	if highlighted {
		c = ColorError
	}
	label := lipgloss.NewStyle().Bold(true).Foreground(c).Render("DISCARD")
	return lipgloss.PlaceHorizontal(w-6, lipgloss.Center, label), CardStyle.BorderForeground(c)
}

func renderChallengeCard(c card, w int, highlighted bool) string {
	body, style := cardBody(c, w, highlighted)
	return style.Width(w).Render(body)
}

func padLines(s string, h int) string {
	for lipgloss.Height(s) < h {
		s += "\n"
	}
	return s
}

func renderVictory(d *deck, w int) string {
	var body string
	for _, c := range d.cards {
		if c.title == d.winner {
			body = renderChallengeCard(c, w, false)
			break
		}
	}
	if body == "" {
		body = renderChallengeCard(card{title: d.winner}, w, false)
	}

	banner := lipgloss.NewStyle().Bold(true).Foreground(ColorActive).Render(spaced("WINNER"))

	return lipgloss.JoinVertical(lipgloss.Left,
		TitleStyle.Render(d.title),
		"",
		lipgloss.PlaceHorizontal(w, lipgloss.Center, banner),
		"",
		body,
		"",
		HelpStyle.Render("ctrl+r reset  esc back  q quit"),
	)
}

func truncateLines(s string, n int) string {
	lines := strings.Split(s, "\n")
	if len(lines) <= n {
		return s
	}
	return strings.Join(lines[:n], "\n") + " …"
}
