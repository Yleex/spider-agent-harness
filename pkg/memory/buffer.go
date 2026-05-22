package memory

import "spider/pkg/schema"

type Buffer struct {
	buffer []schema.Message
	limit  int
}

func NewBuffer(limit int) *Buffer {
	return &Buffer{
		buffer: make([]schema.Message, 0, limit),
		limit:  limit,
	}
}

func (b *Buffer) Add(msg schema.Message) {
	b.buffer = append(b.buffer, msg)
	if b.limit > 0 && len(b.buffer) > b.limit {
		excess := len(b.buffer) - b.limit
		b.buffer = b.buffer[excess:]
	}
}

func (b *Buffer) Messages() []schema.Message {
	out := make([]schema.Message, len(b.buffer))
	copy(out, b.buffer)
	return out
}

func (b *Buffer) Clear() {
	b.buffer = b.buffer[:0]
}

func (b *Buffer) Len() int {
	return len(b.buffer)
}
