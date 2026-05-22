package orchestrator

type Strategy int

const (
	StrategyAislado      Strategy = iota
	StrategySecuencial
	StrategyConcurrente
)

type Subtask struct {
	ID          string
	AgentName   string
	Task        string
	DependsOn   []string
	Strategy    Strategy
}

type Plan struct {
	Steps []Subtask
}

func NewPlan() *Plan {
	return &Plan{}
}

func (p *Plan) Add(s Subtask) {
	p.Steps = append(p.Steps, s)
}

func (p *Plan) ConcurrentGroups() [][]Subtask {
	done := map[string]bool{}
	var groups [][]Subtask

	remaining := make([]Subtask, len(p.Steps))
	copy(remaining, p.Steps)

	for len(remaining) > 0 {
		var ready []Subtask
		var next []Subtask

		for _, s := range remaining {
			depsMet := true
			for _, dep := range s.DependsOn {
				if !done[dep] {
					depsMet = false
					break
				}
			}
			if depsMet {
				ready = append(ready, s)
			} else {
				next = append(next, s)
			}
		}

		if len(ready) == 0 && len(next) > 0 {
			ready = next[:1]
			next = next[1:]
		}

		for _, s := range ready {
			done[s.ID] = true
		}
		groups = append(groups, ready)
		remaining = next
	}

	return groups
}
