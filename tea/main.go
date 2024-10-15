package main

import (
	"os"

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
	messages := []openai.ChatCompletionMessageParamUnion{
		openai.SystemMessage("You are a helpful assistant."),
	}
	p := tea.NewProgram(chatModel(client, messages), tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
