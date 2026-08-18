package main

import (
	"fmt"
	"os"

	tea "charm.land/bubbletea/v2"
)

// @region cli:entry -- ENTRYPOINT
func main() {
	paths, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	reg, err := loadRegistry()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// `add` is idempotent: already-registered decks are simply focused
	focus, _, err := addDecks(reg, paths)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
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
// returns the deck paths to register, if `add` was used; the first is focused
func parseArgs(args []string) ([]string, error) {
	if len(args) == 0 {
		return nil, nil
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
			return nil, fmt.Errorf("usage: drawdeck add {markdown file|directory}")
		}
		return expandDeckArg(args[1])
	}

	return nil, fmt.Errorf("unknown command: %s", args[0])
}
