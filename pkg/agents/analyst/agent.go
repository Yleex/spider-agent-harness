package analyst

import (
	"spider/pkg/agent"
	"spider/pkg/tool"
	"spider/pkg/tool/builtin"
)

const systemPrompt = `Eres un especialista en ANALIZAR calidad de tests. Tu única responsabilidad es medir y reportar métricas de calidad.

Reglas:
- Mide cobertura de código (líneas, ramas, funciones)
- Detecta tests flaky analizando histórico de ejecuciones
- Identifica tests huérfanos, redundantes o mal escritos
- Genera reportes claros con recomendaciones
- NO ejecutes tests, NO generes tests nuevos, NO diagnostiques fallos
- Trabaja sobre datos ya existentes o ejecuciones previas`

func New(cfg agent.Config) agent.Agent {
	base := agent.NewBase(cfg)
	specs := []tool.Tool{
		builtin.ReadFileTool(),
		builtin.ListFilesTool(),
		builtin.BashTool(),
	}
	if cfg.MemoryStore != nil {
		specs = append(specs, builtin.MemorySearchTool(cfg.MemoryStore))
		specs = append(specs, builtin.MemorySaveTool(cfg.MemoryStore))
	}
	return &analystAgent{BaseAgent: base, toolSpecs: specs}
}

type analystAgent struct {
	*agent.BaseAgent
	toolSpecs []tool.Tool
}

func (a *analystAgent) SystemPrompt() string { return systemPrompt }

func (a *analystAgent) Tools() []tool.Tool { return a.toolSpecs }
