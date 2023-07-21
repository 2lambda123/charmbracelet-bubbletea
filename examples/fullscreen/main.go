package fullscreen

// A simple program that opens the alternate screen buffer then counts down
// from 5 and then exits.

import (
	"fmt"
	"log"
	"time"

	tea "github.com/rprtr258/bubbletea"
)

type model int

type tickMsg time.Time

func Main() {
	p := tea.NewProgram(model(5)).WithAltScreen()
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(tick(), tea.EnterAltScreen)
}

func (m model) Update(message tea.Msg) (model, tea.Cmd) {
	switch msg := message.(type) {
	case tea.MsgKey:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			return m, tea.Quit
		}

	case tickMsg:
		m--
		if m <= 0 {
			return m, tea.Quit
		}
		return m, tick()
	}

	return m, nil
}

func (m model) View(r tea.Renderer) {
	r.Write(fmt.Sprintf("\n\n     Hi. This program will exit in %d seconds...", m))
	return
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}
