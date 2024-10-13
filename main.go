package main

import (
	"os"

	"LMChain/chat"

	"github.com/joho/godotenv"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
)

func main() {
	godotenv.Load()
	client := openai.NewClient(
		option.WithAPIKey(os.Getenv("OPENAI_API_KEY")), // defaults to os.LookupEnv("OPENAI_API_KEY")
	)
	chat.SendMessage(client, chat.Tools, []openai.ChatCompletionMessageParamUnion{
		openai.UserMessage(os.Args[1]),
	})
}
