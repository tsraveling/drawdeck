package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
)

func testRegistry(t *testing.T, deckPaths ...string) *registry {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	r, err := loadRegistry()
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range deckPaths {
		if err := r.add(p); err != nil {
			t.Fatal(err)
		}
	}
	return r
}

func sizedRoot(t *testing.T, reg *registry) rootModel {
	t.Helper()
	m := makeRootModel(reg, "")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 100, Height: 40})
	return updated.(rootModel)
}

// every view must render without panicking at a realistic size
func TestViewsRender(t *testing.T) {
	deckPath := writeTemp(t, sample)
	reg := testRegistry(t, deckPath)
	m := sizedRoot(t, reg)

	out := m.View().Content
	if !strings.Contains(out, "Example Deck") {
		t.Errorf("list view missing deck title:\n%s", out)
	}
	if !strings.Contains(out, "(1 / 5)") {
		t.Errorf("list view missing counts:\n%s", out)
	}

	// open the deck
	updated, cmd := m.Update(openDeckMsg{path: deckPath})
	m = updated.(rootModel)
	_ = cmd
	if m.mode != modeDetail {
		t.Fatal("did not switch to detail mode")
	}
	if out := m.View().Content; !strings.Contains(out, "draw card") {
		t.Errorf("detail view missing draw prompt:\n%s", out)
	}
}

func TestEmptyRegistryRenders(t *testing.T) {
	reg := testRegistry(t)
	m := sizedRoot(t, reg)
	if out := m.View().Content; !strings.Contains(out, "No decks yet") {
		t.Errorf("missing empty state:\n%s", out)
	}
}

func TestMissingDeckFileRenders(t *testing.T) {
	gone := filepath.Join(t.TempDir(), "gone.md")
	reg := testRegistry(t, gone)
	m := sizedRoot(t, reg)
	if out := m.View().Content; !strings.Contains(out, "(missing)") {
		t.Errorf("missing broken-deck marker:\n%s", out)
	}
}

func TestAddPromptValidation(t *testing.T) {
	deckPath := writeTemp(t, sample)
	reg := testRegistry(t, deckPath)
	l := makeListView(reg, "")

	if err := l.validateAdd(deckPath); err == nil {
		t.Error("expected duplicate rejection")
	}
	if err := l.validateAdd("/tmp/whatever.txt"); err == nil {
		t.Error("expected non-markdown rejection")
	}
	if err := l.validateAdd(""); err == nil {
		t.Error("expected empty rejection")
	}
	fresh := writeTemp(t, "# Fresh\n\n- [ ] a\n")
	if err := l.validateAdd(fresh); err != nil {
		t.Errorf("valid path rejected: %v", err)
	}
}

// three taps inside the window must draw and commit to disk
func TestTapGestureDraws(t *testing.T) {
	cfg.holdSupported = false
	defer func() { cfg.holdSupported = false }()

	path := writeTemp(t, sample)
	d, _ := loadDeck(path)
	before := d.doneCount()

	v := makeDetailView(d)
	for range 3 {
		v, _ = v.registerTap()
	}

	if !v.flip.active {
		t.Error("draw should have started the flip animation")
	}

	reloaded, _ := loadDeck(path)
	if reloaded.doneCount() != before+1 {
		t.Errorf("doneCount %d -> %d, want +1", before, reloaded.doneCount())
	}
	if reloaded.current == "" {
		t.Error("current not written to frontmatter")
	}
}

// the window is hard: taps that straddle it must not accumulate
func TestTapWindowExpires(t *testing.T) {
	cfg.holdSupported = false
	d, _ := loadDeck(writeTemp(t, sample))
	v := makeDetailView(d)

	v, _ = v.registerTap()
	v.tapStart = time.Now().Add(-2 * tapWindow)
	v, _ = v.tickCharge()

	if v.taps != 0 || v.tapPct != 0 {
		t.Errorf("expired window left taps=%d pct=%v", v.taps, v.tapPct)
	}
}

// a completed hold must not also register as a tap on its release
func TestHoldReleaseDoesNotDoubleCount(t *testing.T) {
	cfg.holdSupported = true
	defer func() { cfg.holdSupported = false }()

	d, _ := loadDeck(writeTemp(t, sample))
	v := makeDetailView(d)

	v, _ = v.handleKey(tea.KeyPressMsg{Code: ' ', Text: " "})
	if !v.holding {
		t.Fatal("space press should begin a hold")
	}

	v.holdStart = time.Now().Add(-2 * chargeDuration)
	v, _ = v.tickCharge()
	if !v.flip.active {
		t.Fatal("full hold should draw")
	}

	v, _ = v.Update(tea.KeyReleaseMsg{Code: ' ', Text: " "})
	if v.taps != 0 {
		t.Errorf("hold release counted as a tap: taps=%d", v.taps)
	}
}

// an aborted hold falls through to the tap counter so both gestures coexist
func TestAbortedHoldCountsAsTap(t *testing.T) {
	cfg.holdSupported = true
	defer func() { cfg.holdSupported = false }()

	d, _ := loadDeck(writeTemp(t, sample))
	v := makeDetailView(d)

	v, _ = v.handleKey(tea.KeyPressMsg{Code: ' ', Text: " "})
	v, _ = v.Update(tea.KeyReleaseMsg{Code: ' ', Text: " "})

	if v.taps != 1 {
		t.Errorf("taps = %d, want 1", v.taps)
	}
	if v.holding {
		t.Error("hold should have ended")
	}
}

func TestExhaustedDeck(t *testing.T) {
	path := writeTemp(t, "# Done\n\n- [x] a\n- [x] b\n")
	d, _ := loadDeck(path)
	v := makeDetailView(d)

	if !d.exhausted() {
		t.Fatal("deck should be exhausted")
	}
	if out := v.View(); !strings.Contains(out, "Deck exhausted") {
		t.Errorf("missing exhausted message:\n%s", out)
	}

	v, _ = v.registerTap()
	v, _ = v.registerTap()
	v, _ = v.registerTap()
	if v.flip.active {
		t.Error("exhausted deck should not draw")
	}
}

func TestResetFromDetail(t *testing.T) {
	path := writeTemp(t, sample)
	d, _ := loadDeck(path)
	v := makeDetailView(d)
	v.reset()

	raw, _ := os.ReadFile(path)
	if strings.Contains(string(raw), "[x]") {
		t.Errorf("reset left checked boxes:\n%s", raw)
	}
}

func TestFlipRendersAtAllFrames(t *testing.T) {
	cfg.setSize(100, 40)
	f := flip{active: true}
	for i := range flipFrames {
		f.frame = i
		if out := f.view(60, 6); out == "" {
			t.Errorf("frame %d rendered empty", i)
		}
	}
}

func TestHelpRenders(t *testing.T) {
	out := renderHelp(helpSource, 70)
	if !strings.Contains(out, "drawdeck") {
		t.Error("help missing title")
	}
	if strings.Contains(out, "```") {
		t.Error("code fences should be stripped")
	}
}
