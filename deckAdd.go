package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// @region deck:add -- DECK PATH EXPANSION

// resolves an add argument to deck paths: a markdown file yields itself,
// a directory yields every readable .md directly inside it
func expandDeckArg(in string) ([]string, error) {
	if strings.TrimSpace(in) == "" {
		return nil, fmt.Errorf("enter a path")
	}

	abs, err := resolvePath(in)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(abs)
	if err != nil {
		return nil, fmt.Errorf("cannot read: %s", filepath.Base(abs))
	}

	if !info.IsDir() {
		if !strings.EqualFold(filepath.Ext(abs), ".md") {
			return nil, fmt.Errorf("not a markdown file: %s", filepath.Base(abs))
		}
		if _, err := loadDeck(abs); err != nil {
			return nil, err
		}
		return []string{abs}, nil
	}

	return markdownFilesIn(abs)
}

// not recursive; hidden files and unreadable decks are skipped
func markdownFilesIn(dir string) ([]string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("cannot read: %s", filepath.Base(dir))
	}

	var out []string
	for _, e := range entries {
		name := e.Name()
		if e.IsDir() || strings.HasPrefix(name, ".") {
			continue
		}
		if !strings.EqualFold(filepath.Ext(name), ".md") {
			continue
		}
		path := filepath.Join(dir, name)
		if _, err := loadDeck(path); err != nil {
			continue
		}
		out = append(out, path)
	}

	if len(out) == 0 {
		return nil, fmt.Errorf("no decks in %s", filepath.Base(dir))
	}
	return out, nil
}

// @region deck:register -- BULK REGISTRATION

// registers every path not already present; returns the first path, for
// focus, and how many were new
func addDecks(reg *registry, paths []string) (string, int, error) {
	added := 0
	for _, p := range paths {
		if reg.has(p) {
			continue
		}
		if err := reg.add(p); err != nil {
			return "", added, err
		}
		added++
	}
	if len(paths) == 0 {
		return "", 0, nil
	}
	return paths[0], added, nil
}
