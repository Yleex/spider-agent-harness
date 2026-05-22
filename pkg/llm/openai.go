package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"spider/pkg/schema"
	"spider/pkg/tool"
)

type OpenAIProvider struct {
	apiKey  string
	model   string
	client  *http.Client
}

func NewOpenAI(apiKey, model string) *OpenAIProvider {
	return &OpenAIProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

type openAIReq struct {
	Model    string            `json:"model"`
	Messages []openAIMsg       `json:"messages"`
	Tools    []openAITool      `json:"tools,omitempty"`
	Temp     float64           `json:"temperature"`
}

type openAIMsg struct {
	Role    string          `json:"role"`
	Content json.RawMessage `json:"content"`
}

type openAITool struct {
	Type     string         `json:"type"`
	Function openAIFunction `json:"function"`
}

type openAIFunction struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type openAIResp struct {
	Choices []struct {
		Message struct {
			Role         string            `json:"role"`
			Content      *string           `json:"content"`
			ToolCalls    []openAIToolCall  `json:"tool_calls"`
		} `json:"message"`
	} `json:"choices"`
}

type openAIToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"`
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"`
	} `json:"function"`
}

func (p *OpenAIProvider) Name() string { return "openai" }

func (p *OpenAIProvider) Chat(ctx context.Context, messages []schema.Message, tools []tool.Tool) (schema.Message, error) {
	openAIMsgs := make([]openAIMsg, len(messages))
	for i, m := range messages {
		var content json.RawMessage
		var parts []map[string]any

		for _, block := range m.Content {
			switch block.Type {
			case "text":
				parts = append(parts, map[string]any{
					"type": "text",
					"text": block.Text,
				})
			case "tool_call":
				args, _ := json.Marshal(block.ToolCall.Args)
				parts = append(parts, map[string]any{
					"type": "function",
					"id":   block.ToolCall.ID,
					"function": map[string]any{
						"name":      block.ToolCall.Name,
						"arguments": string(args),
					},
				})
			case "tool_result":
				data, _ := json.Marshal(block.ToolCall.Result.Data)
				parts = append(parts, map[string]any{
					"type":       "tool_result",
					"tool_call_id": block.ToolCall.ID,
					"content":    string(data),
				})
			}
		}

		if len(parts) == 1 && parts[0]["type"] == "text" {
			content, _ = json.Marshal(parts[0]["text"])
		} else if len(parts) > 0 {
			content, _ = json.Marshal(parts)
		} else {
			content, _ = json.Marshal("")
		}

		openAIMsgs[i] = openAIMsg{Role: string(m.Role), Content: content}
	}

	openAITools := make([]openAITool, len(tools))
	for i, t := range tools {
		params := map[string]any{
			"type":       "object",
			"properties": make(map[string]any),
			"required":   []string{},
		}

		props := params["properties"].(map[string]any)
		var required []string
		for name, p := range t.Parameters {
			prop := map[string]any{"type": p.Type, "description": p.Description}
			if len(p.Enum) > 0 {
				prop["enum"] = p.Enum
			}
			props[name] = prop
			if p.Required {
				required = append(required, name)
			}
		}
		params["required"] = required
		paramJSON, _ := json.Marshal(params)

		openAITools[i] = openAITool{
			Type: "function",
			Function: openAIFunction{
				Name:        t.Name,
				Description: t.Description,
				Parameters:  paramJSON,
			},
		}
	}

	reqBody := openAIReq{
		Model:    p.model,
		Messages: openAIMsgs,
		Tools:    openAITools,
		Temp:     0.7,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return schema.Message{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/chat/completions", bytes.NewReader(body))
	if err != nil {
		return schema.Message{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return schema.Message{}, fmt.Errorf("request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return schema.Message{}, fmt.Errorf("reading response: %w", err)
	}

	if resp.StatusCode != 200 {
		return schema.Message{}, fmt.Errorf("API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result openAIResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return schema.Message{}, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(result.Choices) == 0 {
		return schema.Message{}, fmt.Errorf("no choices in response")
	}

	choice := result.Choices[0].Message

	var blocks []schema.ContentBlock

	if choice.Content != nil && *choice.Content != "" {
		blocks = append(blocks, schema.ContentBlock{Type: "text", Text: *choice.Content})
	}

	for _, tc := range choice.ToolCalls {
		var args map[string]any
		json.Unmarshal([]byte(tc.Function.Arguments), &args)

		blocks = append(blocks, schema.ContentBlock{
			Type: "tool_call",
			ToolCall: &schema.ToolCall{
				ID:   tc.ID,
				Name: tc.Function.Name,
				Args: args,
			},
		})
	}

	return schema.Message{Role: schema.RoleAssistant, Content: blocks}, nil
}
