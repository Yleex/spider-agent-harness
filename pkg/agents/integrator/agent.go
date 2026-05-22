package integrator

import (
	"spider/pkg/agent"
	"spider/pkg/tool"
	"spider/pkg/tool/builtin"
)

const systemPrompt = `Eres un especialista en tests de INTEGRACIÓN y E2E. Tu única responsabilidad es manejar servicios externos, fixtures y tests de integración.

Reglas:
- Levanta y administra servicios externos (bases de datos, APIs, contenedores)
- Gestiona fixtures y datos de prueba
- Ejecuta contract tests contra APIs externas
- Verifica conectividad y disponibilidad de servicios
- NO generes tests unitarios, NO analices cobertura, NO diagnostiques fallos unitarios
- Reporta claramente el estado de cada integración`

func New(cfg agent.Config) agent.Agent {
	base := agent.NewBase(cfg)
	specs := []tool.Tool{
		builtin.BashTool(),
		builtin.ReadFileTool(),
		builtin.ListFilesTool(),
	}
	if cfg.MemoryStore != nil {
		specs = append(specs, builtin.MemorySearchTool(cfg.MemoryStore))
		specs = append(specs, builtin.MemorySaveTool(cfg.MemoryStore))
	}
	return &integratorAgent{BaseAgent: base, toolSpecs: specs}
}

type integratorAgent struct {
	*agent.BaseAgent
	toolSpecs []tool.Tool
}

func (a *integratorAgent) SystemPrompt() string { return systemPrompt }

func (a *integratorAgent) Tools() []tool.Tool { return a.toolSpecs }
