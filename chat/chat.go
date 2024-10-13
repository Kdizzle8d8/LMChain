package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
)

func SendMessage(message, filePath, model string) (string, error) {
	prompt := fmt.Sprintf("User Prompt: %s", message)
	if filePath != "" {
		content, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("reading file: %w", err)
		}
		prompt += fmt.Sprintf("\nAttachments: Count: 1 File Path: %s Content: %s", filePath, content)
	}

	req := OllamaReq{
		Model: model,
		Messages: []Message{
			{Role: "user", Content: prompt},
		},
		Stream: true,
	}

	return SendOllamaRequest(req)
}

func SendOllamaRequest(req OllamaReq) (string, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := http.Post(os.Getenv("OLLAMA_URL")+"/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()
	return HandleResponseStream(resp.Body)
}

func HandleResponseStream(body io.ReadCloser) (string, error) {
	defer body.Close()

	decoder := json.NewDecoder(body)
	fullResponse := ""

	for {
		var streamResponse OllamaStreamResponse

		if err := decoder.Decode(&streamResponse); err != nil {
			if err == io.EOF {
				break
			}
			return fullResponse, fmt.Errorf("decoding stream: %w", err)
		}

		fmt.Print(streamResponse.Message.Content)
		fullResponse += streamResponse.Message.Content

		if streamResponse.Done {
			break
		}
	}
	return fullResponse, nil
}
