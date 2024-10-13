package main

import (
	"LMChain/chat"
	"fmt"
	"log"
	"os"
	"time"
)

const (
    url   = "http://localhost:11434"
	model = "llama3.1"
)

func main() {

	if len(os.Args) < 2 {
		log.Fatal("Usage: go run main.go <message>")
	}

	response, err := chat.SendMessageWithTool(model, []chat.ToolCallMessage{
		{Role: "user", Content: os.Args[1]},
	})
	if err != nil {
		log.Fatal("Error:", err)
	}

	currentTime := time.Now().Format("2006-01-02 15:04:05")
	content := fmt.Sprintf("%s\n%s\n", currentTime, response)
	os.WriteFile("responseLog.txt", []byte(content), 0644)
}
