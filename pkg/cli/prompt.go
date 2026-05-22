package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

type Answer struct {
	Value string
}

func Prompt(msg string) string {
	fmt.Print("  " + Cyan("?") + " " + msg + " ")
	reader := bufio.NewReader(os.Stdin)
	line, _ := reader.ReadString('\n')
	return strings.TrimRight(line, "\r\n")
}

func PromptDefault(msg, def string) string {
	val := Prompt(msg + " [" + Bold(def) + "]")
	if val == "" {
		return def
	}
	return val
}

func Confirm(msg string) bool {
	resp := Prompt(msg + " (y/N)")
	resp = strings.TrimSpace(strings.ToLower(resp))
	return resp == "y" || resp == "yes"
}

type Option struct {
	Label       string
	Description string
}

func Select(msg string, options []Option) string {
	fmt.Println("  " + Cyan("?") + " " + msg)
	for i, opt := range options {
		fmt.Printf("    %d) %s — %s\n", i+1, Bold(opt.Label), opt.Description)
	}

	for {
		line := Prompt("Elije 1-" + fmt.Sprintf("%d", len(options)) + ":")
		idx := 0
		for i := range line {
			idx = idx*10 + int(line[i]-'0')
		}
		if idx >= 1 && idx <= len(options) {
			return options[idx-1].Label
		}
	}
}

func RunInit() {
	Banner()
	Section("Configuración inicial de Spider")

	provider := Select("¿Qué proveedor de LLM quieres usar?", []Option{
		{Label: "openai", Description: "OpenAI (GPT-4o, GPT-4, GPT-3.5) — recomendado"},
		{Label: "anthropic", Description: "Anthropic (Claude 3, Claude 3.5)"},
		{Label: "openai-compatible", Description: "Cualquier API compatible con OpenAI (Ollama, Groq, Together, vLLM, etc.)"},
	})

	var apiKey, apiBase, model string

	switch provider {
	case "openai":
		apiKey = Prompt("OPENAI_API_KEY:")
		model = PromptDefault("Modelo", "gpt-4o")
	case "anthropic":
		apiKey = Prompt("ANTHROPIC_API_KEY:")
		model = PromptDefault("Modelo", "claude-sonnet-4-20250514")
	case "openai-compatible":
		apiBase = Prompt("API Base URL (ej: http://localhost:11434/v1 para Ollama):")
		apiKey = PromptDefault("API Key (opcional, dejar vacío si no requiere)", "")
		model = PromptDefault("Modelo", "llama3")
		provider = "openai"
	}

	content := "# Configuración generada por spider init\n"
	content += fmt.Sprintf("SPIDER_PROVIDER=%s\n", provider)
	content += fmt.Sprintf("SPIDER_MODEL=%s\n", model)
	if apiKey != "" {
		content += fmt.Sprintf("OPENAI_API_KEY=%s\n", apiKey)
		if provider != "openai" {
			content += fmt.Sprintf("ANTHROPIC_API_KEY=%s\n", apiKey)
		}
	}
	if apiBase != "" {
		content += fmt.Sprintf("SPIDER_API_BASE=%s\n", apiBase)
	}

	os.WriteFile(".env", []byte(content), 0644)
	Done("Archivo .env creado en " + Bold(".env"))
	Info("Ya puedes usar: " + Cyan("spider run writer \"genera tests para mi paquete\""))
}

func RunList() {
	Banner()
	Section("Agentes disponibles")

	type agentInfo struct {
		name string
		desc string
	}

	agents := []agentInfo{
		{"writer", "Genera tests unitarios y de integración a partir del código fuente"},
		{"runner", "Ejecuta suites de test y reporta resultados crudos"},
		{"analyst", "Analiza cobertura, detecta flakyness y genera reportes de calidad"},
		{"debugger", "Diagnostica test failures y propone correcciones"},
		{"integrator", "Maneja servicios externos, fixtures y tests de integración/E2E"},
		{"planner", "Orquestador: descompone tareas complejas y las ejecuta en paralelo vía sub-agentes"},
	}

	for _, a := range agents {
		fmt.Printf("  %s  %s\n", Cyan(a.name), Bold(a.desc))
	}

	fmt.Println()
	Info("Usa: " + Cyan("spider run <agente> \"<tarea>\""))
}

func RunHelp() {
	Banner()
	Section("Cómo usar Spider")

	fmt.Println(`
  Comandos:

    spider init             Configura API key, proveedor y modelo
    spider list             Muestra los agentes disponibles
    spider run <a> "<t>"    Ejecuta un agente con una tarea
    spider help             Muestra esta ayuda

  Ejemplos:

    spider run writer "genera tests para el paquete pkg/agent"
    spider run runner "ejecuta los tests del proyecto"
    spider run analyst "analiza la cobertura de tests"
    spider run debugger "diagnostica por qué falla TestRun"
    spider run integrator "prepara servicios para test de integración"
    spider run planner "asegura la calidad del proyecto antes del release"

  Agente planner:

    El agente planner descompone tareas complejas en subtareas independientes
    y las ejecuta en paralelo usando goroutines. Soporta 3 estrategias:
      • concurrente — subtareas independientes se ejecutan en paralelo
      • secuencial — cada subtarea hereda el resultado de la anterior
      • aislado — cada subtarea corre en su propia sesión sin compartir contexto

  Variables de entorno:

    OPENAI_API_KEY           API key de OpenAI
    ANTHROPIC_API_KEY        API key de Anthropic
    SPIDER_PROVIDER          openai | anthropic (default: openai)
    SPIDER_MODEL             Modelo a usar
    SPIDER_API_BASE          URL base para APIs compatibles con OpenAI
    SPIDER_MAX_ITERATIONS    Pasos máximos del agente (default: 25)
    SPIDER_ALLOW_EXTERNAL    Permitir escritura fuera del proyecto (default: false)
    SPIDER_MEMORY_DIR        Directorio de memorias persistentes

  Consejos:

    • Si no tienes API key, ejecuta 'spider init' para configurarlo
    • Para usar Ollama local: SPIDER_API_BASE=http://localhost:11434/v1
    • Configura SPIDER_ALLOW_EXTERNAL=true con precaución
    • Las memorias de sesiones se guardan en ~/.spider/memory/`)
}
