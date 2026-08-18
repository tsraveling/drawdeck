package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const sample = `# Example Deck

Some prose that must survive untouched.

- [ ] A woolly yak
- [ ] A giant elephant
    > This one is my favorite
    > and has a second line
- [x] An adorable baby hippo
* [ ] Star marker card
    - [ ] indented, must be ignored

## Another header

+ [ ] Plus marker card
`

func writeTemp(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "deck.md")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestParse(t *testing.T) {
	d, err := loadDeck(writeTemp(t, sample))
	if err != nil {
		t.Fatal(err)
	}

	if d.title != "Example Deck" {
		t.Errorf("title = %q", d.title)
	}

	want := []string{
		"A woolly yak", "A giant elephant", "An adorable baby hippo",
		"Star marker card", "Plus marker card",
	}
	if len(d.cards) != len(want) {
		t.Fatalf("got %d cards, want %d: %+v", len(d.cards), len(want), d.cards)
	}
	for i, w := range want {
		if d.cards[i].title != w {
			t.Errorf("card %d = %q, want %q", i, d.cards[i].title, w)
		}
	}

	if got := d.cards[1].notes; got != "This one is my favorite and has a second line" {
		t.Errorf("notes = %q", got)
	}
	if !d.cards[2].checked {
		t.Error("hippo should be checked")
	}
	if d.doneCount() != 1 {
		t.Errorf("doneCount = %d", d.doneCount())
	}
}

func TestNoHeaderFallsBackToFilename(t *testing.T) {
	d, err := loadDeck(writeTemp(t, "- [ ] lonely card\n"))
	if err != nil {
		t.Fatal(err)
	}
	if d.title != "deck" {
		t.Errorf("title = %q, want %q", d.title, "deck")
	}
}

func TestRejectsNonMarkdown(t *testing.T) {
	if _, err := loadDeck("/tmp/nope.txt"); err == nil {
		t.Error("expected error for non-.md path")
	}
}

// the load-bearing guarantee: everything except the frontmatter block and
// the one checked line is byte-identical
func TestSurgicalWrite(t *testing.T) {
	path := writeTemp(t, sample)
	d, _ := loadDeck(path)

	idx := d.drawable()[0]
	d.setLineChecked(d.cards[idx].line, true)
	d.setCurrent(d.cards[idx].title)
	if err := d.write(); err != nil {
		t.Fatal(err)
	}

	raw, _ := os.ReadFile(path)
	got := string(raw)

	if !strings.HasPrefix(got, "---\ncurrent: \"A woolly yak\"\n---\n\n") {
		t.Fatalf("frontmatter not inserted at top:\n%s", got)
	}
	if !strings.Contains(got, "- [x] A woolly yak") {
		t.Error("card not checked off")
	}
	if !strings.Contains(got, "Some prose that must survive untouched.") {
		t.Error("prose lost")
	}
	if !strings.Contains(got, "## Another header") {
		t.Error("second header lost")
	}
	if !strings.Contains(got, "    - [ ] indented, must be ignored") {
		t.Error("indented checkbox mutated")
	}
	if strings.Contains(got, "- [x] A giant elephant") {
		t.Error("wrong card checked")
	}

	// reload and confirm round-trip
	d2, _ := loadDeck(path)
	if d2.current != "A woolly yak" {
		t.Errorf("current = %q", d2.current)
	}
	if len(d2.cards) != 5 {
		t.Errorf("card count changed to %d", len(d2.cards))
	}
}

func TestTitlesWithSpecialCharsRoundTrip(t *testing.T) {
	for _, title := range []string{`Ship it: phase 2`, `- leading dash`, `has "quotes"`, `back\slash`} {
		d, _ := loadDeck(writeTemp(t, "# T\n\n- [ ] "+title+"\n"))
		d.setCurrent(title)
		if err := d.write(); err != nil {
			t.Fatal(err)
		}
		d2, _ := loadDeck(d.src)
		if d2.current != title {
			t.Errorf("round-trip %q -> %q", title, d2.current)
		}
	}
}

func TestClearCurrentRemovesEmptyBlock(t *testing.T) {
	path := writeTemp(t, sample)
	d, _ := loadDeck(path)
	d.setCurrent("A woolly yak")
	d.write()

	d2, _ := loadDeck(path)
	d2.clearCurrent()
	d2.write()

	raw, _ := os.ReadFile(path)
	if strings.HasPrefix(string(raw), "---") {
		t.Errorf("empty frontmatter block left behind:\n%s", raw)
	}
	if !strings.HasPrefix(string(raw), "# Example Deck") {
		t.Errorf("file should start with header again:\n%s", raw)
	}
}

func TestResetUnchecksEverything(t *testing.T) {
	path := writeTemp(t, sample)
	d, _ := loadDeck(path)
	d.setCurrent("A woolly yak")
	d.resetAll()
	d.write()

	d2, _ := loadDeck(path)
	if d2.doneCount() != 0 {
		t.Errorf("doneCount after reset = %d", d2.doneCount())
	}
	if d2.current != "" {
		t.Errorf("current after reset = %q", d2.current)
	}
	if len(d2.cards) != 5 {
		t.Errorf("card count = %d", len(d2.cards))
	}
}

// both keys must coexist, and clearing one must not drop the block
func TestFrontmatterKeysCoexist(t *testing.T) {
	path := writeTemp(t, sample)
	d, _ := loadDeck(path)
	d.setCurrent("A woolly yak")
	d.setWinner("A giant elephant")
	d.write()

	d2, _ := loadDeck(path)
	if d2.current != "A woolly yak" || d2.winner != "A giant elephant" {
		t.Fatalf("round-trip failed: current=%q winner=%q", d2.current, d2.winner)
	}

	d2.clearCurrent()
	d2.write()

	d3, _ := loadDeck(path)
	if d3.current != "" {
		t.Errorf("current should be gone, got %q", d3.current)
	}
	if d3.winner != "A giant elephant" {
		t.Errorf("winner should survive, got %q", d3.winner)
	}

	d3.clearWinner()
	d3.write()

	raw, _ := os.ReadFile(path)
	if strings.HasPrefix(string(raw), "---") {
		t.Errorf("block should be gone once empty:\n%s", raw)
	}
}

func TestResetClearsWinner(t *testing.T) {
	path := writeTemp(t, sample)
	d, _ := loadDeck(path)
	d.setWinner("A woolly yak")
	d.setCurrent("A giant elephant")
	d.resetAll()
	d.write()

	d2, _ := loadDeck(path)
	if d2.winner != "" || d2.current != "" || d2.doneCount() != 0 {
		t.Errorf("reset incomplete: winner=%q current=%q done=%d",
			d2.winner, d2.current, d2.doneCount())
	}
}

func TestDrawableExcludesCurrent(t *testing.T) {
	d, _ := loadDeck(writeTemp(t, "# T\n\n- [ ] a\n- [ ] b\n"))
	d.current = "a"
	got := d.drawable()
	if len(got) != 1 || d.cards[got[0]].title != "b" {
		t.Errorf("drawable = %+v, want just b", got)
	}
}
