package chat

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"github.com/openai/openai-go"
)

func SendMessage(client *openai.Client, tools []openai.ChatCompletionToolParam, messages []openai.ChatCompletionMessageParamUnion) []openai.ChatCompletionMessageParamUnion {
	stream := client.Chat.Completions.NewStreaming(context.TODO(), openai.ChatCompletionNewParams{
		Model:    openai.F(openai.ChatModelGPT4o),
		Tools:    openai.F(tools),
		Messages: openai.F(messages),
	})

	acc := openai.ChatCompletionAccumulator{}
	channel := make(chan ChannelMessage)
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
		if tool, ok := acc.JustFinishedToolCall(); ok {
			channel <- ChannelMessage{
				Type:    "tool call",
				Content: fmt.Sprintf("%+v", tool),
			}
		}
		if len(chunk.Choices) > 0 {
			channel <- ChannelMessage{
				Type:    "assistant chunk",
				Content: chunk.Choices[0].Delta.Content,
			}
		}
	}
	if err := stream.Err(); err != nil {
		log.Printf("Error in stream: %v", err)
		return messages
	}
	if len(acc.Choices) == 0 {
		log.Println("No choices returned from the API")
		return messages
	}
	messages = append(messages, acc.Choices[0].Message)

	for _, toolCall := range acc.Choices[0].Message.ToolCalls {
		toolInfo, ok := ToolMap[toolCall.Function.Name]
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
		res := toolInfo.Func(args)
		messages = append(messages, openai.ToolMessage(toolCall.ID, res))
		channel <- ChannelMessage{
			Type:    "tool result",
			Content: fmt.Sprintf("%s: %v", toolCall.Function.Name, res),
		}
	}

	if len(acc.Choices[0].Message.ToolCalls) > 0 {
		return SendMessage(client, tools, messages)
	}

	return messages
}

type ChannelMessage struct {
	Type    string
	Content string
}
