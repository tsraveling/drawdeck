package main

import "strings"

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
