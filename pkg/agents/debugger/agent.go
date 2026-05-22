package debugger

import (
	"spider/pkg/agent"
	"spider/pkg/tool"
	"spider/pkg/tool/builtin"
)

const systemPrompt = `Eres un especialista en DEBUG de tests fallidos. Tu única responsabilidad es diagnosticar fallos y proponer correcciones.

Reglas:
- Reproduce el fallo para obtener el stack trace y mensaje de error
- Analiza el código fuente y el test para identificar la causa raíz
- Compara snapshots o salidas esperadas vs reales
- Propone fixes concretos (nunca apliques cambios sin aprobación)
- NO ejecutes suites completas, NO generes tests nuevos, NO midas cobertura
- Si el fix requiere modificar código fuente, muestra el diff exacto y pide aprobación`

func New(cfg agent.Config) agent.Agent {
	base := agent.NewBase(cfg)
	specs := []tool.Tool{
		builtin.BashTool(),
		builtin.ReadFileTool(),
	}
	return &debuggerAgent{BaseAgent: base, toolSpecs: specs}
}

type debuggerAgent struct {
	*agent.BaseAgent
	toolSpecs []tool.Tool
}

func (a *debuggerAgent) SystemPrompt() string { return systemPrompt }

func (a *debuggerAgent) Tools() []tool.Tool { return a.toolSpecs }
