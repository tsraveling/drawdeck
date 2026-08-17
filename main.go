package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

func printHelp() {
	fmt.Println("drawdeck")
	fmt.Println()
	fmt.Println("Usage: drawdeck [options]")
	fmt.Println()
	fmt.Println("  -h, --help     show this help")
	fmt.Println("  -v, --version  show version")
}

func main() {
	for _, arg := range os.Args[1:] {
		switch arg {
		case "-h", "--help":
			printHelp()
			return
		case "-v", "--version":
			fmt.Println(Version)
			return
		default:
			fmt.Fprintf(os.Stderr, "unknown argument: %s\n", arg)
			os.Exit(1)
		}
	}

	var m tea.Model
	m, _ = makeMainViewModel()

	p := tea.NewProgram(m, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
