package orchestrator

import (
	"context"
	"fmt"
	"spider/pkg/agent"
	"spider/pkg/memory"
	"sync"
)

type Pool struct {
	sem    chan struct{}
	agents map[string]func(agent.Config) agent.Agent
}

type subtaskOutcome struct {
	SubTask Subtask
	Output  string
	Err     error
}

func NewPool(maxConcurrent int) *Pool {
	return &Pool{
		sem:    make(chan struct{}, maxConcurrent),
		agents: make(map[string]func(agent.Config) agent.Agent),
	}
}

func (p *Pool) Register(name string, factory func(agent.Config) agent.Agent) {
	p.agents[name] = factory
}

func (p *Pool) Run(ctx context.Context, baseCfg agent.Config, plan Plan, store *memory.SharedResultStore) ([]subtaskOutcome, error) {
	groups := plan.ConcurrentGroups()
	var allResults []subtaskOutcome

	prevResult := ""

	for _, group := range groups {
		if len(group) == 0 {
			continue
		}

		var wg sync.WaitGroup
		groupResults := make([]subtaskOutcome, len(group))

		for sIdx, sub := range group {
			wg.Add(1)
			p.sem <- struct{}{}

			prev := prevResult
			go func(s Subtask, idx int) {
				defer wg.Done()
				defer func() { <-p.sem }()

				out := p.runSubTask(ctx, baseCfg, s, store, prev)
				groupResults[idx] = out

				if store != nil {
					store.Set(s.ID, memory.SubTaskResult{
						SubTaskID: s.ID,
						AgentName: s.AgentName,
						Task:      s.Task,
						Output:    out.Output,
						Err:       out.Err,
					})
				}
			}(sub, sIdx)
		}

		wg.Wait()

		for _, r := range groupResults {
			allResults = append(allResults, r)
			if r.Err == nil && r.Output != "" {
				prevResult = r.Output
			}
		}
	}

	return allResults, nil
}

func (p *Pool) runSubTask(ctx context.Context, baseCfg agent.Config, sub Subtask, store *memory.SharedResultStore, prevResult string) subtaskOutcome {
	factory, ok := p.agents[sub.AgentName]
	if !ok {
		return subtaskOutcome{SubTask: sub, Err: fmt.Errorf("agente no encontrado: %s", sub.AgentName)}
	}

	subCfg := baseCfg.Clone()
	subCfg.Name = sub.AgentName

	switch sub.Strategy {
	case StrategyAislado:
		subCfg.SharedStore = nil
		subCfg.InjectedContext = ""
	case StrategySecuencial:
		subCfg.SharedStore = store
		subCfg.InjectedContext = prevResult
	case StrategyConcurrente:
		subCfg.SharedStore = store
		subCfg.InjectedContext = ""
	}

	a := factory(subCfg)
	out, err := a.Run(ctx, sub.Task)
	return subtaskOutcome{SubTask: sub, Output: out, Err: err}
}


