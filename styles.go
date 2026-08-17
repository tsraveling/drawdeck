package main

import "github.com/charmbracelet/lipgloss"

var (
	// Colors
	ColorPrimary   = lipgloss.Color("205")
	ColorSelection = lipgloss.Color("215")
	ColorError     = lipgloss.Color("197")
	ColorMuted     = lipgloss.Color("240")
	ColorDimBright = lipgloss.Color("248")
	ColorBasic     = lipgloss.Color("250")
	ColorActive    = lipgloss.Color("76")

	// Styles

	ViewStyle = lipgloss.NewStyle().
			MarginTop(1).
			PaddingTop(1).
			PaddingLeft(2).
			PaddingBottom(1).
			MarginBottom(1).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(ColorPrimary)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	ItemStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	SelectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(ColorSelection)

	UnselectedItemStyle = lipgloss.NewStyle().
				PaddingLeft(0).
				Foreground(ColorDimBright)

	ErrorStyle = lipgloss.NewStyle().Foreground(ColorError)

	ActiveStyle = lipgloss.NewStyle().Foreground(ColorActive)

	HelpStyle = lipgloss.NewStyle().Foreground(ColorMuted)
)
