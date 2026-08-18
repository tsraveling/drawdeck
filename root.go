package main

import (
	tea "charm.land/bubbletea/v2"
)

type mode int

const (
	modeList mode = iota
	modeDetail
	modeTournament
)

// owns mode switching and global messages; the views own everything else
type rootModel struct {
	mode       mode
	list       listView
	detail     detailView
	tournament tournament
}

func makeRootModel(reg *registry, focus string) rootModel {
	return rootModel{mode: modeList, list: makeListView(reg, focus)}
}

func (m rootModel) Init() tea.Cmd { return nil }

func (m rootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		cfg.setSize(msg.Width, msg.Height)
		return m, nil

	case tea.KeyboardEnhancementsMsg:
		// hold-to-draw becomes available the moment this lands; tap always works
		cfg.holdSupported = msg.SupportsEventTypes()
		return m, nil

	case openDeckMsg:
		d, err := loadDeck(msg.path)
		if err != nil {
			return m, nil
		}
		m.detail = makeDetailView(d)
		m.mode = modeDetail
		return m, nil

	case backToListMsg:
		m.mode = modeList
		return m, nil

	case startTournamentMsg:
		d, err := loadDeck(msg.path)
		if err != nil {
			return m, nil
		}
		// entry wipes any prior play so the bracket starts from a clean deck
		d.resetAll()
		if err := d.write(); err != nil {
			return m, nil
		}
		var cmd tea.Cmd
		m.tournament, cmd = makeTournament(d)
		m.mode = modeTournament
		return m, cmd

	case leaveTournamentMsg:
		m.mode = modeList
		m.list.reload()
		return m, nil
	}

	var cmd tea.Cmd
	switch m.mode {
	case modeTournament:
		m.tournament, cmd = m.tournament.Update(msg)
	case modeDetail:
		m.detail, cmd = m.detail.Update(msg)
	default:
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m rootModel) View() tea.View {
	var content string
	switch m.mode {
	case modeTournament:
		content = m.tournament.View()
	case modeDetail:
		content = m.detail.View()
	default:
		content = m.list.View()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.KeyboardEnhancements = tea.KeyboardEnhancements{ReportEventTypes: true}
	return v
}
