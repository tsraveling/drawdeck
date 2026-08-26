package main

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// @region ui:wrap -- TEXT WRAP HELPER
func wrapText(s string, width int) string {
	if width < 1 {
		return s
	}
	var out []string
	for para := range strings.SplitSeq(s, "\n") {
		line := ""
		for word := range strings.FieldsSeq(para) {
			switch {
			case line == "":
				line = word
			case len(line)+1+len(word) <= width:
				line += " " + word
			default:
				out = append(out, line)
				line = word
			}
		}
		out = append(out, line)
	}
	return strings.Join(out, "\n")
}

// @region ui:truncate -- WIDTH-LIMITED TRUNCATION
// cuts a string down to at most w display columns
func truncateWidth(s string, w int) string {
	if lipgloss.Width(s) <= w {
		return s
	}
	var b strings.Builder
	for _, r := range s {
		if lipgloss.Width(b.String()+string(r)) > w {
			break
		}
		b.WriteRune(r)
	}
	return b.String()
}
