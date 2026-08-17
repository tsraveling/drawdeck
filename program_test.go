package main

import (
	"bytes"
	"strings"
	"testing"
	tea "charm.land/bubbletea/v2"
)

// drives the real program loop end to end: a full render must reach the writer
func TestProgramRendersAndQuits(t *testing.T) {
	deckPath := writeTemp(t, sample)
	reg := testRegistry(t, deckPath)

	var out bytes.Buffer
	p := tea.NewProgram(
		makeRootModel(reg, ""),
		tea.WithInput(strings.NewReader("q")),
		tea.WithOutput(&out),
		tea.WithWindowSize(100, 40),
	)
	if _, err := p.Run(); err != nil {
		t.Fatalf("program error: %v", err)
	}

	got := out.String()
	if !strings.Contains(got, "Example Deck") {
		t.Errorf("deck title never rendered; got %d bytes:\n%q", len(got), got)
	}
	if !strings.Contains(got, "DRAWDECK") {
		t.Error("app title never rendered")
	}
}
