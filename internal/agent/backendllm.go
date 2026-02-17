package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/tmc/langchaingo/llms"
)

// BackendLLM implements llms.Model by calling the Aerostack API proxy.
// Used when user has no own keys but has run 'aerostack login'.
type BackendLLM struct {
	baseURL string
	apiKey  string
	client  *http.Client
}

// NewBackendLLM creates an LLM that uses the Aerostack AI proxy.
func NewBackendLLM(baseURL, apiKey string) *BackendLLM {
	if baseURL == "" {
		baseURL = "https://api.aerocall.ai"
	}
	return &BackendLLM{
		baseURL: baseURL,
		apiKey:  apiKey,
		client:  http.DefaultClient,
	}
}

// Call implements llms.Model (simplified text-only interface).
func (b *BackendLLM) Call(ctx context.Context, prompt string, options ...llms.CallOption) (string, error) {
	messages := []llms.MessageContent{llms.TextParts(llms.ChatMessageTypeHuman, prompt)}
	resp, err := b.GenerateContent(ctx, messages, options...)
	if err != nil {
		return "", err
	}
	if len(resp.Choices) == 0 {
		return "", fmt.Errorf("empty response")
	}
	return resp.Choices[0].Content, nil
}

// GenerateContent implements llms.Model.
func (b *BackendLLM) GenerateContent(ctx context.Context, messages []llms.MessageContent, options ...llms.CallOption) (*llms.ContentResponse, error) {
	opts := &llms.CallOptions{}
	for _, o := range options {
		o(opts)
	}

	apiMessages, err := convertToAPIMessages(messages)
	if err != nil {
		return nil, err
	}

	reqBody := map[string]interface{}{
		"messages": apiMessages,
	}
	if len(opts.Tools) > 0 {
		reqBody["tools"] = convertToolsToAPI(opts.Tools)
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	url := b.baseURL + "/api/v1/cli/ai/complete"
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-API-Key", b.apiKey)

	resp, err := b.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("backend request failed: %w", err)
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		var errBody struct {
			Error struct {
				Message string `json:"message"`
			} `json:"error"`
		}
		_ = json.Unmarshal(body, &errBody)
		msg := errBody.Error.Message
		if msg == "" {
			msg = string(body)
		}
		return nil, fmt.Errorf("AI proxy error (%d): %s", resp.StatusCode, msg)
	}

	var result struct {
		Content   string `json:"content"`
		ToolCalls []struct {
			ID        string `json:"id"`
			Name      string `json:"name"`
			Arguments string `json:"arguments"`
		} `json:"toolCalls"`
	}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	choice := &llms.ContentChoice{
		Content: result.Content,
	}
	if len(result.ToolCalls) > 0 {
		choice.ToolCalls = make([]llms.ToolCall, 0, len(result.ToolCalls))
		for _, tc := range result.ToolCalls {
			choice.ToolCalls = append(choice.ToolCalls, llms.ToolCall{
				ID: tc.ID,
				FunctionCall: &llms.FunctionCall{
					Name:      tc.Name,
					Arguments: tc.Arguments,
				},
			})
		}
	}

	return &llms.ContentResponse{Choices: []*llms.ContentChoice{choice}}, nil
}

func convertToAPIMessages(messages []llms.MessageContent) ([]map[string]interface{}, error) {
	out := make([]map[string]interface{}, 0, len(messages))
	for _, mc := range messages {
		role := roleToAPI(mc.Role)
		msg := map[string]interface{}{"role": role}

		switch mc.Role {
		case llms.ChatMessageTypeTool:
			if len(mc.Parts) != 1 {
				continue
			}
			if tcr, ok := mc.Parts[0].(llms.ToolCallResponse); ok {
				msg["tool_call_id"] = tcr.ToolCallID
				msg["content"] = tcr.Content
			}
		case llms.ChatMessageTypeAI:
			var content string
			var toolCalls []map[string]interface{}
			for _, p := range mc.Parts {
				switch v := p.(type) {
				case llms.TextContent:
					content += v.Text
				case llms.ToolCall:
					if v.FunctionCall != nil {
						toolCalls = append(toolCalls, map[string]interface{}{
							"id":   v.ID,
							"type": "function",
							"function": map[string]interface{}{
								"name":      v.FunctionCall.Name,
								"arguments": v.FunctionCall.Arguments,
							},
						})
					}
				}
			}
			if content != "" {
				msg["content"] = content
			}
			if len(toolCalls) > 0 {
				msg["tool_calls"] = toolCalls
			}
		default:
			var content string
			for _, p := range mc.Parts {
				if tc, ok := p.(llms.TextContent); ok {
					content += tc.Text
				}
			}
			msg["content"] = content
		}
		out = append(out, msg)
	}
	return out, nil
}

func roleToAPI(r llms.ChatMessageType) string {
	switch r {
	case llms.ChatMessageTypeSystem:
		return "system"
	case llms.ChatMessageTypeHuman, llms.ChatMessageTypeGeneric:
		return "user"
	case llms.ChatMessageTypeAI:
		return "assistant"
	case llms.ChatMessageTypeTool, llms.ChatMessageTypeFunction:
		return "tool"
	default:
		return "user"
	}
}

func convertToolsToAPI(tools []llms.Tool) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(tools))
	for _, t := range tools {
		if t.Function == nil {
			continue
		}
		out = append(out, map[string]interface{}{
			"type": "function",
			"function": map[string]interface{}{
				"name":        t.Function.Name,
				"description": t.Function.Description,
				"parameters":  t.Function.Parameters,
			},
		})
	}
	return out
}
