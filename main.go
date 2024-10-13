package main

import (
	"os"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	godotenv.Load()
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	prompt := openai.SystemMessage("You are a helpful assistant. Your training data is not realtime, so you've been equiped with function calls in order to gain access to the most up to date and accurate info. When executing python code, add a print statement using the print() function or there will be no output.")
	messages := []openai.ChatCompletionMessageParamUnion{}
	if prompt != nil {
		messages = append(messages, prompt)
	}
}

type model struct {
	viewport viewport.Model
	messages []openai.ChatCompletionMessageParamUnion
	textarea textarea.Model
	err      error
}

// View implements tea.Model.

func initialModel() model {
	ti := textarea.New()
	ti.Placeholder = "Message the assistant..."
	ti.Focus()

	ti.SetWidth(40)
	ti.SetHeight(7)

	vp := viewport.New(40, 12)
	vp.SetContent("Send a message")

	return model{
		textarea: ti,
		viewport: vp,
		messages: []openai.ChatCompletionMessageParamUnion{},
		err:      nil,
	}
}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			return m, processInput
		case "shift+enter":
			return m, processInput
		}
	}

	return m, nil
}

func (m model) View() string {
	return ""
}
