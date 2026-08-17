package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"slices"
)

const registryFile = "decks.json"

// on-disk shape: paths only, never cached titles
type registryData struct {
	Decks []string `json:"decks"`
}

type registry struct {
	path  string
	decks []string
}

func registryPath() (string, error) {
	dir, err := ensureConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, registryFile), nil
}

func loadRegistry() (*registry, error) {
	path, err := registryPath()
	if err != nil {
		return nil, err
	}

	r := &registry{path: path}

	raw, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return r, nil
	}
	if err != nil {
		return nil, err
	}

	var data registryData
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	r.decks = data.Decks
	return r, nil
}

func (r *registry) save() error {
	raw, err := json.MarshalIndent(registryData{Decks: r.decks}, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.path, append(raw, '\n'), 0o644)
}

func (r *registry) has(abs string) bool {
	return slices.Contains(r.decks, abs)
}

func (r *registry) add(abs string) error {
	if r.has(abs) {
		return nil
	}
	r.decks = append(r.decks, abs)
	return r.save()
}

func (r *registry) remove(abs string) error {
	idx := slices.Index(r.decks, abs)
	if idx < 0 {
		return nil
	}
	r.decks = slices.Delete(r.decks, idx, idx+1)
	return r.save()
}
