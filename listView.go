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

// emitted when the user starts tournament mode on a deck
type startTournamentMsg struct{ path string }

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

// @region list:view -- DECK LIST MODEL
type listView struct {
	reg     *registry
	entries []deckEntry
	cursor  int
	prompt  *textPrompt
	confirm confirm

	// transient inline message, cleared on the next navigation
	notice string

	// renders the notice as success rather than as a warning
	noticeOK bool
}

func makeListView(reg *registry, focus string) listView {
	l := listView{reg: reg}
	l.reload()
	l.focusOn(focus)
	return l
}

// @region list:reload -- RELOAD DECKS FROM DISK
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

func (l *listView) setNotice(msg string, ok bool) {
	l.notice = msg
	l.noticeOK = ok
}

func (l *listView) clearNotice() {
	l.notice = ""
	l.noticeOK = false
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

// @region list:add -- ADD DECK VALIDATION
// rejects missing, unreadable, and already-registered paths; a directory
// passes as long as it holds at least one unregistered deck
func (l *listView) validateAdd(in string) error {
	paths, err := expandDeckArg(in)
	if err != nil {
		return err
	}
	for _, p := range paths {
		if !l.reg.has(p) {
			return nil
		}
	}
	return fmt.Errorf("already added")
}

// @region list:keys -- LIST INPUT
func (l listView) Update(msg tea.Msg) (listView, tea.Cmd) {
	// modal and confirm swallow input while active
	if l.prompt != nil {
		p, cmd, res := l.prompt.Update(msg)
		l.prompt = &p
		switch res {
		case promptCancel:
			l.prompt = nil
		case promptCommit:
			paths, err := expandDeckArg(p.value())
			l.prompt = nil
			if err == nil {
				focus, added, _ := addDecks(l.reg, paths)
				l.reload()
				l.focusOn(focus)
				if added > 1 {
					l.setNotice(fmt.Sprintf("added %d decks", added), true)
				}
			}
		}
		return l, cmd
	}

	if l.confirm.active() {
		kind := l.confirm.kind
		switch l.confirm.handle(msg) {
		case confirmYes:
			l.confirm = confirm{}
			if kind == confirmStartTournament {
				path := l.selectedPath()
				return l, func() tea.Msg { return startTournamentMsg{path} }
			}
			l.reg.remove(l.selectedPath())
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

	l.clearNotice()

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
	case "t":
		return l.startTournament()
	case "enter", "l", "right":
		if e := l.selected(); e != nil && e.deck != nil {
			path := e.path
			return l, func() tea.Msg { return openDeckMsg{path} }
		}
	case "a":
		p := newTextPrompt("Add deck", "deck.md or a folder", l.validateAdd)
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

// @region list:tourney -- TOURNAMENT ENTRY GUARD
// a tournament needs a readable deck of at least two cards, and warns before
// discarding any existing play
func (l listView) startTournament() (listView, tea.Cmd) {
	e := l.selected()
	if e == nil || e.deck == nil {
		return l, nil
	}
	if len(e.deck.cards) < 2 {
		l.setNotice("tournament needs at least 2 cards", false)
		return l, nil
	}

	if e.deck.doneCount() > 0 || e.deck.current != "" || e.deck.winner != "" {
		l.confirm = newConfirm(confirmStartTournament, "Tournament mode will reset deck. Continue?")
		return l, nil
	}

	path := e.path
	return l, func() tea.Msg { return startTournamentMsg{path} }
}

func (l listView) selected() *deckEntry {
	if l.cursor < 0 || l.cursor >= len(l.entries) {
		return nil
	}
	return &l.entries[l.cursor]
}

// @region list:render -- LIST RENDER
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

		// the short delete confirm replaces the count on its own row; longer
		// prompts render below the list instead
		rightRendered := CountStyle.Render(right)
		if selected && l.confirm.kind == confirmDeleteDeck {
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
	if l.confirm.kind == confirmStartTournament {
		b.WriteString(lipgloss.PlaceHorizontal(cfg.contentWidth(), lipgloss.Center, l.confirm.View()))
		b.WriteString("\n\n")
	}
	if l.notice != "" {
		style := ErrorStyle
		if l.noticeOK {
			style = ActiveStyle
		}
		b.WriteString(style.Render(l.notice))
		b.WriteString("\n\n")
	}
	b.WriteString(HelpStyle.Render("↑↓/jk navigate  enter open  t tournament  a add  d delete  r refresh  q quit"))

	body := ViewStyle.Width(w).Render(b.String())
	if l.prompt != nil {
		return overlayCenter(l.prompt.View(), body)
	}
	return body
}
