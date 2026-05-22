package memory

import (
	"context"
	"fmt"
	"spider/pkg/llm"
	"spider/pkg/schema"
	"strings"
	"time"
)

type SummarizingCompactor struct {
	provider         llm.Provider
	cfg              CompactConfig
	sessionID        string
}

func NewCompactor(provider llm.Provider, cfg CompactConfig, sessionID string) *SummarizingCompactor {
	if cfg.ContextLimit <= 0 {
		cfg.ContextLimit = 128000
	}
	if cfg.Threshold <= 0 {
		cfg.Threshold = 0.75
	}
	if cfg.ReserveExchanges <= 0 {
		cfg.ReserveExchanges = 5
	}
	return &SummarizingCompactor{
		provider:  provider,
		cfg:       cfg,
		sessionID: sessionID,
	}
}

func estimateTokens(messages []schema.Message) int {
	var total int
	for _, m := range messages {
		for _, c := range m.Content {
			switch c.Type {
			case "text":
				total += len([]rune(c.Text)) * 2
			case "tool_call":
				total += len(c.ToolCall.Name) * 2
			case "tool_result":
				if c.ToolCall.Result != nil {
					switch v := c.ToolCall.Result.Data.(type) {
					case string:
						total += len([]rune(v)) * 2
					}
					total += len(c.ToolCall.Result.Error) * 2
				}
			}
		}
	}
	return total
}

func (c *SummarizingCompactor) ShouldCompact(messages []schema.Message) bool {
	if len(messages) < c.cfg.ReserveExchanges*2+3 {
		return false
	}
	estimated := estimateTokens(messages)
	limit := int(float64(c.cfg.ContextLimit) * c.cfg.Threshold)
	return estimated > limit
}

func (c *SummarizingCompactor) Compact(ctx context.Context, messages []schema.Message) ([]schema.Message, *SummaryEntry, error) {
	if len(messages) < 2 {
		return messages, nil, nil
	}

	systemMsg := messages[0]

	preserveCount := c.cfg.ReserveExchanges * 2
	if preserveCount >= len(messages)-1 {
		preserveCount = len(messages) - 2
		if preserveCount < 0 {
			preserveCount = 0
		}
	}

	preserved := messages[len(messages)-preserveCount:]
	toSummarize := messages[1 : len(messages)-preserveCount]

	if len(toSummarize) == 0 {
		return messages, nil, nil
	}

	tokensBefore := estimateTokens(toSummarize)

	var summaryText strings.Builder
	summaryText.WriteString("Historial comprimido de la conversación hasta este punto:\n\n")
	for _, m := range toSummarize {
		for _, c := range m.Content {
			switch c.Type {
			case "text":
				summaryText.WriteString(fmt.Sprintf("[%s] %s\n", m.Role, truncate(c.Text, 500)))
			case "tool_call":
				summaryText.WriteString(fmt.Sprintf("[%s llamó a tool %s con args=%v]\n", m.Role, c.ToolCall.Name, c.ToolCall.Args))
			case "tool_result":
				status := "ok"
				if c.ToolCall.Result != nil && !c.ToolCall.Result.Success {
					status = "error: " + c.ToolCall.Result.Error
				}
				summaryText.WriteString(fmt.Sprintf("[resultado de tool: %s]\n", status))
			}
		}
	}

	prompt := fmt.Sprintf(`Eres un compresor de memoria. Tu tarea es resumir el siguiente historial de conversación de un agente de IA.

Debes conservar TODA la información IMPORTANTE:
- Decisiones tomadas y por qué
- Archivos creados o modificados
- Errores encontrados y soluciones
- Datos concretos (nombres, rutas, valores)
- Tareas pendientes o bloqueos

Sé conciso pero completo. El resumen será leído por el agente para recordar el contexto.

Historial a resumir:

%s`, summaryText.String())

	resp, err := c.provider.Chat(ctx, []schema.Message{
		schema.NewTextMessage(schema.RoleSystem, "Eres un compresor de memoria eficiente y preciso."),
		schema.NewTextMessage(schema.RoleUser, prompt),
	}, nil)
	if err != nil {
		return nil, nil, fmt.Errorf("compactando memoria: %w", err)
	}

	var summary string
	for _, b := range resp.Content {
		if b.Type == "text" {
			summary += b.Text
		}
	}

	summaryMsg := schema.NewTextMessage(schema.RoleSystem, fmt.Sprintf(
		"[MEMORIA COMPRIMIDA — sesión %s]\n\n%s", c.sessionID, summary))

	tokensAfter := estimateTokens([]schema.Message{summaryMsg})
	tokensSaved := tokensBefore - tokensAfter
	if tokensSaved < 0 {
		tokensSaved = 0
	}

	entry := &SummaryEntry{
		Date:       time.Now(),
		SessionID:  c.sessionID,
		Summary:    summary,
		TokensSaved: tokensSaved,
	}

	result := make([]schema.Message, 0, len(preserved)+2)
	result = append(result, systemMsg)
	result = append(result, summaryMsg)
	result = append(result, preserved...)

	return result, entry, nil
}

func truncate(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "..."
}
