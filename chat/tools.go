package chat

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/openai/openai-go"
)

var Tools = []openai.ChatCompletionToolParam{
	{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("add"),
			Description: openai.String("Add two numbers"),
			Parameters: openai.F(openai.FunctionParameters{
				"type": "object",
				"properties": map[string]interface{}{
					"a": map[string]interface{}{"type": "number"},
					"b": map[string]interface{}{"type": "number"},
				},
				"required":             []string{"a", "b"},
				"additionalProperties": false,
			}),
			Strict: openai.F(true),
		}),
	},
	{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("executeArbitraryPython"),
			Description: openai.String("Execute arbitrary python code"),
			Parameters: openai.F(openai.FunctionParameters{
				"type":        "object",
				"description": "Execute arbitrary python code. Use this primarily for math calculations. ",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "The python code to execute as a string. add a print statement to ensure output.",
					},
				},
				"required":             []string{"code"},
				"additionalProperties": false,
			}),
		}),
	},
	{
		Type: openai.F(openai.ChatCompletionToolTypeFunction),
		Function: openai.F(openai.FunctionDefinitionParam{
			Name:        openai.String("executeArbitraryCommand"),
			Description: openai.String("Execute arbitrary command"),
			Parameters: openai.F(openai.FunctionParameters{
				"type":        "object",
				"description": "Execute arbitrary shell command. Use this for tasks that require shell access. Do not do any destructive operations.",
				"properties": map[string]interface{}{
					"command": map[string]interface{}{
						"type":        "string",
						"description": "The shell command to execute. Use proper bash syntax.",
					},
				},
				"required":             []string{"command"},
				"additionalProperties": false,
			}),
		}),
	},
}

var ToolMap = map[string]func(args map[string]interface{}) string{
	"getToday":                getToday,
	"add":                     add,
	"executeArbitraryPython":  executeArbitraryPython,
	"executeArbitraryCommand": executeArbitraryCommand,
}

func getToday(args map[string]interface{}) string {
	return time.Now().Format(time.DateOnly)
}

func add(args map[string]interface{}) string {
	a, ok := args["a"].(float64)
	if !ok {
		return "Error: Invalid input. a must be a number."
	}
	b, ok := args["b"].(float64)
	if !ok {
		return "Error: Invalid input. b must be a number."
	}
	return fmt.Sprintf("%f", a+b)
}

func executeArbitraryPython(args map[string]interface{}) string {
	code, ok := args["code"].(string)
	if !ok {
		return "Error: Invalid input. code must be a string."
	}
	os.WriteFile("temp.py", []byte(code), 0644)
	out, err := exec.Command("python", "temp.py").Output()
	fmt.Println("Python output:", string(out))
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return string(out)
}

func executeArbitraryCommand(args map[string]interface{}) string {
	command, ok := args["command"].(string)
	if !ok {
		return "Error: Invalid input. command must be a string."
	}
	out, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}
	return string(out)
}
