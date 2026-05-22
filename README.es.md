# Spider

[🇬🇧 English](README.md) · [🇪🇸 Español](README.es.md)

[![Versión Go](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![Licencia](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![PRs bienvenidos](https://img.shields.io/badge/PRs-welcome-brightgreen)]()

**Spider** es un harness multi-modelo para agentes de IA especializado en testing de software. Ejecuta un loop ReAct (Razonar → Actuar → Observar) con proveedores LLM, herramientas, memoria y seguridad con supervisión humana — todo desde la línea de comandos.

## Instalación

### Opción 1: npx (sin instalación, descarga automática)

```sh
npx spider-agent-harness run writer "genera tests para cmd/"
```

> Requiere Node.js 18+. El binario se descarga automáticamente al primer uso.

### Opción 2: Go install

```sh
go install github.com/user/spider@latest
```

### Opción 3: Descarga manual

Descarga el binario pre-compilado desde [Releases](https://github.com/user/spider/releases) para tu sistema operativo/arquitectura.

## Inicio rápido

```sh
# 1. Configuración interactiva (pide API key y proveedor)
spider init

# 2. Ejecuta un agente de testing
spider run writer "genera tests unitarios para el paquete pkg/agent"

# 3. Ve todos los agentes disponibles
spider list
```

## Los 5 agentes de testing

| Agente | Responsabilidad | Tools clave |
|---|---|---|
| `writer` | Genera archivos de test a partir del código fuente | `read_file`, `write_file`, `list_files` |
| `runner` | Ejecuta suites de test y reporta resultados | `bash` |
| `analyst` | Mide cobertura, detecta tests flaky | `bash`, `read_file` |
| `debugger` | Diagnostica fallos y propone correcciones | `bash`, `read_file` |
| `integrator` | Gestiona servicios externos, fixtures, E2E | `bash`, `read_file`, `list_files` |

Cada agente tiene un system prompt especializado y un conjunto de herramientas limitado. Sin solapamiento: uno genera, otro ejecuta, otro analiza, etc.

## Casos de uso

### "Añadir tests a un proyecto heredado sin cobertura"
```sh
spider run writer "genera tests para todos los archivos en pkg/"
spider run runner "ejecuta los tests generados"
```

### "Un test falla y no sé por qué"
```sh
spider run debugger "diagnostica el fallo en TestFoo"
```

### "Control de calidad antes de un release"
```sh
spider run analyst "analiza la cobertura de tests y detecta flakyness"
```

### "Necesito tests de integración para un microservicio"
```sh
spider run integrator "levanta postgres + redis, ejecuta contract tests"
```

## Soporte multi-modelo

Soportados de fábrica:

- **OpenAI** — GPT-4o, GPT-4, GPT-3.5
- **Anthropic** — Claude 3 Opus, Claude 3.5 Sonnet
- **Cualquier API compatible con OpenAI** — Ollama, Groq, Together AI, vLLM, etc.

### Ejemplos

```sh
# Ollama (local)
export SPIDER_API_BASE=http://localhost:11434/v1
export SPIDER_MODEL=llama3

# Groq
export SPIDER_API_BASE=https://api.groq.com/openai/v1
export SPIDER_MODEL=mixtral-8x7b-32768

# Proveedor personalizado (vía código)
import "spider/pkg/llm"
llm.RegisterProvider("personalizado", func(cfg schema.ProviderConfig) llm.Provider {
    // devuelve tu proveedor
})
```

## Seguridad (no negociable)

Tres capas obligatorias:

1. **Scope Enforcer** — Ninguna herramienta puede leer o escribir archivos fuera del directorio del proyecto. Symlinks, path traversal y rutas absolutas se resuelven y bloquean.

2. **Permission Checker** — `bash` y `write_file` requieren aprobación explícita por defecto. Se pueden añadir reglas por patrón (ej: denegar `rm -rf`).

3. **Human-in-the-Loop** — Antes de ejecutar una herramienta peligrosa, se pregunta al usuario. Las escrituras fuera del proyecto requieren **doble confirmación** con vista previa del diff.

## Sistema de memoria

Spider compacta automáticamente el historial de la conversación cuando supera el 75% del contexto del modelo. Los resúmenes se persisten como archivos `.md` en `~/.spider/memory/` y son buscables por los agentes.

```sh
ls ~/.spider/memory/
# index.json
# 2026-05-22_writer_3f7a8b2c.md
# 2026-05-22_runner_a1b2c3d4.md
```

Los agentes pueden consultar esta memoria usando la herramienta `memory_search` para recordar sesiones pasadas sin consumir el contexto completo de nuevo.

## Referencia de variables de entorno

| Variable | Default | Descripción |
|---|---|---|
| `OPENAI_API_KEY` | — | API key de OpenAI |
| `ANTHROPIC_API_KEY` | — | API key de Anthropic |
| `SPIDER_PROVIDER` | `openai` | Nombre del proveedor |
| `SPIDER_MODEL` | `gpt-4o` | ID del modelo |
| `SPIDER_API_BASE` | — | URL base (para APIs compatibles con OpenAI) |
| `SPIDER_MAX_ITERATIONS` | `25` | Máximo de pasos ReAct |
| `SPIDER_ALLOW_EXTERNAL` | `false` | Permitir escritura fuera del proyecto |
| `SPIDER_COMPACT_ENABLED` | `true` | Activar compactación de memoria |
| `SPIDER_CONTEXT_LIMIT` | `128000` | Tokens de contexto del modelo |
| `SPIDER_MEMORY_DIR` | `~/.spider/memory/` | Directorio de almacenamiento de memoria |

## Desarrollo

```sh
make build   # compilar
make test    # ejecutar tests
make clean   # limpiar binario

# Compilar sin Makefile
go build -o spider ./cmd/spider
```

## Referencia rápida de arquitectura

```
LLM Provider (OpenAI / Anthropic / Ollama …)
       │
  ┌────▼────┐
  │  Agent  │  Loop ReAct: pensar → actuar → observar
  │ Runtime │
  └────┬────┘
       │
  ┌────▼────┐    ┌──────────┐
  │  Tools  │    │  Memory  │
  │(scope + │    │(compactor│
  │ perms)  │    │ + store) │
  └─────────┘    └──────────┘
```

## FAQ

**¿Necesito una API key?** Sí. Ejecuta `spider init` para configurarla.

**¿Puedo usarlo sin conexión?** Sí, con una instancia local de Ollama.

**¿Puedo añadir mi propio agente?** Sí. Crea una factoría en `pkg/agents/` y regístrala en `main.go`.

**¿Es seguro para CI/CD?** Sí, con `SPIDER_ALLOW_EXTERNAL=false`.

## Licencia

MIT
