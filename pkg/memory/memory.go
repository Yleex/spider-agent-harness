package memory

import (
	"spider/pkg/schema"
)

type Memory interface {
	Add(msg schema.Message)
	Messages() []schema.Message
	Clear()
	Len() int
}
