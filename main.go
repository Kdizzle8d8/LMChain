package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

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

	if len(os.Args) > 1 {
		if os.Args[1] == "ask" {
			messages := []openai.ChatCompletionMessageParamUnion{
				openai.UserMessage(strings.Join(os.Args[2:], " ")),
			}
			chat.SendMessage(client, chat.Tools, messages)
		} else if os.Args[1] == "chat" {
			fmt.Println("Entering chat mode. Type 'exit' to quit.")
			chatLoop(client, []openai.ChatCompletionMessageParamUnion{})
		} else {
			fmt.Println("Invalid command")
		}
	}
	// Print the number of messages in the conversation

	// Optionally, you can do more with the messages array here
	// For example, you could print the role of the last message:
}

func chatLoop(client *openai.Client, messages []openai.ChatCompletionMessageParamUnion) {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("\033[34m\n[You] > \033[0m")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)
		if input == "exit" {
			break
		}
		messages = append(messages, openai.UserMessage(input))
		chat.SendMessage(client, chat.Tools, messages)
	}
}
