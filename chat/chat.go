package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/openai/openai-go"
)

func SendMessage(client *openai.Client, tools []openai.ChatCompletionToolParam, messages []openai.ChatCompletionMessageParamUnion) {
	stream := client.Chat.Completions.NewStreaming(context.TODO(), openai.ChatCompletionNewParams{
		Model:    openai.F(openai.ChatModelGPT4o),
		Tools:    openai.F(tools),
		Messages: openai.F(messages),
	})
	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
		if tool, ok := acc.JustFinishedToolCall(); ok {
			fmt.Println("Finished tool call:", tool.Name, "With arguments:", tool.Arguments)
		}
		if len(chunk.Choices) > 0 {
			fmt.Print(chunk.Choices[0].Delta.Content)
		}
	}
	if err := stream.Err(); err != nil {
		log.Printf("Error in stream: %v", err)
		return
	}
	if len(acc.Choices) == 0 {
		log.Println("No choices returned from the API")
		return
	}

	// Add the assistant's message to the conversation
	messages = append(messages, acc.Choices[0].Message)

	// Handle all tool calls before sending the next message
	for _, toolCall := range acc.Choices[0].Message.ToolCalls {
		function, ok := ToolMap[toolCall.Function.Name]
		if !ok {
			log.Printf("Error: Unknown tool call %s", toolCall.Function.Name)
			continue
		}
		var args map[string]interface{}
		err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args)
		if err != nil {
			log.Printf("Error unmarshalling arguments: %v", err)
			continue
		}
		res := function(args)
		// Add the tool response to the conversation
		messages = append(messages, openai.ToolMessage(toolCall.ID, res))
	}

	// Only send a new message if there were tool calls
	if len(acc.Choices[0].Message.ToolCalls) > 0 {
		SendMessage(client, tools, messages)
	}
}
