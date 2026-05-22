# AGENTS.md

Spider — Agent Testing Harness. Go module `spider`, Go 1.22.

## Build & test

- Build: `go build -o spider ./cmd/spider`
- Test all: `go test ./...`
- Scope tests: `go test ./pkg/scope/`

## Architecture

```
cmd/spider/main.go     → CLI entrypoint (subcommands: init, list, run, help)
pkg/
  agent/               → Agent interface + ReAct loop + registry
  agents/{writer,runner,analyst,debugger,integrator}  → 5 testing specialists
  llm/                 → Provider interface + registry + OpenAI/Anthropic implementations
  tool/                → Tool system + builtin (bash, filesystem, memory_search)
  scope/               → PathValidator — blocks writes outside project root
  permission/          → Checker with allow/deny/ask rules
  approval/            → Human-in-the-loop + double-confirm for external writes
  memory/              → Compactor (LLM-summarizes old context) + FileStore (.md persistence)
  config/              → Env-var loading
  cli/                 → UI helpers (colors, prompts, help text)
  schema/              → Shared types (Message, ToolCall, ProviderConfig)
npm/                   → package.json + install.js for `npx spider-agent-harness`
.goreleaser.yaml       → Cross-compile for linux/darwin/windows × amd64/arm64
```

## Key commands

```sh
spider init        # interactive API key + provider setup
spider list        # list available agents
spider run <a> "<t>"  # run agent with task
spider help        # detailed help
```

## Multi-model

`SPIDER_PROVIDER` + `SPIDER_API_BASE` for any OpenAI-compatible endpoint (Ollama, Groq, etc.).
Register new providers via `llm.RegisterProvider()`.

## Security (non-negotiable)

- Scope enforcer blocks all file access outside project root
- bash/write_file default to `ask` permission
- Double confirmation required for external writes

## Conventions

- `go build ./...` before commit
- Agent factories in `pkg/agents/<name>/agent.go`
- New providers register in `pkg/llm/registry.go` init()
- New tools in `pkg/tool/builtin/`
