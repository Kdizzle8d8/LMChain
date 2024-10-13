package main

// A simple program demonstrating the textarea component from the Bubbles
// component library.

import (
	"fmt"
	"log"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}

type errMsg error

type model struct {
	viewport viewport.Model
	textarea textarea.Model
	messages []string
	err      error
}

func initialModel() model {
	ti := textarea.New()
	ti.Placeholder = "Once upon a time..."
	ti.Focus()

	ti.SetWidth(60)
	ti.SetHeight(10)

	vp := viewport.New(40, 5)
	vp.SetContent("Send a message")

	return model{
		textarea: ti,
		messages: []string{},
		viewport: vp,
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.textarea.Focused() {
				m.textarea.Blur()
			}
		case tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyShiftTab:
			v := m.textarea.Value()
			if v == "" {
				return m, nil
			}
			m.messages = append(m.messages, "You: "+v)
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
            return m,nil
		default:
			if !m.textarea.Focused() {
				cmd = m.textarea.Focus()
				cmds = append(cmds, cmd)
			}
		}

	// We handle errors just like any other message
	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textarea, cmd = m.textarea.Update(msg)
	cmds = append(cmds, cmd)
	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}
