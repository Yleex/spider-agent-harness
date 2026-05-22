package writer

import (
	"spider/pkg/agent"
	"spider/pkg/tool"
	"spider/pkg/tool/builtin"
)

const systemPrompt = `Eres un especialista en GENERAR tests. Tu única responsabilidad es crear archivos de test.

Reglas:
- Analiza el código fuente antes de escribir cualquier test
- Genera tests unitarios, de tabla, y mocks cuando sea necesario
- Cubre casos felices, casos borde y errores
- Sigue las convenciones del lenguaje y del proyecto
- NO ejecutes tests, NO analices cobertura, NO diagnostiques fallos
- Si encuentras un bug en el código fuente, documéntalo como comentario en el test`

func New(cfg agent.Config) agent.Agent {
	base := agent.NewBase(cfg)
	specs := []tool.Tool{
		builtin.ReadFileTool(),
		builtin.ListFilesTool(),
		builtin.WriteFileTool(),
	}
	return &writerAgent{BaseAgent: base, toolSpecs: specs}
}

type writerAgent struct {
	*agent.BaseAgent
	toolSpecs []tool.Tool
}

func (a *writerAgent) SystemPrompt() string { return systemPrompt }

func (a *writerAgent) Tools() []tool.Tool { return a.toolSpecs }
