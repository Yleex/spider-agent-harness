package memory

import (
	"spider/pkg/schema"
	"sync"
)

type Buffer struct {
	buffer []schema.Message
	limit  int
	mu     sync.RWMutex
}

func NewBuffer(limit int) *Buffer {
	return &Buffer{
		buffer: make([]schema.Message, 0, limit),
		limit:  limit,
	}
}

func (b *Buffer) Add(msg schema.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buffer = append(b.buffer, msg)
	if b.limit > 0 && len(b.buffer) > b.limit {
		excess := len(b.buffer) - b.limit
		b.buffer = b.buffer[excess:]
	}
}

func (b *Buffer) Messages() []schema.Message {
	b.mu.RLock()
	defer b.mu.RUnlock()
	out := make([]schema.Message, len(b.buffer))
	copy(out, b.buffer)
	return out
}

func (b *Buffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buffer = b.buffer[:0]
}

func (b *Buffer) Len() int {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return len(b.buffer)
}

func (b *Buffer) Replace(msgs []schema.Message) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.buffer = make([]schema.Message, len(msgs))
	copy(b.buffer, msgs)
}
