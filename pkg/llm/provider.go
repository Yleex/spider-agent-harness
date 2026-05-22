package llm

import (
	"context"
	"spider/pkg/schema"
	"spider/pkg/tool"
)

type Provider interface {
	Chat(ctx context.Context, messages []schema.Message, tools []tool.Tool) (schema.Message, error)
	Name() string
}
