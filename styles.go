package main

import "charm.land/lipgloss/v2"

// @region ui:styles -- COLORS AND STYLES
var (
	// Colors
	ColorPrimary   = lipgloss.Color("205")
	ColorSelection = lipgloss.Color("215")
	ColorError     = lipgloss.Color("197")
	ColorMuted     = lipgloss.Color("240")
	ColorDimBright = lipgloss.Color("248")
	ColorBasic     = lipgloss.Color("250")
	ColorActive    = lipgloss.Color("76")
	ColorCardBack  = lipgloss.Color("60")

	// Styles

	// both views are borderless; content carries the structure on its own
	ViewStyle = lipgloss.NewStyle().
			MarginTop(1).
			PaddingLeft(2).
			PaddingRight(2).
			MarginBottom(1)

	LabelStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	CardStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(ColorSelection)

	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	CardTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSelection)

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

	DimStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	CountStyle = lipgloss.NewStyle().Foreground(ColorMuted)

	ModalStyle = lipgloss.NewStyle().
			Padding(1, 2).
			Border(lipgloss.RoundedBorder(), true).
			BorderForeground(ColorPrimary)
)
