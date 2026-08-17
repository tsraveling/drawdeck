package main

import (
	tea "charm.land/bubbletea/v2"
)

type mode int

const (
	modeList mode = iota
	modeDetail
)

// owns mode switching and global messages; the views own everything else
type rootModel struct {
	mode   mode
	list   listView
	detail detailView
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
	}

	var cmd tea.Cmd
	if m.mode == modeDetail {
		m.detail, cmd = m.detail.Update(msg)
	} else {
		m.list, cmd = m.list.Update(msg)
	}
	return m, cmd
}

func (m rootModel) View() tea.View {
	content := m.list.View()
	if m.mode == modeDetail {
		content = m.detail.View()
	}

	v := tea.NewView(content)
	v.AltScreen = true
	v.KeyboardEnhancements = tea.KeyboardEnhancements{ReportEventTypes: true}
	return v
}
