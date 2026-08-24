package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// non-indented checkbox line: marker, state, title
var cardRe = regexp.MustCompile(`^([-*+])(\s+)\[([ xX])\]\s?(.*)$`)

// indented quote line, any depth
var noteRe = regexp.MustCompile(`^\s+>\s?(.*)$`)

// top-level markdown header
var titleRe = regexp.MustCompile(`^#\s+(.*)$`)

// priority marker anywhere in a card title
var prioRe = regexp.MustCompile(`\((!{1,2})\)`)

// @region deck:model -- DECK AND CARD TYPES
// draw tiers, highest first
const (
	prioNormal = iota
	prioMedium
	prioTop
)

type card struct {
	title    string
	notes    string
	checked  bool
	priority int

	// index into deck.lines, for surgical rewrites
	line int
}

type deck struct {
	title   string
	cards   []card
	src     string
	current string
	winner  string

	// raw file lines; every write mutates these in place and rewrites
	lines []string

	// frontmatter block bounds (inclusive of the --- fences), -1 when absent
	fmStart int
	fmEnd   int
}

// @region deck:state -- DRAW POOL AND PROGRESS
func (d *deck) doneCount() int {
	n := 0
	for _, c := range d.cards {
		if c.checked {
			n++
		}
	}
	return n
}

// the current card, or nil when there is none or it no longer exists
func (d *deck) currentCard() *card {
	if d.current == "" {
		return nil
	}
	for i := range d.cards {
		if d.cards[i].title == d.current {
			return &d.cards[i]
		}
	}
	return nil
}

// unchecked cards. unless --no-priority is set, the pool narrows to the
// highest priority tier that still has candidates
func (d *deck) drawable() []int {
	var out []int
	best := prioNormal
	for i, c := range d.cards {
		if c.checked {
			continue
		}
		if cfg.noPriority {
			out = append(out, i)
			continue
		}
		if c.priority > best {
			best = c.priority
			out = out[:0]
		}
		if c.priority == best {
			out = append(out, i)
		}
	}
	return out
}

// (!!) is top priority, (!) is medium, anything else is normal
func parsePriority(title string) int {
	m := prioRe.FindStringSubmatch(title)
	if m == nil {
		return prioNormal
	}
	if m[1] == "!!" {
		return prioTop
	}
	return prioMedium
}

func (d *deck) exhausted() bool {
	return len(d.drawable()) == 0
}

// @region deck:load -- LOAD FROM MARKDOWN
func loadDeck(path string) (*deck, error) {
	if !strings.EqualFold(filepath.Ext(path), ".md") {
		return nil, fmt.Errorf("not a markdown file: %s", filepath.Base(path))
	}

	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	d := &deck{src: path, fmStart: -1, fmEnd: -1}
	d.lines = strings.Split(string(raw), "\n")
	d.parse()
	return d, nil
}

// @region deck:parse -- PARSE CARDS AND FRONTMATTER
// rebuilds title, cards, and frontmatter state from d.lines
func (d *deck) parse() {
	d.cards = nil
	d.current = ""
	d.winner = ""
	d.title = ""
	d.fmStart, d.fmEnd = -1, -1

	body := 0

	// frontmatter only counts when it opens on the very first line
	if len(d.lines) > 0 && strings.TrimRight(d.lines[0], "\r") == "---" {
		for i := 1; i < len(d.lines); i++ {
			if strings.TrimRight(d.lines[i], "\r") == "---" {
				d.fmStart, d.fmEnd = 0, i
				body = i + 1
				break
			}
		}
	}

	if d.fmStart >= 0 {
		for i := d.fmStart + 1; i < d.fmEnd; i++ {
			k, v, ok := strings.Cut(d.lines[i], ":")
			if !ok {
				continue
			}
			switch strings.TrimSpace(k) {
			case "current":
				d.current = unquoteValue(v)
			case "winner":
				d.winner = unquoteValue(v)
			}
		}
	}

	for i := body; i < len(d.lines); i++ {
		line := strings.TrimRight(d.lines[i], "\r")

		if d.title == "" {
			if m := titleRe.FindStringSubmatch(line); m != nil {
				d.title = strings.TrimSpace(m[1])
				continue
			}
		}

		m := cardRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}

		c := card{
			title:   strings.TrimSpace(m[4]),
			checked: m[3] != " ",
			line:    i,
		}
		c.priority = parsePriority(c.title)
		c.notes = collectNotes(d.lines, i+1)
		d.cards = append(d.cards, c)
	}

	// no header: fall back to the filename
	if d.title == "" {
		base := filepath.Base(d.src)
		d.title = strings.TrimSuffix(base, filepath.Ext(base))
	}
}

// @region deck:notes -- INDENTED QUOTE NOTES
// gathers consecutive indented quote lines following a card, terminated
// by a blank line or any non-quote content
func collectNotes(lines []string, start int) string {
	var parts []string
	for i := start; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if strings.TrimSpace(line) == "" {
			break
		}
		m := noteRe.FindStringSubmatch(line)
		if m == nil {
			break
		}
		parts = append(parts, strings.TrimSpace(m[1]))
	}
	return strings.Join(parts, " ")
}

// reads a frontmatter scalar, accepting both quoted and bare forms
func unquoteValue(v string) string {
	v = strings.TrimSpace(strings.TrimRight(v, "\r"))
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		inner := v[1 : len(v)-1]
		inner = strings.ReplaceAll(inner, `\"`, `"`)
		return strings.ReplaceAll(inner, `\\`, `\`)
	}
	return v
}

// always emitted double-quoted so titles containing ':' or a leading '-'
// cannot corrupt the block
func quoteValue(v string) string {
	esc := strings.ReplaceAll(v, `\`, `\\`)
	esc = strings.ReplaceAll(esc, `"`, `\"`)
	return `"` + esc + `"`
}
