package main

import (
	tea "github.com/charmbracelet/bubbletea"
	"mcop/src/ui"
)

func startTUI() {
	appModel := ui.NewAppModel()
	p := tea.NewProgram(appModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}

func startTUIWithServer(url string) {
	appModel := ui.NewAppModel()
	appModel.SetInitialServerURL(url)
	p := tea.NewProgram(appModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}