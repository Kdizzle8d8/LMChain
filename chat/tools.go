package chat

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/alecthomas/chroma/quick"
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
				"description": "Execute arbitrary python code. Use this primarily for math calculations. To ensure output add a print statement always use the print function.",
				"properties": map[string]interface{}{
					"code": map[string]interface{}{
						"type":        "string",
						"description": "The python code to execute as a string. ",
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
				"description": "Execute arbitrary shell command. Use this for tasks that require shell access. Do not do any destructive operations. Use mv ~/.Trash instead of rm.",
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

var ToolMap = map[string]struct {
	Func  func(args map[string]interface{}) string
	Print bool
}{
	"getToday":                {Func: getToday, Print: true},
	"add":                     {Func: add, Print: true},
	"executeArbitraryPython":  {Func: executeArbitraryPython, Print: false},
	"executeArbitraryCommand": {Func: executeArbitraryCommand, Print: false},
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

	// Execute the Python code
	os.WriteFile("temp.py", []byte(code), 0644)
	cmd := exec.Command("python", "temp.py")
	out, err := cmd.CombinedOutput()

	// Display the code
	fmt.Println("\033[33m┌" + strings.Repeat("─", 78) + "┐\033[0m")
	lines := strings.Split(code, "\n")
	for _, line := range lines {
		paddedLine := padOrTruncate(line, 76)
		fmt.Print("\033[33m│\033[0m ")
		quick.Highlight(os.Stdout, paddedLine, "python", "terminal256", "dracula")
		fmt.Println(" \033[33m│\033[0m")
	}
	fmt.Println("\033[33m└" + strings.Repeat("─", 78) + "┘\033[0m")

	// Display the output or error in a separate box
	output := strings.TrimSpace(string(out))
	var outputLines []string
	var boxTitle string

	if err != nil {
		boxTitle = "Python Error:"
		errorMsg := err.Error()
		outputLines = append(strings.Split(output, "\n"), "Error: "+errorMsg)
	} else {
		boxTitle = "Python Output:"
		outputLines = strings.Split(output, "\n")
	}

	fmt.Println("\033[33m┌" + strings.Repeat("─", 78) + "┐\033[0m")
	fmt.Println("\033[33m│ " + boxTitle + strings.Repeat(" ", 77-len(boxTitle)) + "│\033[0m")
	fmt.Println("\033[33m├" + strings.Repeat("─", 78) + "┤\033[0m")

	if len(outputLines) == 1 && outputLines[0] == "" {
		fmt.Print("\033[33m│\033[0m ")
		fmt.Print(padOrTruncate("(No output)", 76))
		fmt.Println(" \033[33m│\033[0m")
	} else {
		for _, line := range outputLines {
			fmt.Print("\033[33m│\033[0m ")
			fmt.Print(padOrTruncate(line, 76))
			fmt.Println(" \033[33m│\033[0m")
		}
	}

	fmt.Println("\033[33m└" + strings.Repeat("─", 78) + "┘\033[0m")

	return output
}

func executeArbitraryCommand(args map[string]interface{}) string {
	command, ok := args["command"].(string)
	if !ok {
		return "Error: Invalid input. command must be a string."
	}
	if strings.Contains(command, "rm") || strings.Contains(command, "delete") {
		return "Error: This command is not allowed. use trash instead."
	}


	// Execute the command
	out, err := exec.Command("bash", "-c", command).Output()
	if err != nil {
		return fmt.Sprintf("Error: %s", err)
	}

	// Prepare the output
	output := strings.TrimSpace(string(out))
	outputLines := strings.Split(output, "\n")

	fmt.Println("\033[33m┌" + strings.Repeat("─", 78) + "┐\033[0m")

	// Print the command
	fmt.Print("\033[33m│\033[0m ")
	quick.Highlight(os.Stdout, padOrTruncate(command, 76), "bash", "terminal256", "monokai")
	fmt.Println(" \033[33m│\033[0m")

	// Print a separator
	fmt.Println("\033[33m├" + strings.Repeat("─", 78) + "┤\033[0m")

	// Print the output or a message if there's no output
	if len(outputLines) == 1 && outputLines[0] == "" {
		fmt.Print("\033[33m│\033[0m ")
		fmt.Print(padOrTruncate("(No output)", 76))
		fmt.Println(" \033[33m│\033[0m")
	} else {
		for _, line := range outputLines {
			fmt.Print("\033[33m│\033[0m ")
			fmt.Print(padOrTruncate(line, 76))
			fmt.Println(" \033[33m│\033[0m")
		}
	}

	fmt.Println("\033[33m└" + strings.Repeat("─", 78) + "┘\033[0m")

	if output == "" {
		return "(No output)"
	}
	return output
}

// Helper function to pad or truncate a string to a specific length
func padOrTruncate(s string, length int) string {
	if len(s) > length {
		return s[:length-3] + "..."
	}
	return s + strings.Repeat(" ", length-len(s))
}
