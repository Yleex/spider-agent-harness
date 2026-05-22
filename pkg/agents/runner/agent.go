package runner

import (
	"spider/pkg/agent"
	"spider/pkg/tool"
	"spider/pkg/tool/builtin"
)

const systemPrompt = `Eres un especialista en EJECUTAR tests. Tu única responsabilidad es correr suites de test y reportar resultados crudos.

Reglas:
- Ejecuta tests en los paquetes o archivos indicados
- Reporta resultados: pasados, fallidos, errores de compilación
- Maneja paralelización cuando sea seguro
- NO generes tests, NO analices cobertura, NO diagnostiques fallos
- NO modifiques código fuente ni archivos de test`

func New(cfg agent.Config) agent.Agent {
	base := agent.NewBase(cfg)
	specs := []tool.Tool{
		builtin.BashTool(),
		builtin.ReadFileTool(),
	}
	if cfg.MemoryStore != nil {
		specs = append(specs, builtin.MemorySearchTool(cfg.MemoryStore))
		specs = append(specs, builtin.MemorySaveTool(cfg.MemoryStore))
	}
	return &runnerAgent{BaseAgent: base, toolSpecs: specs}
}

type runnerAgent struct {
	*agent.BaseAgent
	toolSpecs []tool.Tool
}

func (a *runnerAgent) SystemPrompt() string { return systemPrompt }

func (a *runnerAgent) Tools() []tool.Tool { return a.toolSpecs }
