package main

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// emitted when the user opens a deck
type openDeckMsg struct{ path string }

type deckEntry struct {
	path string
	deck *deck
	err  error
}

// deck title when readable, filename otherwise
func (e deckEntry) sortKey() string {
	if e.deck != nil {
		return strings.ToLower(e.deck.title)
	}
	base := filepath.Base(e.path)
	return strings.ToLower(strings.TrimSuffix(base, filepath.Ext(base)))
}

type listView struct {
	reg     *registry
	entries []deckEntry
	cursor  int
	prompt  *textPrompt
	confirm confirm
}

func makeListView(reg *registry, focus string) listView {
	l := listView{reg: reg}
	l.reload()
	l.focusOn(focus)
	return l
}

// re-reads every registered deck file; titles and counts are never cached
func (l *listView) reload() {
	prev := l.selectedPath()

	l.entries = nil
	for _, p := range l.reg.decks {
		e := deckEntry{path: p}
		e.deck, e.err = loadDeck(p)
		l.entries = append(l.entries, e)
	}
	sort.Slice(l.entries, func(a, b int) bool {
		return l.entries[a].sortKey() < l.entries[b].sortKey()
	})

	if prev != "" {
		l.focusOn(prev)
	}
	l.clampCursor()
}

func (l *listView) focusOn(path string) {
	if path == "" {
		return
	}
	for i, e := range l.entries {
		if e.path == path {
			l.cursor = i
			return
		}
	}
}

func (l *listView) clampCursor() {
	if l.cursor >= len(l.entries) {
		l.cursor = len(l.entries) - 1
	}
	if l.cursor < 0 {
		l.cursor = 0
	}
}

func (l listView) selectedPath() string {
	if l.cursor < 0 || l.cursor >= len(l.entries) {
		return ""
	}
	return l.entries[l.cursor].path
}

// rejects non-markdown, missing, unreadable, and already-registered paths
func (l *listView) validateAdd(in string) error {
	if strings.TrimSpace(in) == "" {
		return fmt.Errorf("enter a path")
	}
	abs, err := resolvePath(in)
	if err != nil {
		return err
	}
	if !strings.EqualFold(filepath.Ext(abs), ".md") {
		return fmt.Errorf("must be a .md file")
	}
	if l.reg.has(abs) {
		return fmt.Errorf("already added")
	}
	if _, err := loadDeck(abs); err != nil {
		return fmt.Errorf("cannot read: %s", filepath.Base(abs))
	}
	return nil
}

func (l listView) Update(msg tea.Msg) (listView, tea.Cmd) {
	// modal and confirm swallow input while active
	if l.prompt != nil {
		p, cmd, res := l.prompt.Update(msg)
		l.prompt = &p
		switch res {
		case promptCancel:
			l.prompt = nil
		case promptCommit:
			abs, err := resolvePath(p.value())
			l.prompt = nil
			if err == nil {
				l.reg.add(abs)
				l.reload()
				l.focusOn(abs)
			}
		}
		return l, cmd
	}

	if l.confirm.active() {
		switch l.confirm.handle(msg) {
		case confirmYes:
			l.reg.remove(l.selectedPath())
			l.confirm = confirm{}
			l.reload()
		case confirmNo:
			l.confirm = confirm{}
		}
		return l, nil
	}

	km, ok := msg.(tea.KeyPressMsg)
	if !ok {
		return l, nil
	}

	switch km.String() {
	case "q", "ctrl+c":
		return l, tea.Quit
	case "up", "k":
		if l.cursor > 0 {
			l.cursor--
		}
	case "down", "j":
		if l.cursor < len(l.entries)-1 {
			l.cursor++
		}
	case "enter", "l", "right":
		if e := l.selected(); e != nil && e.deck != nil {
			path := e.path
			return l, func() tea.Msg { return openDeckMsg{path} }
		}
	case "a":
		p := newTextPrompt("Add deck", "path/to/deck.md", l.validateAdd)
		l.prompt = &p
	case "d":
		if l.selected() != nil {
			l.confirm = newConfirm(confirmDeleteDeck, "Delete?")
		}
	case "r":
		l.reload()
	}

	return l, nil
}

func (l listView) selected() *deckEntry {
	if l.cursor < 0 || l.cursor >= len(l.entries) {
		return nil
	}
	return &l.entries[l.cursor]
}

func (l listView) View() string {
	w := cfg.viewWidth()

	var b strings.Builder
	b.WriteString(TitleStyle.Render("DRAWDECK"))
	b.WriteString("\n\n")

	if len(l.entries) == 0 {
		empty := DimStyle.Render("No decks yet — press `a` to add one")
		b.WriteString(lipgloss.PlaceHorizontal(cfg.contentWidth(), lipgloss.Center, empty))
		b.WriteString("\n")
	}

	for i, e := range l.entries {
		selected := i == l.cursor

		var left, right string
		if e.deck != nil {
			left = e.deck.title
			right = fmt.Sprintf("(%d / %d)", e.deck.doneCount(), len(e.deck.cards))
		} else {
			left = e.path
			right = "(missing)"
		}

		cursor := "  "
		if selected {
			cursor = "› "
		}

		style := UnselectedItemStyle
		if selected {
			style = SelectedItemStyle
		}
		if e.deck == nil {
			style = ErrorStyle
		}

		// the delete confirm replaces the count on its own row
		rightRendered := CountStyle.Render(right)
		if selected && l.confirm.active() {
			rightRendered = l.confirm.View()
		}

		gap := max(cfg.contentWidth()-lipgloss.Width(cursor+left)-lipgloss.Width(rightRendered), 1)
		b.WriteString(cursor)
		b.WriteString(style.Render(left))
		b.WriteString(strings.Repeat(" ", gap))
		b.WriteString(rightRendered)
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(HelpStyle.Render("↑↓/jk navigate  enter open  a add  d delete  r refresh  q quit"))

	body := ViewStyle.Width(w).Render(b.String())
	if l.prompt != nil {
		return overlayCenter(l.prompt.View(), body)
	}
	return body
}
