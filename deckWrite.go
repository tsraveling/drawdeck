package main

import (
	"os"
	"strings"
)

// flips the state character of a checkbox line in place. Only the single
// byte between the brackets changes; marker, indent, and spacing survive.
func (d *deck) setLineChecked(idx int, checked bool) {
	if idx < 0 || idx >= len(d.lines) {
		return
	}
	line := d.lines[idx]
	open := strings.Index(line, "[")
	if open < 0 || open+2 >= len(line) || line[open+2] != ']' {
		return
	}
	state := byte(' ')
	if checked {
		state = 'x'
	}
	d.lines[idx] = line[:open+1] + string(state) + line[open+2:]
}

// writes a frontmatter scalar, creating the block at the top of the file
// when it does not yet exist
func (d *deck) setFrontmatterKey(key, value string) {
	entry := key + ": " + quoteValue(value)

	if d.fmStart < 0 {
		block := []string{"---", entry, "---", ""}
		d.lines = append(block, d.lines...)
		d.shiftCardLines(len(block))
		d.fmStart, d.fmEnd = 0, 2
		return
	}

	for i := d.fmStart + 1; i < d.fmEnd; i++ {
		if k, _, ok := strings.Cut(d.lines[i], ":"); ok && strings.TrimSpace(k) == key {
			d.lines[i] = entry
			return
		}
	}

	// block exists but has no such key yet
	d.lines = append(d.lines[:d.fmEnd], append([]string{entry}, d.lines[d.fmEnd:]...)...)
	d.fmEnd++
	d.shiftCardLinesFrom(d.fmEnd-1, 1)
}

// drops a frontmatter key; removes the whole block only if no other key remains
func (d *deck) clearFrontmatterKey(key string) {
	if d.fmStart < 0 {
		return
	}

	idx := -1
	others := 0
	for i := d.fmStart + 1; i < d.fmEnd; i++ {
		if strings.TrimSpace(d.lines[i]) == "" {
			continue
		}
		if k, _, ok := strings.Cut(d.lines[i], ":"); ok && strings.TrimSpace(k) == key {
			idx = i
			continue
		}
		others++
	}
	if idx < 0 {
		return
	}

	if others == 0 {
		// remove the fences too, plus one trailing blank line if we added one
		end := d.fmEnd + 1
		if end < len(d.lines) && strings.TrimSpace(d.lines[end]) == "" {
			end++
		}
		removed := end - d.fmStart
		d.lines = append(d.lines[:d.fmStart], d.lines[end:]...)
		d.shiftCardLines(-removed)
		d.fmStart, d.fmEnd = -1, -1
		return
	}

	d.lines = append(d.lines[:idx], d.lines[idx+1:]...)
	d.fmEnd--
	d.shiftCardLinesFrom(idx, -1)
}

func (d *deck) setCurrent(title string) {
	d.current = title
	d.setFrontmatterKey("current", title)
}

func (d *deck) clearCurrent() {
	d.current = ""
	d.clearFrontmatterKey("current")
}

func (d *deck) setWinner(title string) {
	d.winner = title
	d.setFrontmatterKey("winner", title)
}

func (d *deck) clearWinner() {
	d.winner = ""
	d.clearFrontmatterKey("winner")
}

// unchecks every card and forgets both the current card and any winner
func (d *deck) resetAll() {
	for i := range d.cards {
		d.setLineChecked(d.cards[i].line, false)
		d.cards[i].checked = false
	}
	d.clearCurrent()
	d.clearWinner()
}

func (d *deck) shiftCardLines(delta int) {
	for i := range d.cards {
		d.cards[i].line += delta
	}
}

func (d *deck) shiftCardLinesFrom(from, delta int) {
	for i := range d.cards {
		if d.cards[i].line >= from {
			d.cards[i].line += delta
		}
	}
}

func (d *deck) write() error {
	return os.WriteFile(d.src, []byte(strings.Join(d.lines, "\n")), 0o644)
}
