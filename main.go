package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	tea "charm.land/bubbletea/v2"
)

// @region cli:entry -- ENTRYPOINT
func main() {
	focus, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	reg, err := loadRegistry()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// `add` is idempotent: an already-registered deck is simply focused
	if focus != "" && !reg.has(focus) {
		if err := reg.add(focus); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
	}

	run(makeRootModel(reg, focus))
}

func run(m tea.Model) {
	if _, err := tea.NewProgram(m).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// @region cli:args -- COMMAND LINE ARGS
// returns the absolute path of a deck to focus, if `add` was used
func parseArgs(args []string) (string, error) {
	if len(args) == 0 {
		return "", nil
	}

	switch args[0] {
	case "-v", "--version":
		fmt.Println(Version)
		os.Exit(0)
	case "-h", "--help":
		run(makeHelpModel())
		os.Exit(0)
	case "add":
		if len(args) < 2 {
			return "", fmt.Errorf("usage: drawdeck add {markdown file}")
		}
		return validateDeckArg(args[1])
	}

	return "", fmt.Errorf("unknown command: %s", args[0])
}

func validateDeckArg(in string) (string, error) {
	abs, err := resolvePath(in)
	if err != nil {
		return "", err
	}
	if !strings.EqualFold(filepath.Ext(abs), ".md") {
		return "", fmt.Errorf("not a markdown file: %s", filepath.Base(abs))
	}
	if _, err := loadDeck(abs); err != nil {
		return "", err
	}
	return abs, nil
}
