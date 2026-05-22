# Spider

[🇬🇧 English](README.md) · [🇪🇸 Español](README.es.md)

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![Licencia](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**Spider es tu equipo de testing con IA. Decile qué querés probar y él escribe, ejecuta y corrige tests automáticamente.**

## ¿Qué es Spider?

- **Te ahorra horas** — en lugar de escribir tests a mano, describí lo que necesitás en español y Spider hace el trabajo
- **No necesitas saber de IA** — funciona con ChatGPT, Claude, o modelos locales gratuitos como Ollama. Solo necesitás una API key
- **Tu proyecto está seguro** — Spider pide permiso antes de ejecutar comandos o modificar archivos. Nada pasa sin tu aprobación

## Inicio rápido

```sh
# 1. Configuración única (necesitás una API key de OpenAI o Anthropic)
npx spider-agent-harness init

# 2. Pedile a Spider que escriba tests para tu código
npx spider-agent-harness run writer "genera tests para pkg/agent"

# 3. Vê todos los asistentes disponibles
npx spider-agent-harness list
```

## ¿Qué podés hacer con Spider?

### Tu proyecto no tiene ni un test
```sh
spider run writer "genera tests para todos los archivos en pkg/"
spider run runner "ejecuta los tests generados"
```

### Un test falla y no sabés por qué
```sh
spider run debugger "diagnostica el fallo en TestFoo"
```

### Necesitás un informe de calidad antes de un release
```sh
spider run analyst "analiza la cobertura de tests y detecta tests flaky"
```

### Querés automatizar todo de principio a fin
```sh
spider run planner "genera tests para todo el proyecto, ejecutalos, analiza cobertura y corrige los fallos"
```

### Necesitás tests que usen base de datos o APIs externas
```sh
spider run integrator "levanta postgres + redis, ejecuta tests de integración"
```

## Cómo funciona

1. **Describís una tarea en español** — Spider entiende lo que necesitás
2. **Spider planifica y ejecuta** — lee tu código, escribe tests, los ejecuta y analiza los resultados
3. **Trabaja como un equipo de especialistas** — cada uno con un rol específico (writer, runner, debugger, etc.)
4. **Vos tenés el control** — Spider pide permiso antes de ejecutar comandos o modificar archivos

## Los 6 asistentes

| Asistente | Qué hace |
|---|---|
| **writer** | Lee tu código y escribe tests |
| **runner** | Ejecuta los tests y te dice si pasan o fallan |
| **analyst** | Revisa qué porcentaje de tu código está cubierto por tests |
| **debugger** | Investiga por qué falla un test y sugiere cómo arreglarlo |
| **integrator** | Prepara bases de datos, APIs y servicios externos para testing |
| **planner** | El líder del equipo — divide tareas grandes en pasos más chicos y coordina a los demás |

## Instalación

### La más fácil: npx (sin instalación)

```sh
npx spider-agent-harness run writer "genera tests para cmd/"
```

> Requiere Node.js 18+. El binario se descarga automáticamente al primer uso.

### Alternativa: Go install

```sh
go install github.com/Yleex/spider-agent-harness@latest
```

> Requiere Go 1.22+. Se instala como `spider` en tu `$GOPATH/bin`.

### Alternativa: Descarga manual

Descargá el binario desde la [página de Releases](https://github.com/Yleex/spider-agent-harness/releases) para tu sistema operativo.

## FAQ

**¿Necesito una API key?** Sí. Ejecutá `spider init` para configurar una (OpenAI o Anthropic).

**¿Puedo usarlo sin conexión?** Sí, con un modelo local como Ollama.

**¿Es seguro?** Sí. Spider solo puede acceder a archivos dentro de tu proyecto. Pide permiso antes de ejecutar comandos o escribir archivos.

**¿Puede ejecutar tests en paralelo?** Sí. El asistente `planner` ejecuta automáticamente tareas independientes al mismo tiempo para ahorrar tiempo.

**¿Puedo crear mi propio asistente?** Sí. Si sos desarrollador, podés crear asistentes personalizados. Vê la sección de desarrollo más abajo.

### Soporte multi-modelo

Spider funciona con cualquier proveedor compatible con el formato de API de OpenAI:

- **OpenAI** — GPT-4o, GPT-4, GPT-3.5
- **Anthropic** — Claude 3 Opus, Claude 3.5 Sonnet
- **Ollama** (local) — `export SPIDER_API_BASE=http://localhost:11434/v1`
- **Groq** — `export SPIDER_API_BASE=https://api.groq.com/openai/v1`
- Cualquier API compatible con OpenAI

Agregá proveedores personalizados con `llm.RegisterProvider()` en `pkg/llm/registry.go`.

### Seguridad

Tres capas que no se pueden desactivar:

1. **Scope Enforcer** — Ninguna herramienta puede leer o escribir archivos fuera del directorio del proyecto. Symlinks y path traversal se resuelven y bloquean.
2. **Permission Checker** — `bash` y `write_file` requieren aprobación explícita por defecto. Se pueden agregar reglas por patrón (ej: denegar `rm -rf`).
3. **Human-in-the-Loop** — Antes de ejecutar una herramienta peligrosa, se pregunta al usuario. Las escrituras fuera del proyecto requieren doble confirmación.

### Variables de entorno

| Variable | Default | Descripción |
|---|---|---|
| `OPENAI_API_KEY` | — | API key de OpenAI |
| `ANTHROPIC_API_KEY` | — | API key de Anthropic |
| `SPIDER_PROVIDER` | `openai` | Nombre del proveedor |
| `SPIDER_MODEL` | `gpt-4o` | ID del modelo |
| `SPIDER_API_BASE` | — | URL base (para APIs compatibles con OpenAI) |
| `SPIDER_MAX_ITERATIONS` | `25` | Máximo de pasos de conversación |
| `SPIDER_ALLOW_EXTERNAL` | `false` | Permitir escritura fuera del proyecto |
| `SPIDER_MEMORY_DIR` | `~/.spider/memory/` | Directorio de almacenamiento de memoria |

### Arquitectura

```
              LLM Provider (OpenAI / Anthropic / Ollama …)
                      │
            ┌─────────▼──────────┐
            │  Planner Agent     │  divide tareas en pasos
            │  (orchestrator)    │  y los ejecuta en paralelo
            └──┬──────┬──────┬───┘
               │             │
       ┌───────▼──┐   ┌───────▼──┐
       │  Writer  │   │  Runner  │  sub-agentes ejecutándose
       └───────┬──┘   └───────┬──┘  en paralelo (máx 3)
               └──────┬───────┘
                      │
              ┌───────▼───────┐
              │    Analyst    │  corre después de writer + runner
              └───────┬───────┘
                      │
              ┌───────▼───────┐
              │   Debugger    │  solo si hay tests fallando
              └───────────────┘
```

### Desarrollo

```sh
make build   # compilar
make test    # ejecutar tests
make clean   # limpiar binario

# Compilar sin Makefile
go build -o spider ./cmd/spider
```

### Sistema de memoria

Spider compacta automáticamente el historial de la conversación cuando supera el 75% del contexto del modelo. Los resúmenes se guardan como archivos `.md` en `~/.spider/memory/` y pueden ser buscados por los agentes para recordar sesiones pasadas.

## Licencia

MIT
