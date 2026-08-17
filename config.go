package main

import (
	"os"
	"path/filepath"
	"strings"
)

// clamp bounds for view width
const (
	minViewWidth = 30
	maxViewWidth = 80
)

type config struct {
	// terminal dimensions
	ww int
	wh int

	// true once the terminal reports Kitty keyboard protocol support
	holdSupported bool
}

var cfg config

// width used by both views, clamped to a comfortable reading range
func (c *config) viewWidth() int {
	return max(minViewWidth, min(c.ww, maxViewWidth))
}

// usable width inside ViewStyle's horizontal padding
func (c *config) contentWidth() int {
	return c.viewWidth() - 4
}

func (c *config) setSize(w, h int) {
	c.ww = w
	c.wh = h
}

// resolves $XDG_CONFIG_HOME/drawdeck, falling back to ~/.config/drawdeck
func configDir() (string, error) {
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "drawdeck"), nil
}

func ensureConfigDir() (string, error) {
	dir, err := configDir()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	return dir, nil
}

func expandPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "~" || strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			return path
		}
		if path == "~" {
			return home
		}
		return filepath.Join(home, path[2:])
	}
	return path
}

// expands ~, resolves relative to cwd, and cleans the result
func resolvePath(path string) (string, error) {
	return filepath.Abs(expandPath(path))
}
