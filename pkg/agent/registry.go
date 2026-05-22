package agent

import "fmt"

type Registry struct {
	agents map[string]func(Config) Agent
}

func NewRegistry() *Registry {
	return &Registry{agents: make(map[string]func(Config) Agent)}
}

func (r *Registry) Register(name string, factory func(Config) Agent) {
	r.agents[name] = factory
}

func (r *Registry) Create(name string, cfg Config) (Agent, error) {
	fn, ok := r.agents[name]
	if !ok {
		return nil, fmt.Errorf("agente no encontrado: %s", name)
	}
	return fn(cfg), nil
}

func (r *Registry) List() []string {
	var names []string
	for n := range r.agents {
		names = append(names, n)
	}
	return names
}
