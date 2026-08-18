package main

import (
	"fmt"
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
)

func deckOf(t *testing.T, n int) *deck {
	t.Helper()
	var b strings.Builder
	b.WriteString("# Bracket\n\n")
	for i := range n {
		fmt.Fprintf(&b, "- [ ] card %d\n", i)
	}
	d, err := loadDeck(writeTemp(t, b.String()))
	if err != nil {
		t.Fatal(err)
	}
	return d
}

// drives one step by picking a side, skipping the highlight delay
func pick(tr tournament, side int) tournament {
	tr.highlight = side
	tr.commitPick()
	return tr
}

func TestRoundNaming(t *testing.T) {
	cases := map[int]string{2: "Final", 4: "Semifinals", 8: "Quarterfinals", 33: "Round of 33", 17: "Round of 17"}
	for n, want := range cases {
		if got := roundName(n); got != want {
			t.Errorf("roundName(%d) = %q, want %q", n, got, want)
		}
	}
	if got := flashText(2); got != "FINAL ROUND!" {
		t.Errorf("flashText(2) = %q", got)
	}
	if got := flashText(8); got != "QUARTERFINALS!" {
		t.Errorf("flashText(8) = %q", got)
	}
	if got := flashText(33); got != "ROUND OF 33!" {
		t.Errorf("flashText(33) = %q", got)
	}
}

func TestStepCounts(t *testing.T) {
	cases := map[int]int{33: 17, 32: 16, 6: 3, 3: 2, 2: 1}
	for cards, want := range cases {
		tr := tournament{round: make([]card, cards)}
		if got := tr.totalSteps(); got != want {
			t.Errorf("%d cards -> %d steps, want %d", cards, got, want)
		}
	}
}

// the header drops its counter only on the Final
func TestHeaderFormat(t *testing.T) {
	tr := tournament{round: make([]card, 33)}
	if got := tr.header(); got != "Round of 33: 1 / 17" {
		t.Errorf("header = %q", got)
	}
	tr = tournament{round: make([]card, 2)}
	if got := tr.header(); got != "Final" {
		t.Errorf("final header = %q", got)
	}
	tr = tournament{round: make([]card, 8)}
	if got := tr.header(); got != "Quarterfinals: 1 / 4" {
		t.Errorf("quarterfinal header = %q", got)
	}
}

// 6 cards must walk 6 -> 3 -> 2 -> 1 when every odd card is kept
func TestBracketProgression(t *testing.T) {
	d := deckOf(t, 6)
	tr, _ := makeTournament(d)
	tr.phase = phaseChallenge

	sizes := []int{len(tr.round)}
	for tr.phase != phaseVictory {
		before := len(tr.round)
		for range tr.totalSteps() {
			tr = pick(tr, 0) // always keep left, which also keeps odd cards
		}
		if len(tr.round) == before {
			t.Fatalf("round size stuck at %d", before)
		}
		sizes = append(sizes, len(tr.round))
	}

	want := []int{6, 3, 2, 1}
	if fmt.Sprint(sizes) != fmt.Sprint(want) {
		t.Errorf("round sizes = %v, want %v", sizes, want)
	}
	if d.winner == "" {
		t.Error("no winner recorded")
	}
}

// discarding the odd card of a 3-round leaves one survivor, who wins outright
func TestOddDiscardEndsTournament(t *testing.T) {
	d := deckOf(t, 3)
	tr, _ := makeTournament(d)
	tr.phase = phaseChallenge

	tr = pick(tr, 0) // challenge: keep left
	if !tr.isOddStep() {
		t.Fatal("step 2 of a 3-card round should be the odd decision")
	}
	tr = pick(tr, 1) // discard the odd card

	if tr.phase != phaseVictory {
		t.Errorf("phase = %v, want victory", tr.phase)
	}
	if len(tr.round) != 1 {
		t.Errorf("survivors = %d, want 1", len(tr.round))
	}
}

// keeping the odd card of a 3-round produces a normal final
func TestOddKeepProducesFinal(t *testing.T) {
	d := deckOf(t, 3)
	tr, _ := makeTournament(d)
	tr.phase = phaseChallenge

	tr = pick(tr, 0)
	tr = pick(tr, 0) // keep the odd card

	if tr.phase == phaseVictory {
		t.Fatal("keeping the odd card should not end the tournament")
	}
	if len(tr.round) != 2 {
		t.Errorf("final round has %d cards, want 2", len(tr.round))
	}
	if got := tr.header(); got != "Final" {
		t.Errorf("header = %q, want Final", got)
	}
}

func TestWinnerPersisted(t *testing.T) {
	d := deckOf(t, 4)
	tr, _ := makeTournament(d)
	tr.phase = phaseChallenge

	for tr.phase != phaseVictory {
		for range tr.totalSteps() {
			tr = pick(tr, 0)
		}
	}

	reloaded, err := loadDeck(d.src)
	if err != nil {
		t.Fatal(err)
	}
	if reloaded.winner == "" {
		t.Fatal("winner not written to frontmatter")
	}
	if reloaded.doneCount() != 0 {
		t.Errorf("tournament should not check any boxes, got %d", reloaded.doneCount())
	}
}

func TestTournamentViewsRender(t *testing.T) {
	cfg.setSize(100, 40)
	d := deckOf(t, 5)
	tr, _ := makeTournament(d)

	if out := tr.View(); !strings.Contains(out, spaced("ROUND OF 5!")) {
		t.Errorf("flash did not render:\n%s", out)
	}

	tr.phase = phaseChallenge
	if out := tr.View(); !strings.Contains(out, "Round of 5: 1 / 3") {
		t.Errorf("challenge header missing:\n%s", out)
	}

	// jump to the odd step and confirm the discard box appears
	tr.idx = tr.totalSteps() - 1
	if !tr.isOddStep() {
		t.Fatal("expected an odd step")
	}
	out := tr.View()
	if !strings.Contains(out, "DISCARD") {
		t.Errorf("discard box missing:\n%s", out)
	}
	if !strings.Contains(out, "keep") {
		t.Errorf("odd-step help missing:\n%s", out)
	}
}

func TestStackedLayoutBelowThreshold(t *testing.T) {
	cfg.setSize(60, 40)
	defer cfg.setSize(100, 40)

	d := deckOf(t, 4)
	tr, _ := makeTournament(d)
	tr.phase = phaseChallenge

	if out := tr.View(); out == "" {
		t.Error("stacked layout rendered empty")
	}
}

// a deck with a winner must show victory instead of the draw mechanic
func TestDetailLocksOnWinner(t *testing.T) {
	cfg.setSize(100, 40)
	d := deckOf(t, 4)
	d.setWinner("card 2")
	d.write()

	v := makeDetailView(d)
	out := v.View()
	if !strings.Contains(out, "W I N N E R") {
		t.Errorf("victory banner missing:\n%s", out)
	}
	if strings.Contains(out, "draw card") {
		t.Errorf("draw prompt should be unreachable:\n%s", out)
	}

	// space must not draw
	before := d.doneCount()
	v, _ = v.registerTap()
	v, _ = v.registerTap()
	v, _ = v.registerTap()
	if d.doneCount() != before {
		t.Error("drawing should be locked out while a winner exists")
	}
}

// esc must leave outright on the victory screen; arming an invisible confirm
// there swallows the next keypress
func TestVictoryEscLeavesWithoutConfirm(t *testing.T) {
	cfg.setSize(100, 40)
	d := deckOf(t, 2)
	tr, _ := makeTournament(d)
	tr.phase = phaseChallenge
	tr = pick(tr, 0)

	if tr.phase != phaseVictory {
		t.Fatal("expected victory")
	}
	tr.flash.active = false

	tr, cmd := tr.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
	if tr.confirm.active() {
		t.Error("esc should not arm a confirm on the victory screen")
	}
	if cmd == nil {
		t.Fatal("esc should emit a leave command")
	}
	if _, ok := cmd().(leaveTournamentMsg); !ok {
		t.Error("esc should leave to the list")
	}
}

// ctrl+r must actually reset, and its prompt must be visible
func TestVictoryResetPromptsAndClears(t *testing.T) {
	cfg.setSize(100, 40)
	d := deckOf(t, 2)
	tr, _ := makeTournament(d)
	tr.phase = phaseChallenge
	tr = pick(tr, 0)
	tr.flash.active = false

	tr, _ = tr.Update(tea.KeyPressMsg{Code: 'r', Mod: tea.ModCtrl})
	if !tr.confirm.active() {
		t.Fatal("ctrl+r should prompt")
	}
	if out := tr.View(); !strings.Contains(out, "reset this deck") {
		t.Errorf("reset prompt not rendered on victory screen:\n%s", out)
	}

	tr, cmd := tr.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
	if cmd == nil {
		t.Fatal("confirming should emit a command")
	}
	if _, ok := cmd().(leaveTournamentMsg); !ok {
		t.Error("reset should return to the list")
	}

	reloaded, _ := loadDeck(d.src)
	if reloaded.winner != "" {
		t.Errorf("winner should be cleared, got %q", reloaded.winner)
	}
}

func TestListRejectsTinyDeck(t *testing.T) {
	cfg.setSize(100, 40)
	small := writeTemp(t, "# Tiny\n\n- [ ] only one\n")
	reg := testRegistry(t, small)
	l := makeListView(reg, "")

	l, _ = l.startTournament()
	if l.notice == "" {
		t.Error("expected a rejection notice for a 1-card deck")
	}
}

func TestListPromptsBeforeResetting(t *testing.T) {
	cfg.setSize(100, 40)
	path := writeTemp(t, sample)
	reg := testRegistry(t, path)
	l := makeListView(reg, "")

	// sample has a checked card, so this must confirm first
	l, cmd := l.startTournament()
	if !l.confirm.active() || l.confirm.kind != confirmStartTournament {
		t.Error("expected a reset confirmation")
	}
	if cmd != nil {
		t.Error("should not start until confirmed")
	}
}
