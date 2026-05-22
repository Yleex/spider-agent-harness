package agent

import (
	"context"
	"spider/pkg/llm"
	"spider/pkg/memory"
	"spider/pkg/permission"
	"spider/pkg/schema"
	"spider/pkg/scope"
	"spider/pkg/tool"
)

type Agent interface {
	Name() string
	SystemPrompt() string
	Tools() []tool.Tool
	Run(ctx context.Context, task string) (string, error)
}

type Config struct {
	Name          string
	SystemPrompt  string
	Provider      llm.Provider
	MaxIterations int
	AllowExternal bool
	ScopeVal      *scope.Validator
	PermCheck     *permission.Checker
	Approver      permission.Approver
}

type BaseAgent struct {
	config Config
	mem    memory.Memory
}

func NewBase(cfg Config) *BaseAgent {
	if cfg.MaxIterations <= 0 {
		cfg.MaxIterations = 25
	}
	return &BaseAgent{
		config: cfg,
		mem:    memory.NewBuffer(100),
	}
}

func (a *BaseAgent) Name() string { return a.config.Name }

func (a *BaseAgent) SystemPrompt() string { return a.config.SystemPrompt }

func (a *BaseAgent) Tools() []tool.Tool { return nil }

func (a *BaseAgent) Run(ctx context.Context, task string) (string, error) {
	a.mem.Clear()
	a.mem.Add(schema.NewTextMessage(schema.RoleSystem, a.config.SystemPrompt))
	a.mem.Add(schema.NewTextMessage(schema.RoleUser, task))

	tools := a.Tools()
	toolInstances := make([]tool.ToolInstance, len(tools))
	for i, t := range tools {
		toolInstances[i] = tool.ToolInstance{
			Spec:          t,
			ScopeVal:      a.config.ScopeVal,
			PermCheck:     a.config.PermCheck,
			Approver:      a.config.Approver,
			AllowExternal: a.config.AllowExternal,
		}
	}

	for iter := 0; iter < a.config.MaxIterations; iter++ {
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		default:
		}

		resp, err := a.config.Provider.Chat(ctx, a.mem.Messages(), tools)
		if err != nil {
			return "", err
		}

		a.mem.Add(resp)

		hasToolCalls := false
		for _, block := range resp.Content {
			if block.Type == "tool_call" {
				hasToolCalls = true
				tc := block.ToolCall

				var instance *tool.ToolInstance
				for _, ti := range toolInstances {
					if ti.Spec.Name == tc.Name {
						instance = &ti
						break
					}
				}

				var result schema.ToolCallResult
				if instance == nil {
					result = schema.ToolCallResult{
						Success: false,
						Error:   "tool not found: " + tc.Name,
					}
				} else {
					data, err := instance.Call(ctx, tc.Args)
					if err != nil {
						result = schema.ToolCallResult{
							Success: false,
							Error:   err.Error(),
						}
					} else {
						result = schema.ToolCallResult{Success: true, Data: data}
					}
				}

				a.mem.Add(schema.NewToolResultMessage(tc.ID, result))
			}
		}

		if !hasToolCalls {
			text := ""
			for _, block := range resp.Content {
				if block.Type == "text" {
					text += block.Text
				}
			}
			return text, nil
		}
	}

	return "", nil
}
