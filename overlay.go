package main

import "charm.land/lipgloss/v2"

// @region ui:overlay -- MODAL COMPOSITING
// composites fg over bg: horizontally centered on the app box, vertically
// centered in the terminal. The background is padded to full height first so
// the modal floats free of whatever short box happens to be underneath it.
func overlayCenter(fg, bg string) string {
	boxW := lipgloss.Width(bg)

	if cfg.ww > 0 && cfg.wh > lipgloss.Height(bg) {
		bg = lipgloss.Place(cfg.ww, cfg.wh, lipgloss.Left, lipgloss.Top, bg)
	}

	fgW, fgH := lipgloss.Width(fg), lipgloss.Height(fg)
	x := max((boxW-fgW)/2, 0)
	y := max((lipgloss.Height(bg)-fgH)/2, 0)

	return lipgloss.NewCompositor(
		lipgloss.NewLayer(bg),
		lipgloss.NewLayer(fg).X(x).Y(y).Z(1),
	).Render()
}
