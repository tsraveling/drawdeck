package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"charm.land/bubbles/v2/progress"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	chargeTickDur = 25 * time.Millisecond

	// how long a hold must be sustained to draw; the tap window scales with it
	// so neither gesture is the strictly easier path
	chargeDuration = 1000 * time.Millisecond
	tapWindow      = 1000 * time.Millisecond

	tapEase    = 80 * time.Millisecond
	tapsToDraw = 3
)

type chargeTickMsg struct{}

func chargeTick() tea.Cmd {
	return tea.Tick(chargeTickDur, func(time.Time) tea.Msg { return chargeTickMsg{} })
}

// emitted when the user leaves a deck
type backToListMsg struct{}

type detailView struct {
	deck *deck
	bar  progress.Model

	// hold gesture
	holding   bool
	holdStart time.Time
	holdPct   float64

	// tap gesture
	taps     int
	tapStart time.Time
	tapPct   float64

	// set while a hold is in flight so its release does not also count as a tap
	holdClaimed bool

	flip       flip
	flipHeight int

	confirm confirm
	err     error
}

func makeDetailView(d *deck) detailView {
	return detailView{
		deck: d,
		bar:  progress.New(progress.WithoutPercentage(), progress.WithColors(ColorPrimary, ColorSelection)),
	}
}

func (v detailView) charging() bool {
	return v.holding || v.taps > 0
}

func (v detailView) chargePct() float64 {
	if v.holding {
		return v.holdPct
	}
	return v.tapPct
}

func (v *detailView) cancelCharge() {
	v.holding = false
	v.holdPct = 0
	v.taps = 0
	v.tapPct = 0
}

// re-reads from disk so an edit made in another window is respected
func (v *detailView) refresh() {
	d, err := loadDeck(v.deck.src)
	if err != nil {
		v.err = err
		return
	}
	v.err = nil
	v.deck = d
}

// picks a card, commits it to the file immediately, then animates the reveal
func (v *detailView) draw() tea.Cmd {
	v.cancelCharge()
	v.refresh()
	if v.err != nil {
		return nil
	}

	pool := v.deck.drawable()
	if len(pool) == 0 {
		return nil
	}

	idx := pool[rand.Intn(len(pool))]
	c := v.deck.cards[idx]

	v.deck.setLineChecked(c.line, true)
	v.deck.cards[idx].checked = true
	v.deck.setCurrent(c.title)
	if err := v.deck.write(); err != nil {
		v.err = err
		return nil
	}

	v.flipHeight = lipgloss.Height(v.renderCard())
	return v.flip.start()
}

func (v *detailView) reset() {
	v.refresh()
	if v.err != nil {
		return
	}
	v.deck.resetAll()
	if err := v.deck.write(); err != nil {
		v.err = err
	}
}

func (v detailView) Update(msg tea.Msg) (detailView, tea.Cmd) {
	if v.confirm.active() {
		switch v.confirm.handle(msg) {
		case confirmYes:
			v.confirm = confirm{}
			v.reset()
		case confirmNo:
			v.confirm = confirm{}
		}
		return v, nil
	}

	switch msg := msg.(type) {
	case flipTickMsg:
		cmd, running := v.flip.advance()
		if !running {
			return v, nil
		}
		return v, cmd

	case chargeTickMsg:
		return v.tickCharge()

	case tea.KeyReleaseMsg:
		if msg.Key().String() != "space" && msg.Key().String() != " " {
			return v, nil
		}
		// in hold-capable terminals the tap counter advances on release, so a
		// completed hold must not also register as a tap
		if v.holding {
			v.holding = false
			v.holdPct = 0
			if !v.holdClaimed {
				return v.registerTap()
			}
			v.holdClaimed = false
		}
		return v, nil

	case tea.KeyPressMsg:
		return v.handleKey(msg)
	}

	return v, nil
}

func (v detailView) handleKey(msg tea.KeyPressMsg) (detailView, tea.Cmd) {
	key := msg.String()

	if key == "space" || key == " " {
		if v.flip.active || v.deck.exhausted() {
			return v, nil
		}
		if msg.Key().IsRepeat {
			return v, nil
		}
		if cfg.holdSupported {
			// hold drives the bar; the tap counter advances on release instead
			if !v.holding {
				v.holding = true
				v.holdStart = time.Now()
				v.holdPct = 0
				return v, chargeTick()
			}
			return v, nil
		}
		return v.registerTap()
	}

	switch key {
	case "esc":
		if v.charging() {
			v.cancelCharge()
			return v, nil
		}
		return v, func() tea.Msg { return backToListMsg{} }
	case "q", "ctrl+c":
		if v.charging() {
			return v, nil
		}
		return v, tea.Quit
	case "ctrl+r":
		v.confirm = newConfirm(confirmResetDeck, "Are you sure you want to reset this deck?")
	}

	return v, nil
}

func (v detailView) registerTap() (detailView, tea.Cmd) {
	if v.flip.active || v.deck.exhausted() {
		return v, nil
	}

	var cmd tea.Cmd
	if v.taps == 0 {
		v.tapStart = time.Now()
		cmd = chargeTick()
	}
	v.taps++

	if v.taps >= tapsToDraw {
		v.holdClaimed = true
		return v, v.draw()
	}
	return v, cmd
}

func (v detailView) tickCharge() (detailView, tea.Cmd) {
	now := time.Now()

	if v.holding {
		v.holdPct = min(float64(now.Sub(v.holdStart))/float64(chargeDuration), 1)
		if v.holdPct >= 1 {
			v.holdClaimed = true
			return v, v.draw()
		}
		return v, chargeTick()
	}

	if v.taps > 0 {
		// hard window measured from the first tap, not sliding
		if now.Sub(v.tapStart) > tapWindow {
			v.taps = 0
			v.tapPct = 0
			return v, nil
		}
		target := float64(v.taps) / float64(tapsToDraw)
		step := float64(chargeTickDur) / float64(tapEase)
		if v.tapPct < target {
			v.tapPct = min(v.tapPct+step, target)
		}
		return v, chargeTick()
	}

	return v, nil
}

func (v detailView) cardWidth() int {
	return cfg.contentWidth()
}

func (v detailView) renderCard() string {
	c := v.deck.currentCard()
	if c == nil {
		return ""
	}

	inner := v.cardWidth() - 6
	parts := []string{CardTitleStyle.Render(wrapText(c.title, inner))}
	if c.notes != "" {
		parts = append(parts, "", DimStyle.Render(wrapText(c.notes, inner)))
	}

	return CardStyle.Width(v.cardWidth()).
		Render(lipgloss.JoinVertical(lipgloss.Left, parts...))
}

func (v detailView) View() string {
	w := cfg.viewWidth()

	var b strings.Builder
	b.WriteString(TitleStyle.Render(v.deck.title))
	b.WriteString("  ")
	b.WriteString(CountStyle.Render(fmt.Sprintf("(%d / %d)", v.deck.doneCount(), len(v.deck.cards))))
	b.WriteString("\n\n")

	switch {
	case v.flip.active:
		b.WriteString(LabelStyle.Render("Current card:"))
		b.WriteString("\n")
		b.WriteString(v.flip.view(v.cardWidth(), max(v.flipHeight, 3)))
	case v.deck.currentCard() != nil:
		b.WriteString(LabelStyle.Render("Current card:"))
		b.WriteString("\n")
		b.WriteString(v.renderCard())
	default:
		b.WriteString(DimStyle.Render(v.drawPrompt()))
	}
	b.WriteString("\n\n")

	if v.charging() {
		v.bar.SetWidth(v.cardWidth())
		b.WriteString(v.bar.ViewAs(v.chargePct()))
	} else if v.deck.currentCard() != nil {
		b.WriteString(DimStyle.Render(v.drawPrompt()))
	}
	b.WriteString("\n\n")

	if v.confirm.active() {
		b.WriteString(lipgloss.PlaceHorizontal(cfg.contentWidth(), lipgloss.Center, v.confirm.View()))
		b.WriteString("\n\n")
	}
	if v.err != nil {
		b.WriteString(ErrorStyle.Render(v.err.Error()))
		b.WriteString("\n\n")
	}

	b.WriteString(HelpStyle.Render("esc back  ctrl+r reset  q quit"))

	return ViewStyle.Width(w).Render(b.String())
}

// the prompt doubles as the help for whichever gesture is primary
func (v detailView) drawPrompt() string {
	if v.deck.exhausted() {
		return "Deck exhausted — ctrl+r to reset"
	}
	if cfg.holdSupported {
		return "hold space to draw card"
	}
	return "tap space ×3 to draw card"
}

func wrapText(s string, width int) string {
	if width < 1 {
		return s
	}
	var out []string
	for para := range strings.SplitSeq(s, "\n") {
		line := ""
		for word := range strings.FieldsSeq(para) {
			switch {
			case line == "":
				line = word
			case len(line)+1+len(word) <= width:
				line += " " + word
			default:
				out = append(out, line)
				line = word
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}
