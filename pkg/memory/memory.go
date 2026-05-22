package memory

import (
	"context"
	"spider/pkg/schema"
	"time"
)

type Memory interface {
	Add(msg schema.Message)
	Messages() []schema.Message
	Clear()
	Len() int
}

type CompactConfig struct {
	ContextLimit     int
	Threshold        float64
	ReserveExchanges int
}

type SummaryEntry struct {
	Date      time.Time `json:"date"`
	AgentName string    `json:"agent"`
	SessionID string    `json:"session"`
	Tags      []string  `json:"tags"`
	Summary   string    `json:"summary"`
	TokensSaved int    `json:"tokens_compressed"`
}

type Compactor interface {
	ShouldCompact(messages []schema.Message) bool
	Compact(ctx context.Context, messages []schema.Message) ([]schema.Message, *SummaryEntry, error)
}
