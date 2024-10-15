package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/openai/openai-go"
)

type model struct {
	client             *openai.Client
	messages           []openai.ChatCompletionMessageParamUnion
	messageChannel     chan string
	viewport           viewport.Model
	textarea           textarea.Model
	err                error
	waitingForResponse bool
	width              int
	height             int
}

func chatModel(client *openai.Client, messages []openai.ChatCompletionMessageParamUnion) model {
	ti := textarea.New()
	ti.Placeholder = "Type your message here..."
	ti.Focus()

	ti.SetWidth(80) // Set a default width
	ti.SetHeight(3) // Set height to 3 lines

	vp := viewport.New(80, 20) // Adjust viewport width to account for borders
	vpContent := []string{}
	vpContent = updateMessages(messages, vpContent, 80)
	vp.SetContent(strings.Join(vpContent, "\n"))

	return model{
		textarea:           ti,
		viewport:           vp,
		messages:           messages,
		messageChannel:     make(chan string),
		client:             client,
		err:                nil,
		waitingForResponse: false,
		width:              80, // Default width
		height:             24, // Default height
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKeyMsg(msg)
	case tea.WindowSizeMsg:
		return m.handleWindowSizeMsg(msg)
	case string:
		return m.handleStringMsg(msg)
	case nil:
		m.textarea.Focus()
	}

	// Always update viewport
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// Ensure content is visible after updates
	// m.viewport.SetContent(strings.Join(updateMessages(m.messages, []string{}, m.width), "\n"))

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return lipgloss.JoinVertical(lipgloss.Left,
		m.viewport.View(),
		lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("86")).
			Width(m.width).
			Padding(0, 1).
			Render(m.textarea.View()),
		helpStyle("↑/↓: scroll • tab: switch focus • ctrl+c: quit"),
	)
}

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render

func checkForMessage(messageChannel chan string) tea.Cmd {
	return func() tea.Msg {
		msg, ok := <-messageChannel
		if !ok {
			return nil // Channel closed, no more messages
		}
		return msg
	}
}

func sendMessage(messageChannel chan string, client *openai.Client, m model) {
	stream := client.Chat.Completions.NewStreaming(context.TODO(), openai.ChatCompletionNewParams{
		Model:    openai.F(openai.ChatModelGPT4o),
		Messages: openai.F(m.messages),
	})
	acc := openai.ChatCompletionAccumulator{}
	for stream.Next() {
		chunk := stream.Current()
		acc.AddChunk(chunk)
		if tool, ok := acc.JustFinishedToolCall(); ok {
			messageChannel <- "Detected tool call: " + tool.Name
		}
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			time.Sleep(10 * time.Millisecond)
			messageChannel <- chunk.Choices[0].Delta.Content
		}
	}
	if err := stream.Err(); err != nil {
		log.Printf("Error in stream: %v", err)
		return
	}
	messageChannel <- "<<<END_OF_STREAM>>>"
}

func getTextFromMessage(message openai.ChatCompletionMessageParamUnion) string {
	msgJson, err := json.Marshal(message)
	if err != nil {
		log.Println("Error marshalling message:", err)
		return ""
	}
	os.WriteFile("message.json", msgJson, 0644)
	var messageJson messageJson
	err = json.Unmarshal(msgJson, &messageJson)
	if err != nil {
		log.Println("Error unmarshalling message:", err)
		return ""
	}
	os.WriteFile("message.txt", []byte(messageJson.Content[0].Text), 0644)
	if len(messageJson.Content) > 0 && messageJson.Content[0].Text != "" {
		os.WriteFile("message.txt", []byte(messageJson.Content[0].Text), 0644)
		return messageJson.Content[0].Text
	}
	return ""
}

type messageJson struct {
	Content []struct {
		Text string `json:"text"`
		Type string `json:"type"`
	} `json:"content"`
	Role string `json:"role"`
}

func updateMessages(messages []openai.ChatCompletionMessageParamUnion, vpContent []string, width int) []string {
	userStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("15"))     // White
	assistantStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("2")) // Green
	boldText := lipgloss.NewStyle().Bold(true)
	for _, message := range messages {
		var prefix, content string
		switch msg := message.(type) {
		case openai.ChatCompletionUserMessageParam:
			prefix = boldText.Render("User: ")
			content = userStyle.Render(getTextFromMessage(msg))
		case openai.ChatCompletionAssistantMessageParam:
			prefix = boldText.Render("Assistant: ")
			content = assistantStyle.Render(getTextFromMessage(msg))
		}
		vpContent = append(vpContent, prefix+content)
	}
	return vpContent
}

// Helper function to find the maximum of multiple integers
func max(nums ...int) int {
	if len(nums) == 0 {
		return 0
	}
	max := nums[0]
	for _, num := range nums[1:] {
		if num > max {
			max = num
		}
	}
	return max
}

func (m model) handleKeyMsg(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var (
		cmd  tea.Cmd
		cmds []tea.Cmd
	)

	switch msg.Type {
	case tea.KeyCtrlC, tea.KeyEsc:
		return m, tea.Quit
	case tea.KeyEnter:
		if m.textarea.Focused() && !m.waitingForResponse {
			userInput := m.textarea.Value()
			m.messages = append(m.messages, openai.UserMessage(userInput))
			m.viewport.GotoBottom()
			m.textarea.Reset()
			m.textarea.Blur()
			m.waitingForResponse = true
			go sendMessage(m.messageChannel, m.client, m)
			cmds = append(cmds, checkForMessage(m.messageChannel))
		}
	case tea.KeyUp, tea.KeyDown, tea.KeyPgUp, tea.KeyPgDown:
		if !m.textarea.Focused() {
			m.viewport, cmd = m.viewport.Update(msg)
			cmds = append(cmds, cmd)
		}
	case tea.KeyTab:
		if m.textarea.Focused() {
			m.textarea.Blur()
		} else {
			m.textarea.Focus()
		}
	default:
		if m.textarea.Focused() {
			m.textarea, cmd = m.textarea.Update(msg)
			cmds = append(cmds, cmd)
		}
	}

	return m, tea.Batch(cmds...)
}

func (m model) handleWindowSizeMsg(msg tea.WindowSizeMsg) (tea.Model, tea.Cmd) {
	m.width = msg.Width
	m.height = msg.Height

	textareaHeight := 3 // Fixed height for the textarea
	helpTextHeight := 1 // Height for the help text
	borderHeight := 2   // Account for top and bottom borders of the textarea

	// Calculate the available height for the viewport
	viewportHeight := m.height - textareaHeight - helpTextHeight - borderHeight

	m.viewport.Width = m.width
	m.viewport.Height = viewportHeight

	m.textarea.SetWidth(m.width - 4) // Account for left and right borders

	// Rerender the content to fit the new width
	viewPortMessages := updateMessages(m.messages, []string{}, m.width)
	m.viewport.SetContent(strings.Join(viewPortMessages, "\n"))

	return m, nil
}

func (m model) handleStringMsg(msg string) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	if msg == "<<<END_OF_STREAM>>>" {
		m.waitingForResponse = false
		m.textarea.Focus()
	} else {
		if len(m.messages) > 0 && msg != "" {
			lastMessage := m.messages[len(m.messages)-1]
			switch lastMsg := lastMessage.(type) {
			case openai.ChatCompletionAssistantMessageParam:
				m.messages[len(m.messages)-1] = openai.AssistantMessage(getTextFromMessage(lastMsg) + msg)
			default:
				m.messages = append(m.messages, openai.AssistantMessage(msg))
			}
		}
	}

	viewPortMessages := updateMessages(m.messages, []string{}, m.width)
	m.viewport.SetContent(strings.Join(viewPortMessages, "\n\n"))

	// Adjust scrolling behavior
	if m.viewport.ScrollPercent() >= 0.9 || m.waitingForResponse {
		m.viewport.GotoBottom()
	}

	cmds = append(cmds, checkForMessage(m.messageChannel))

	return m, tea.Batch(cmds...)
}
