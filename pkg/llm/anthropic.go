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

type AnthropicProvider struct {
	apiKey  string
	model   string
	client  *http.Client
}

func NewAnthropic(apiKey, model string) *AnthropicProvider {
	return &AnthropicProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{},
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic" }

type anthropicReq struct {
	Model     string          `json:"model"`
	MaxTokens int             `json:"max_tokens"`
	System    string          `json:"system,omitempty"`
	Messages  []anthropicMsg  `json:"messages"`
	Tools     []anthropicTool `json:"tools,omitempty"`
}

type anthropicMsg struct {
	Role    string            `json:"role"`
	Content []anthropicBlock  `json:"content"`
}

type anthropicBlock struct {
	Type  string `json:"type"`
	Text  string `json:"text,omitempty"`
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Input any    `json:"input,omitempty"`
}

type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	InputSchema json.RawMessage `json:"input_schema"`
}

type anthropicResp struct {
	Content []struct {
		Type  string `json:"type"`
		Text  string `json:"text,omitempty"`
		ID    string `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Input any    `json:"input,omitempty"`
	} `json:"content"`
	StopReason string `json:"stop_reason"`
}

func (p *AnthropicProvider) Chat(ctx context.Context, messages []schema.Message, tools []tool.Tool) (schema.Message, error) {
	var systemMsg string
	var apiMsgs []anthropicMsg

	for _, m := range messages {
		if m.Role == schema.RoleSystem {
			for _, b := range m.Content {
				if b.Type == "text" {
					systemMsg += b.Text + "\n"
				}
			}
			continue
		}

		var blocks []anthropicBlock
		for _, b := range m.Content {
			switch b.Type {
			case "text":
				blocks = append(blocks, anthropicBlock{Type: "text", Text: b.Text})
			case "tool_call":
				blocks = append(blocks, anthropicBlock{
					Type:  "tool_use",
					ID:    b.ToolCall.ID,
					Name:  b.ToolCall.Name,
					Input: b.ToolCall.Args,
				})
			case "tool_result":
				blocks = append(blocks, anthropicBlock{
					Type: "tool_result",
					ID:   b.ToolCall.ID,
					Text: fmt.Sprintf("%v", b.ToolCall.Result.Data),
				})
			}
		}

		role := string(m.Role)
		if role == "tool" {
			role = "user"
		}

		apiMsgs = append(apiMsgs, anthropicMsg{Role: role, Content: blocks})
	}

	anthropicTools := make([]anthropicTool, len(tools))
	for i, t := range tools {
		params := map[string]any{
			"type":       "object",
			"properties": make(map[string]any),
		}
		props := params["properties"].(map[string]any)
		for name, p := range t.Parameters {
			prop := map[string]any{"type": p.Type, "description": p.Description}
			if len(p.Enum) > 0 {
				prop["enum"] = p.Enum
			}
			props[name] = prop
		}
		paramJSON, _ := json.Marshal(params)

		anthropicTools[i] = anthropicTool{
			Name:        t.Name,
			Description: t.Description,
			InputSchema: paramJSON,
		}
	}

	reqBody := anthropicReq{
		Model:     p.model,
		MaxTokens: 4096,
		System:    systemMsg,
		Messages:  apiMsgs,
		Tools:     anthropicTools,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return schema.Message{}, fmt.Errorf("marshal: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(body))
	if err != nil {
		return schema.Message{}, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

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

	var result anthropicResp
	if err := json.Unmarshal(respBody, &result); err != nil {
		return schema.Message{}, fmt.Errorf("unmarshal: %w", err)
	}

	var blocks []schema.ContentBlock
	for _, c := range result.Content {
		switch c.Type {
		case "text":
			blocks = append(blocks, schema.ContentBlock{Type: "text", Text: c.Text})
		case "tool_use":
			input, _ := c.Input.(map[string]any)
			blocks = append(blocks, schema.ContentBlock{
				Type: "tool_call",
				ToolCall: &schema.ToolCall{
					ID:   c.ID,
					Name: c.Name,
					Args: input,
				},
			})
		}
	}

	return schema.Message{Role: schema.RoleAssistant, Content: blocks}, nil
}
