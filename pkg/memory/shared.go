package memory

import (
	"fmt"
	"sync"
)

type SubTaskResult struct {
	SubTaskID string
	AgentName string
	Task      string
	Output    string
	Err       error
}

type SharedResultStore struct {
	mu   sync.RWMutex
	data map[string]SubTaskResult
}

func NewSharedResultStore() *SharedResultStore {
	return &SharedResultStore{
		data: make(map[string]SubTaskResult),
	}
}

func (s *SharedResultStore) Set(key string, val SubTaskResult) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
}

func (s *SharedResultStore) Get(key string) (SubTaskResult, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.data[key]
	return v, ok
}

func (s *SharedResultStore) GetAll() map[string]SubTaskResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]SubTaskResult, len(s.data))
	for k, v := range s.data {
		out[k] = v
	}
	return out
}

func (s *SharedResultStore) Format() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if len(s.data) == 0 {
		return "[Sin resultados de sub-agentes aún]"
	}
	var out string
	for id, r := range s.data {
		status := "ok"
		if r.Err != nil {
			status = fmt.Sprintf("error: %s", r.Err)
		}
		out += fmt.Sprintf("  • %s [%s] → %s\n", id, r.AgentName, status)
	}
	return out
}

func (s *SharedResultStore) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.data)
}

func (s *SharedResultStore) ListByAgent(agentName string) map[string]SubTaskResult {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string]SubTaskResult)
	for k, v := range s.data {
		if v.AgentName == agentName {
			out[k] = v
		}
	}
	return out
}
