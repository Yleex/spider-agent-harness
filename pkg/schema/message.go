package schema

type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

type ToolCall struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Args   map[string]any    `json:"args"`
	Result *ToolCallResult   `json:"result,omitempty"`
}

type ToolCallResult struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   string `json:"error,omitempty"`
}

type ContentBlock struct {
	Type     string         `json:"type"`
	Text     string         `json:"text,omitempty"`
	ToolCall *ToolCall       `json:"tool_call,omitempty"`
}

type Message struct {
	Role    Role            `json:"role"`
	Content []ContentBlock  `json:"content"`
}

func NewTextMessage(role Role, text string) Message {
	return Message{
		Role:    role,
		Content: []ContentBlock{{Type: "text", Text: text}},
	}
}

func NewToolCallMessage(id, name string, args map[string]any) Message {
	return Message{
		Role: RoleAssistant,
		Content: []ContentBlock{{
			Type: "tool_call",
			ToolCall: &ToolCall{ID: id, Name: name, Args: args},
		}},
	}
}

func NewToolResultMessage(id string, result ToolCallResult) Message {
	msg := Message{
		Role: RoleTool,
		Content: []ContentBlock{{
			Type: "tool_result",
			ToolCall: &ToolCall{ID: id, Result: &result},
		}},
	}
	return msg
}
