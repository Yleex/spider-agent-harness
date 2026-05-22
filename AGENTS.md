# AGENTS.md

Spider — Multi-model AI agent harness for software testing. Go module `spider`, Go 1.22.0.

## Build & test

- `go build -o spider ./cmd/spider`
- `go test ./...`
- `make build` / `make test` / `make clean` / `make run`
- `go mod tidy` before commits (goreleaser before hook)
- `go build ./...` before commit

## CLI

```
spider init              # interactive API key + provider setup
spider list              # list available agents
spider run <agent> "<t>" # run agent with task
spider help              # detailed help
```

## Agent tool sets

| Agent | Tools |
|---|---|
| writer | read_file, write_file, list_files [+ memory_search/save if MemoryStore set] |
| runner | bash [+ memory_search if MemoryStore set] |
| analyst | bash, read_file [+ memory_search if MemoryStore set] |
| debugger | bash, read_file [+ memory_search if MemoryStore set] |
| integrator | bash, read_file, list_files [+ memory_search if MemoryStore set] |
| planner | read_file, list_files, bash, run_subtasks, get_results [+ memory_search if MemoryStore set] |

Planner parallel pool: max 3 concurrent workers (hardcoded in `main.go`).

## Plan format (planner agent)

YAML-like key:value lines:
```
id: gen-tests
agent: writer
task: generate tests for pkg/core
depends: []
strategy: concurrente
```

Strategies: `concurrente`/`parallel`, `secuencial`/`sequential`, `aislado` (default when no depends).

## Multi-model

`SPIDER_PROVIDER` + `SPIDER_API_BASE` for any OpenAI-compatible endpoint (Ollama, Groq, etc.).
Register new providers via `llm.RegisterProvider()` in `pkg/llm/registry.go` init().
Defaults: temperature 0.7, max_tokens 4096.

## Security

- Scope enforcer (`pkg/scope/`) resolves symlinks, blocks writes outside project root
- bash/write_file default to `ask` permission (`pkg/permission/`)
- External writes require double confirmation (`pkg/approval/`)
- Tool scope enforced via `PathArgs` field on Tool struct

## Memory

- Auto-compaction triggers at 75% of `SPIDER_CONTEXT_LIMIT`
- Token estimate = rune-length × 2
- Persisted as `.md` files in `~/.spider/memory/` (or `SPIDER_MEMORY_DIR`)
- `NO_COLOR` env var disables ANSI in CLI output

## Environment

| Variable | Default | Note |
|---|---|---|
| `SPIDER_PROVIDER` | `openai` | Provider name |
| `SPIDER_MODEL` | `gpt-4o` | Model ID |
| `SPIDER_API_BASE` | — | For OpenAI-compatible APIs |
| `SPIDER_MAX_ITERATIONS` | `25` | Max ReAct steps |
| `SPIDER_ALLOW_EXTERNAL` | `false` | Allow writes outside project |
| `SPIDER_COMPACT_ENABLED` | `true` | Enable memory compaction |
| `SPIDER_CONTEXT_LIMIT` | `128000` | Model context tokens |
| `SPIDER_COMPACT_THRESHOLD` | `0.75` | Fraction triggering compaction |
| `SPIDER_RESERVE_EXCHANGES` | `5` | Exchanges preserved after compaction |
| `SPIDER_MEMORY_DIR` | `~/.spider/memory/` | Memory storage directory |
| `OPENAI_API_KEY` | — | OpenAI API key |
| `ANTHROPIC_API_KEY` | — | Anthropic API key |

## Adding agents

- Agent factory in `pkg/agents/<name>/agent.go`
- New tools in `pkg/tool/builtin/`
- Register factory in two places in `cmd/spider/main.go`: `runAgent` registry block and `runPlanner` pool block
