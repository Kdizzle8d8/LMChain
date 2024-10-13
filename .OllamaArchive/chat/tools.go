package chat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"time"
)

type OllamaReq struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
	Tools    []Tool    `json:"tools"`
}
type OllamaToolCallReq struct {
	Model    string            `json:"model"`
	Messages []ToolCallMessage `json:"messages"`
	Stream   bool              `json:"stream"`
	Tools    []Tool            `json:"tools"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Name    string `json:"name,omitempty"`
}

type Tool struct {
	Type     string                                   `json:"type"`
	Function Function                                 `json:"function"`
	GoFunc   func(map[string]interface{}) interface{} `json:"-"` // New field, not marshaled to JSON
}

type Function struct {
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Parameters  Parameters `json:"parameters"`
}

type Parameters struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties"`
	Required   []string            `json:"required"`
}

type Property struct {
	Type        string   `json:"type"`
	Description string   `json:"description"`
	Enum        []string `json:"enum,omitempty"`
}
type OllamaStreamResponse struct {
	Model     string          `json:"model"`
	CreatedAt string          `json:"created_at"`
	Message   ResponseMessage `json:"message"`
	Done      bool            `json:"done"`
}
type OllamaFinalResponse struct {
	Model              string          `json:"model"`
	CreatedAt          string          `json:"created_at"`
	Message            ResponseMessage `json:"message"`
	DoneReason         string          `json:"done_reason,omitempty"`
	Done               bool            `json:"done"`
	TotalDuration      int64           `json:"total_duration"`
	LoadDuration       int64           `json:"load_duration"`
	PromptEvalCount    int             `json:"prompt_eval_count"`
	PromptEvalDuration int64           `json:"prompt_eval_duration"`
	EvalCount          int             `json:"eval_count"`
	EvalDuration       int64           `json:"eval_duration"`
}

type ToolCall struct {
	Function ToolFunction `json:"function"`
}

type ToolFunction struct {
	Name      string                 `json:"name"`
	Arguments map[string]interface{} `json:"arguments"`
}

type ResponseMessage struct {
	Role    string      `json:"role"`
	Content string      `json:"content"`
	Images  interface{} `json:"images"`
}

type OllamaToolCallResponse struct {
	Model              string          `json:"model"`
	CreatedAt          string          `json:"created_at"`
	Message            ToolCallMessage `json:"message"`
	DoneReason         string          `json:"done_reason"`
	Done               bool            `json:"done"`
	TotalDuration      int64           `json:"total_duration"`
	LoadDuration       int64           `json:"load_duration"`
	PromptEvalCount    int             `json:"prompt_eval_count"`
	PromptEvalDuration int64           `json:"prompt_eval_duration"`
	EvalCount          int             `json:"eval_count"`
	EvalDuration       int64           `json:"eval_duration"`
}

type ToolCallMessage struct {
	Role      string     `json:"role"`
	Content   string     `json:"content"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

func SendMessageWithTool(model string, messages []ToolCallMessage) (string, error) {
	req := OllamaToolCallReq{
		Model:    model,
		Messages: messages,
		Stream:   false,
		Tools:    Tools,
	}
	reqBody, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	resp, err := http.Post("http://localhost:11434/api/chat", "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return "", fmt.Errorf("sending request: %w", err)
	}
	defer resp.Body.Close()
	var resultObj OllamaToolCallResponse
	err = json.NewDecoder(resp.Body).Decode(&resultObj)
	if err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(resultObj.Message.ToolCalls) > 0 {
		log.Println(resultObj.Message.ToolCalls)
		messages = append(messages, resultObj.Message)

		for _, toolCall := range resultObj.Message.ToolCalls {

			// Handle nested function calls
			for argName, argValue := range toolCall.Function.Arguments {
				if strValue, ok := argValue.(string); ok && isToolName(strValue) {
					nestedResult, err := findFunctionAndExecute(strValue, nil)
					if err != nil {
						return "", fmt.Errorf("executing nested tool %s: %w", strValue, err)
					}
					toolCall.Function.Arguments[argName] = nestedResult
				}
			}

			toolRes, err := findFunctionAndExecute(toolCall.Function.Name, toolCall.Function.Arguments)
			log.Println("Tool", toolCall.Function.Name, " Responded with: ", toolRes)
			if err != nil {
				return "", fmt.Errorf("executing tool %s: %w", toolCall.Function.Name, err)
			}
			messages = append(messages, ToolCallMessage{
				Role:    "tool",
				Content: toolRes,
				ToolCalls: []ToolCall{{
					Function: ToolFunction{
						Name:      toolCall.Function.Name,
						Arguments: toolCall.Function.Arguments,
					},
				}},
			})
		}

		return SendMessageWithTool(model, messages)
	}

	log.Println("Final response:", resultObj.Message.Content)
	return resultObj.Message.Content, nil
}

// Helper function to check if a string is a tool name
func isToolName(name string) bool {
	for _, tool := range Tools {
		if tool.Function.Name == name {
			return true
		}
	}
	return false
}

func getTodayDate() string {
	return time.Now().Format("2006-01-02")
}

func getWeekDay(date string) string {
	parsedDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		return ""
	}
	return parsedDate.Weekday().String()
}

func findFunctionAndExecute(functionName string, args map[string]interface{}) (string, error) {
	for _, tool := range Tools {
		if tool.Function.Name == functionName {
			log.Printf("Executing tool: %s with args: %v\n", functionName, args)
			result := tool.GoFunc(args)
			return fmt.Sprintf("%v", result), nil
		}
	}
	return "", fmt.Errorf("function not found")
}

var Tools = []Tool{
	{
		Type: "function",
		Function: Function{
			Name:        "getTodayDate",
			Description: "Returns the current date.",
			Parameters: Parameters{
				Type:       "object",
				Properties: map[string]Property{},
				Required:   []string{},
			},
		},
		GoFunc: func(args map[string]interface{}) interface{} {
			return getTodayDate()
		},
	},
	{
		Type: "function",
		Function: Function{
			Name:        "getWeekDay",
			Description: "Returns the day of the week for a given date.",
			Parameters: Parameters{
				Type: "object",
				Properties: map[string]Property{
					"date": {
						Type:        "string",
						Description: "The date to get the weekday for, in the format 'YYYY-MM-DD'",
					},
				},
				Required: []string{"date"},
			},
		},
		GoFunc: func(args map[string]interface{}) interface{} {
			date, ok := args["date"].(string)
			if !ok {
				return "Invalid date format"
			}
			return getWeekDay(date)
		},
	},
	// {
	// 	Type: "function",
	// 	Function: Function{
	// 		Name:        "executeArbitraryPythonCode",
	// 		Description: "Executes arbitrary Python code and returns the output as a string. Only use for math related functions, or when another function isn't availbile",
	// 		Parameters: Parameters{
	// 			Type: "object",
	// 			Properties: map[string]Property{
	// 				"code": {
	// 					Type:        "string",
	// 					Description: "The Python code to execute as a string. This will be executed in a temporary file. use valid python syntax. Don't use returns, only print.",
	// 				},
	// 			},
	// 			Required: []string{"code"},
	// 		},
	// 	},
	// 	GoFunc: func(args map[string]interface{}) interface{} {
	// 		code, ok := args["code"].(string)
	// 		if !ok {
	// 			return "Invalid code format"
	// 		}
	// 		output, err := executeArbitraryPythonCode(code)
	// 		if err != nil {
	// 			return fmt.Sprintf("Error executing Python code: %v", err)
	// 		}
	// 		return output
	// 	},
	// },
}

func executeArbitraryPythonCode(code string) (string, error) {
	os.WriteFile("code.py", []byte(code), 0644)
	out, err := exec.Command("python3", "code.py").Output()
	if err != nil {
		return "", fmt.Errorf("executing python code: %w", err)
	}
	return string(out), nil
}
