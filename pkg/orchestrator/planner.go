package orchestrator

import (
	"context"
	"fmt"
	"spider/pkg/agent"
	"spider/pkg/memory"
	"spider/pkg/tool"
	"spider/pkg/tool/builtin"
	"strings"
)

const plannerPrompt = `Eres un ORQUESTADOR de agentes de testing. Tu trabajo es descomponer tareas complejas en subtareas que otros agentes especializados ejecutan.

Tienes acceso a estos agentes:
  • writer — genera tests a partir del código fuente
  • runner — ejecuta suites de test y reporta resultados
  • analyst — analiza cobertura, detecta flakyness
  • debugger — diagnostica fallos y propone fixes
  • integrator — servicios externos, fixtures, E2E

Para cada subtarea DEBES especificar:
  id: nombre único
  agent: writer | runner | analyst | debugger | integrator
  task: descripción clara de qué hacer
  depends: [ids de subtareas que deben completarse antes]
  strategy: aislado | secuencial | concurrente

Reglas:
  • Subtareas independientes pueden ejecutarse en paralelo (strategy=concurrente)
  • Si una subtarea necesita el resultado de otra, usa depends + strategy=secuencial
  • Si una subtarea es completamente independiente, usa strategy=aislado
  • Máximo 3 subtareas concurrentes
  • La ejecución en paralelo ahorra tiempo pero cada agente tiene herramientas propias y trabaja en su copia del proyecto

Al final, sintetiza los resultados de todas las subtareas en una respuesta coherente.`

type PlannerAgent struct {
	*agent.BaseAgent
	pool      *Pool
	store     *memory.SharedResultStore
}

func NewPlanner(cfg agent.Config, pool *Pool, store *memory.SharedResultStore) *PlannerAgent {
	base := agent.NewBase(cfg)
	return &PlannerAgent{
		BaseAgent: base,
		pool:      pool,
		store:     store,
	}
}

func (a *PlannerAgent) SystemPrompt() string { return plannerPrompt }

func (a *PlannerAgent) Tools() []tool.Tool {
	tools := []tool.Tool{
		builtin.ReadFileTool(),
		builtin.ListFilesTool(),
		builtin.BashTool(),
		{
			Name:        "run_subtasks",
			Description: "Ejecuta múltiples subtareas en paralelo según el plan definido. Args: plan (JSON con Steps)",
			Parameters: map[string]tool.ParameterSchema{
				"plan_json": {
					Type:        "string",
					Description: "JSON con el plan de subtareas a ejecutar. Array de objetos con id, agent, task, depends (opcional), strategy (opcional: aislado|secuencial|concurrente). Max 3 concurrentes.",
					Required:    true,
				},
			},
			PathArgs: nil,
			Execute: func(ctx context.Context, args map[string]any) (any, error) {
				planJSON, _ := args["plan_json"].(string)
				if planJSON == "" {
					return nil, fmt.Errorf("plan_json es requerido")
				}
				return a.executePlan(ctx, planJSON)
			},
		},
		{
			Name:        "get_results",
			Description: "Obtiene los resultados de subtareas ya ejecutadas. Filtra por agent si se especifica.",
			Parameters: map[string]tool.ParameterSchema{
				"agent": {
					Type:        "string",
					Description: "Nombre del agente para filtrar (opcional)",
					Required:    false,
				},
			},
			PathArgs: nil,
			Execute: func(ctx context.Context, args map[string]any) (any, error) {
				agentName, _ := args["agent"].(string)
				return a.getResults(agentName), nil
			},
		},
	}
	cfg := a.GetConfig()
	if cfg.MemoryStore != nil {
		tools = append(tools, builtin.MemorySearchTool(cfg.MemoryStore))
	}
	return tools
}

func (a *PlannerAgent) executePlan(ctx context.Context, planJSON string) (string, error) {
	var plan Plan
	if err := parsePlanJSON(planJSON, &plan); err != nil {
		return "", fmt.Errorf("error parseando plan: %w", err)
	}

	if len(plan.Steps) == 0 {
		return "", fmt.Errorf("el plan no contiene subtareas")
	}

	for i := range plan.Steps {
		if plan.Steps[i].Strategy == 0 {
			hasDeps := len(plan.Steps[i].DependsOn) > 0
			if hasDeps {
				plan.Steps[i].Strategy = StrategySecuencial
			} else {
				plan.Steps[i].Strategy = StrategyConcurrente
			}
		}
	}

	results, err := a.pool.Run(ctx, a.GetConfig(), plan, a.store)
	if err != nil {
		return "", fmt.Errorf("error ejecutando plan: %w", err)
	}

	var report strings.Builder
	report.WriteString("## Resultados del plan\n\n")
	for _, r := range results {
		status := "✓"
		if r.Err != nil {
			status = "✗"
		}
		report.WriteString(fmt.Sprintf("### %s (%s) %s\n", r.SubTask.ID, r.SubTask.AgentName, status))
		report.WriteString(fmt.Sprintf("Tarea: %s\n", r.SubTask.Task))
		if r.Err != nil {
			report.WriteString(fmt.Sprintf("Error: %s\n", r.Err))
		} else {
			report.WriteString(fmt.Sprintf("Resultado:\n%s\n", truncateOutput(r.Output, 2000)))
		}
		report.WriteString("\n")
	}
	report.WriteString(fmt.Sprintf("\nTotal: %d subtareas ejecutadas\n", len(results)))

	return report.String(), nil
}

func (a *PlannerAgent) getResults(agentName string) string {
	if a.store == nil {
		return "No hay resultados disponibles"
	}
	if agentName != "" {
		results := a.store.ListByAgent(agentName)
		if len(results) == 0 {
			return fmt.Sprintf("No hay resultados para el agente %s", agentName)
		}
		return formatResultsMap(results)
	}
	return a.store.Format()
}

func parsePlanJSON(jsonStr string, plan *Plan) error {
	lines := strings.Split(jsonStr, "\n")
	var current Subtask

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "```") {
			if current.ID != "" {
				plan.Add(current)
				current = Subtask{}
			}
			continue
		}

		if strings.HasPrefix(line, "-") || strings.HasPrefix(line, "*") {
			if current.ID != "" {
				plan.Add(current)
				current = Subtask{}
			}
			line = strings.TrimLeft(line, "-* ")
		}

		if strings.HasPrefix(strings.ToLower(line), "id:") {
			if current.ID != "" {
				plan.Add(current)
			}
			current = Subtask{ID: strings.TrimSpace(strings.TrimPrefix(line[3:], " "))}
		} else if strings.HasPrefix(strings.ToLower(line), "agent:") {
			agentName := strings.TrimSpace(strings.TrimPrefix(line[6:], " "))
			agentName = strings.TrimPrefix(agentName, "writer")
			agentName = strings.TrimSpace(agentName)
			current.AgentName = strings.TrimSpace(strings.TrimPrefix(line[6:], " "))
		} else if strings.HasPrefix(strings.ToLower(line), "task:") {
			current.Task = strings.TrimSpace(strings.TrimPrefix(line[5:], " "))
		} else if strings.HasPrefix(strings.ToLower(line), "depends:") {
			deps := strings.TrimSpace(strings.TrimPrefix(line[8:], " "))
			deps = strings.Trim(deps, "[]")
			for _, d := range strings.Split(deps, ",") {
				d = strings.TrimSpace(d)
				if d != "" {
					current.DependsOn = append(current.DependsOn, d)
				}
			}
		} else if strings.HasPrefix(strings.ToLower(line), "strategy:") {
			strat := strings.TrimSpace(strings.ToLower(strings.TrimPrefix(line[9:], " ")))
			switch strat {
			case "concurrente", "parallel":
				current.Strategy = StrategyConcurrente
			case "secuencial", "sequential":
				current.Strategy = StrategySecuencial
			default:
				current.Strategy = StrategyAislado
			}
		} else if current.Task != "" && !strings.HasPrefix(line, "#") {
			if current.Task != "" {
				current.Task += " " + line
			}
		}
	}

	if current.ID != "" {
		plan.Add(current)
	}

	return nil
}

func formatResultsMap(results map[string]memory.SubTaskResult) string {
	var out string
	for _, r := range results {
		status := "✓"
		if r.Err != nil {
			status = fmt.Sprintf("✗ %s", r.Err)
		}
		out += fmt.Sprintf("  • %s (%s) %s\n", r.SubTaskID, r.AgentName, status)
	}
	return out
}

func truncateOutput(s string, max int) string {
	runes := []rune(s)
	if len(runes) <= max {
		return s
	}
	return string(runes[:max]) + "...\n[output truncado]"
}

func (a *PlannerAgent) Run(ctx context.Context, task string) (string, error) {
	return a.BaseAgent.Run(ctx, task)
}
