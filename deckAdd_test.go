package main

import (
	"os"
	"path/filepath"
	"testing"
)

func deckDir(t *testing.T, names ...string) string {
	t.Helper()
	dir := t.TempDir()
	for _, n := range names {
		if err := os.WriteFile(filepath.Join(dir, n), []byte("# "+n+"\n\n- [ ] a\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// a directory yields its markdown files, skipping hidden and non-markdown
func TestExpandDeckArgDirectory(t *testing.T) {
	dir := deckDir(t, "a.md", "b.md", ".hidden.md", "notes.txt")
	if err := os.Mkdir(filepath.Join(dir, "nested"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "nested", "deep.md"), []byte("# deep\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	paths, err := expandDeckArg(dir)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{filepath.Join(dir, "a.md"), filepath.Join(dir, "b.md")}
	if len(paths) != len(want) {
		t.Fatalf("got %v, want %v", paths, want)
	}
	for i := range want {
		if paths[i] != want[i] {
			t.Errorf("got %v, want %v", paths, want)
		}
	}
}

func TestExpandDeckArgEmptyDirectory(t *testing.T) {
	if _, err := expandDeckArg(t.TempDir()); err == nil {
		t.Error("expected rejection for a directory with no decks")
	}
}

func TestExpandDeckArgFile(t *testing.T) {
	path := writeTemp(t, sample)
	paths, err := expandDeckArg(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0] != path {
		t.Errorf("got %v, want [%s]", paths, path)
	}
}

// re-adding is a no-op, and the first path is returned for focus either way
func TestAddDecksSkipsRegistered(t *testing.T) {
	dir := deckDir(t, "a.md", "b.md")
	paths, err := expandDeckArg(dir)
	if err != nil {
		t.Fatal(err)
	}

	reg := testRegistry(t)
	focus, added, err := addDecks(reg, paths)
	if err != nil {
		t.Fatal(err)
	}
	if added != 2 || focus != paths[0] {
		t.Fatalf("first add: focus %q added %d", focus, added)
	}

	focus, added, err = addDecks(reg, paths)
	if err != nil {
		t.Fatal(err)
	}
	if added != 0 || focus != paths[0] {
		t.Fatalf("second add: focus %q added %d", focus, added)
	}
	if len(reg.decks) != 2 {
		t.Errorf("registry has %d decks, want 2", len(reg.decks))
	}
}
