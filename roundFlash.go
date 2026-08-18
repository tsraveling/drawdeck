package main

import (
	"image/color"
	"strings"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

const (
	flashDuration   = 1400 * time.Millisecond
	flashFrameDur   = 100 * time.Millisecond
	flashFrameCount = int(flashDuration / flashFrameDur)
)

// bright cycle: pink → orange → yellow → green → cyan
var flashPalette = []color.Color{
	lipgloss.Color("205"),
	lipgloss.Color("208"),
	lipgloss.Color("220"),
	lipgloss.Color("82"),
	lipgloss.Color("51"),
}

type flashTickMsg struct{}

func flashTick() tea.Cmd {
	return tea.Tick(flashFrameDur, func(time.Time) tea.Msg { return flashTickMsg{} })
}

// announces a round for a fixed duration, strobing through the palette
type roundFlash struct {
	active bool
	text   string
	frame  int
}

func (f *roundFlash) start(text string) tea.Cmd {
	f.active = true
	f.text = text
	f.frame = 0
	return flashTick()
}

// advances a frame; returns false once the flash is finished
func (f *roundFlash) advance() (tea.Cmd, bool) {
	f.frame++
	if f.frame >= flashFrameCount {
		f.active = false
		return nil, false
	}
	return flashTick(), true
}

func (f roundFlash) view(width, height int) string {
	color := flashPalette[f.frame%len(flashPalette)]
	banner := lipgloss.NewStyle().
		Bold(true).
		Foreground(color).
		Render(spaced(f.text))

	return lipgloss.Place(width, height, lipgloss.Center, lipgloss.Center, banner)
}

// letterspacing gives the banner some presence without a figlet font
func spaced(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 {
			b.WriteRune(' ')
		}
		b.WriteRune(r)
	}
	return b.String()
}
