package agent

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/tmc/langchaingo/llms"
)

// FileEdit represents a proposed file change for self-healing
type FileEdit struct {
	Path    string
	Content string
}

// ResolveForHealing runs the agent and captures write_file tool calls as proposed edits.
// Returns the text proposal and any file edits for the healer to show in TUI.
func (a *Agent) ResolveForHealing(ctx context.Context, intent string) (proposal string, edits []FileEdit, err error) {
	systemPrompt := `You are the Aerostack AI Agent helping fix a failed command.
Analyze the error and project context. If the fix requires editing files, use the write_file tool.
Use read_file and list_dir to gather context first. Be concise.`

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, intent),
	}

	const maxTurns = 15
	for turn := 0; turn < maxTurns; turn++ {
		response, err := a.llm.GenerateContent(ctx, messages, llms.WithTools(a.GetTools()))
		if err != nil {
			return "", nil, fmt.Errorf("LLM error: %w", err)
		}
		if len(response.Choices) == 0 {
			return "", nil, fmt.Errorf("no response from LLM")
		}
		choice := response.Choices[0]

		if len(choice.ToolCalls) > 0 {
			assistantResp := llms.MessageContent{Role: llms.ChatMessageTypeAI}
			for _, tc := range choice.ToolCalls {
				assistantResp.Parts = append(assistantResp.Parts, tc)
			}
			messages = append(messages, assistantResp)

			for _, tc := range choice.ToolCalls {
				if tc.FunctionCall == nil {
					continue
				}
				name := tc.FunctionCall.Name
				args := tc.FunctionCall.Arguments
				var toolResult string

				if name == "write_file" {
					var params struct {
						Path    string `json:"path"`
						Content string `json:"content"`
					}
					if json.Unmarshal([]byte(args), &params) == nil {
						edits = append(edits, FileEdit{Path: params.Path, Content: params.Content})
						toolResult = fmt.Sprintf("Successfully wrote %d bytes to %s", len(params.Content), params.Path)
					} else {
						toolResult = "Error: failed to parse arguments"
					}
				} else {
					result, execErr := a.ExecuteToolCall(ctx, name, args)
					if execErr != nil {
						toolResult = "Error: " + execErr.Error()
					} else {
						toolResult = result
					}
				}

				toolResp := llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{ToolCallID: tc.ID, Name: name, Content: toolResult},
					},
				}
				messages = append(messages, toolResp)
			}
			continue
		}

		proposal = choice.Content
		return proposal, edits, nil
	}
	return "", edits, fmt.Errorf("max turns exceeded")
}

// Resolve processes a user's intent and executes actions with multi-turn tool loop
func (a *Agent) Resolve(ctx context.Context, intent string) error {
	systemPrompt := `You are the Aerostack AI Agent. You are a "Project-Aware" CLI tool.
Your goal is to help the user understand and modify their project.
You have access to tools to read files, list directories, search for symbols, and write files.
Use these tools to gather information before answering.
If you need to explore the code, start by listing files or searching for relevant symbols.
Be concise and direct.`

	messages := []llms.MessageContent{
		llms.TextParts(llms.ChatMessageTypeSystem, systemPrompt),
		llms.TextParts(llms.ChatMessageTypeHuman, intent),
	}

	const maxTurns = 15
	for turn := 0; turn < maxTurns; turn++ {
		fmt.Println("🤖 Thinking...")
		response, err := a.llm.GenerateContent(ctx, messages, llms.WithTools(a.GetTools()))
		if err != nil {
			return fmt.Errorf("LLM error: %w", err)
		}
		if len(response.Choices) == 0 {
			return fmt.Errorf("no response from LLM")
		}
		choice := response.Choices[0]

		if len(choice.ToolCalls) > 0 {
			assistantResp := llms.MessageContent{Role: llms.ChatMessageTypeAI}
			for _, tc := range choice.ToolCalls {
				assistantResp.Parts = append(assistantResp.Parts, tc)
			}
			messages = append(messages, assistantResp)

			for _, tc := range choice.ToolCalls {
				if tc.FunctionCall == nil {
					continue
				}
				name := tc.FunctionCall.Name
				args := tc.FunctionCall.Arguments
				fmt.Printf("🛠️  %s(%s)\n", name, args)

				result, execErr := a.ExecuteToolCall(ctx, name, args)
				if execErr != nil {
					fmt.Printf("❌ Error: %v\n", execErr)
					result = "Error: " + execErr.Error()
				} else {
					fmt.Printf("📄 %s\n", result)
				}

				toolResp := llms.MessageContent{
					Role: llms.ChatMessageTypeTool,
					Parts: []llms.ContentPart{
						llms.ToolCallResponse{ToolCallID: tc.ID, Name: name, Content: result},
					},
				}
				messages = append(messages, toolResp)
			}
			continue
		}

		fmt.Printf("🤖 %s\n", choice.Content)
		return nil
	}
	return fmt.Errorf("max turns exceeded")
}
