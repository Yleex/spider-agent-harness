# Spider

[🇬🇧 English](README.md) · [🇪🇸 Español](README.es.md)

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)]()

**Spider** is a multi-model AI agent harness specialized for software testing. It runs a ReAct loop (Reason → Act → Observe) with LLM providers, tools, memory, and human-in-the-loop security — all from the command line.

## Installation

### Option 1: npx (no install, binary auto-download)

```sh
npx spider-agent-harness run writer "generate tests for cmd/"
```

> Requires Node.js 18+. The binary is downloaded automatically on first use.

### Option 2: Go install

```sh
go install github.com/user/spider@latest
```

### Option 3: Manual download

Download the pre-built binary from [Releases](https://github.com/user/spider/releases) for your OS/architecture.

## Quick start

```sh
# 1. Interactive setup (asks for API key and provider)
spider init

# 2. Run a testing agent
spider run writer "generate unit tests for the package pkg/agent"

# 3. See all available agents
spider list
```

## The 5 testing agents

| Agent | Responsibility | Key tools |
|---|---|---|
| `writer` | Generates test files from source code | `read_file`, `write_file`, `list_files` |
| `runner` | Executes test suites and reports results | `bash` |
| `analyst` | Measures coverage, detects flaky tests | `bash`, `read_file` |
| `debugger` | Diagnoses failures and proposes fixes | `bash`, `read_file` |
| `integrator` | Manages external services, fixtures, E2E | `bash`, `read_file`, `list_files` |

Each agent has a specialized system prompt and scoped toolset. No overlap: one generates, another runs, another analyzes, etc.

## Use cases

### "Add tests to a legacy project with zero coverage"
```sh
spider run writer "generate tests for all files in pkg/"
spider run runner "execute the generated tests"
```

### "A test is failing and I don't know why"
```sh
spider run debugger "diagnose the failure in TestFoo"
```

### "Quality check before a release"
```sh
spider run analyst "analyze test coverage and detect flaky tests"
```

### "I need integration tests for a microservice"
```sh
spider run integrator "spin up postgres + redis, run contract tests"
```

## Multi-model support

Supported out of the box:

- **OpenAI** — GPT-4o, GPT-4, GPT-3.5
- **Anthropic** — Claude 3 Opus, Claude 3.5 Sonnet
- **Any OpenAI-compatible API** — Ollama, Groq, Together AI, vLLM, etc.

### Examples

```sh
# Ollama (local)
export SPIDER_API_BASE=http://localhost:11434/v1
export SPIDER_MODEL=llama3

# Groq
export SPIDER_API_BASE=https://api.groq.com/openai/v1
export SPIDER_MODEL=mixtral-8x7b-32768

# Custom provider (via code)
import "spider/pkg/llm"
llm.RegisterProvider("custom", func(cfg schema.ProviderConfig) llm.Provider {
    // return your provider
})
```

## Security (non-negotiable)

Three mandatory layers:

1. **Scope Enforcer** — No tool can read or write files outside the project directory. Symlinks, path traversal, and absolute paths are resolved and blocked.

2. **Permission Checker** — `bash` and `write_file` require explicit approval by default. Pattern-based rules (e.g., deny `rm -rf`) can be added.

3. **Human-in-the-Loop** — Before executing a dangerous tool, the user is prompted. External writes require **double confirmation** with a diff preview.

## Memory system

Spider automatically compacts conversation history when it exceeds 75% of the model's context window. Summaries are persisted as `.md` files in `~/.spider/memory/` and are searchable by agents.

```sh
ls ~/.spider/memory/
# index.json
# 2026-05-22_writer_3f7a8b2c.md
# 2026-05-22_runner_a1b2c3d4.md
```

Agents can query this memory using the `memory_search` tool to recall past sessions without re-consuming the full context.

## Environment reference

| Variable | Default | Description |
|---|---|---|
| `OPENAI_API_KEY` | — | OpenAI API key |
| `ANTHROPIC_API_KEY` | — | Anthropic API key |
| `SPIDER_PROVIDER` | `openai` | Provider name |
| `SPIDER_MODEL` | `gpt-4o` | Model ID |
| `SPIDER_API_BASE` | — | Base URL (for OpenAI-compatible APIs) |
| `SPIDER_MAX_ITERATIONS` | `25` | Max ReAct steps |
| `SPIDER_ALLOW_EXTERNAL` | `false` | Allow writes outside project |
| `SPIDER_COMPACT_ENABLED` | `true` | Enable memory compaction |
| `SPIDER_CONTEXT_LIMIT` | `128000` | Model context tokens |
| `SPIDER_MEMORY_DIR` | `~/.spider/memory/` | Memory storage directory |

## Development

```sh
make build   # compile
make test    # run tests
make clean   # remove binary

# Build without Makefile
go build -o spider ./cmd/spider
```

## Architecture quick reference

```
LLM Provider (OpenAI / Anthropic / Ollama …)
       │
  ┌────▼────┐
  │  Agent  │  ReAct loop: think → act → observe
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

**Do I need an API key?** Yes. Run `spider init` to configure one.

**Can I use it offline?** Yes, with a local Ollama instance.

**Can I add my own agent?** Yes. Create a factory in `pkg/agents/` and register it in `main.go`.

**Is it CI/CD safe?** Yes, with `SPIDER_ALLOW_EXTERNAL=false`.

## License

MIT
