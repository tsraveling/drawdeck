package main

import (
	"math"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// ~400ms total at ~22fps
const (
	flipFrames   = 9
	flipFrameDur = 45 * time.Millisecond
)

type flipTickMsg struct{}

func flipTick() tea.Cmd {
	return tea.Tick(flipFrameDur, func(time.Time) tea.Msg { return flipTickMsg{} })
}

// horizontal squash-and-expand. Every transition runs identically, including
// the first draw of a session, which flips from a card back.
type flip struct {
	active bool
	frame  int
}

func (f *flip) start() tea.Cmd {
	f.active = true
	f.frame = 0
	return flipTick()
}

// advances a frame; returns false once the animation is finished
func (f *flip) advance() (tea.Cmd, bool) {
	f.frame++
	if f.frame >= flipFrames {
		f.active = false
		return nil, false
	}
	return flipTick(), true
}

// |cos| sweeps full → edge-on → full, which reads as a rotation
func flipWidth(frame, full int) int {
	t := float64(frame) / float64(flipFrames-1)
	frac := math.Abs(math.Cos(math.Pi * t))
	return max(int(frac*float64(full)), 1)
}

// the hatched card back, drawn at the frame's interpolated width and
// centered in the full width so the card appears to pivot in place
func (f flip) view(full, height int) string {
	w := flipWidth(f.frame, full)
	body := cardBack(w, height)
	return lipgloss.PlaceHorizontal(full, lipgloss.Center, body)
}

func cardBack(w, height int) string {
	style := lipgloss.NewStyle().Foreground(ColorCardBack)

	// too narrow for a border: draw the edge-on sliver
	if w < 4 {
		col := strings.TrimRight(strings.Repeat("│\n", height), "\n")
		return style.Render(col)
	}

	inner := w - 2
	rows := max(height-2, 1)

	var b strings.Builder
	for i := range rows {
		if i > 0 {
			b.WriteString("\n")
		}
		// offset each row so the hatch reads as a weave
		row := strings.Repeat("▚▞", inner)
		b.WriteString(trimToWidth(row, inner, i%2))
	}

	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder(), true).
		BorderForeground(ColorCardBack).
		Foreground(ColorCardBack).
		Render(b.String())
}

// takes `width` runes from `s` starting at `offset`
func trimToWidth(s string, width, offset int) string {
	r := []rune(s)
	if offset >= len(r) {
		offset = 0
	}
	r = r[offset:]
	if len(r) > width {
		r = r[:width]
	}
	for len(r) < width {
		r = append(r, ' ')
	}
	return string(r)
}
