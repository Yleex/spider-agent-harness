# Spider

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-MIT-blue)](LICENSE)

**Spider is your AI-powered testing team. Tell it what to test, and it writes, runs, and fixes tests automatically.**

## What is Spider?

- **Saves you hours** — instead of writing tests by hand, describe what you need in plain English and Spider does the work
- **No AI expertise required** — works with ChatGPT, Claude, or free local models like Ollama. You only need an API key
- **Keeps your project safe** — Spider asks for permission before running commands or changing files. Nothing happens without your approval

## Quick start

```sh
# 1. One-time setup (you'll need an API key from OpenAI or Anthropic)
npx spider-agent-harness init

# 2. Ask Spider to write tests for your code
npx spider-agent-harness run writer "generate tests for pkg/agent"

# 3. See all available assistants
npx spider-agent-harness list
```

## What can you do with it?

### Your project has zero tests
```sh
spider run writer "generate tests for all files in pkg/"
spider run runner "run the generated tests"
```

### A test is failing and you don't know why
```sh
spider run debugger "diagnose the failure in TestFoo"
```

### You need a quality report before a release
```sh
spider run analyst "analyze test coverage and detect flaky tests"
```

### You want everything automated from start to finish
```sh
spider run planner "generate tests for the entire project, run them, analyze coverage, and fix any failures"
```

### You need tests that use a database or external API
```sh
spider run integrator "set up postgres + redis, run integration tests"
```

## How it works

1. **You describe a task in plain English** — Spider understands what you need
2. **Spider plans and executes** — it reads your code, writes tests, runs them, and analyzes results
3. **It works like a team of specialists** — each with a specific role (writer, runner, debugger, etc.)
4. **You stay in control** — Spider asks before running commands or modifying files

## The 6 assistants

| Assistant | What it does |
|---|---|
| **writer** | Reads your code and writes tests for it |
| **runner** | Runs your tests and tells you if they pass or fail |
| **analyst** | Checks how much of your code is covered by tests |
| **debugger** | Investigates why a test is failing and suggests how to fix it |
| **integrator** | Sets up databases, APIs, and external services for testing |
| **planner** | The team leader — breaks big tasks into smaller steps and coordinates the others |

## Installation

### Easiest: npx (no install required)

```sh
npx spider-agent-harness run writer "generate tests for cmd/"
```

> Requires Node.js 18+. The binary downloads automatically on first use.

### Alternative: Go install

```sh
go install github.com/Yleex/spider-agent-harness@latest
```

> Requires Go 1.22+. Installs as `spider` in your `$GOPATH/bin`.

### Alternative: Manual download

Download the pre-built binary from the [Releases page](https://github.com/Yleex/spider-agent-harness/releases) for your operating system.

## FAQ

**Do I need an API key?** Yes. Run `spider init` to configure one (OpenAI or Anthropic).

**Can I use it offline?** Yes, with a local model like Ollama.

**Is it safe?** Yes. Spider can only access files inside your project folder. It asks for permission before running commands or writing files.

**Can it run tests in parallel?** Yes. The `planner` assistant automatically runs independent tasks at the same time to save time.

**Can I add my own assistant?** Yes. If you're a developer, you can create custom assistants. See the development docs below.

## Multi-model support

Spider works with any provider compatible with OpenAI's API format:

- **OpenAI** — GPT-4o, GPT-4, GPT-3.5
- **Anthropic** — Claude 3 Opus, Claude 3.5 Sonnet
- **Ollama** (local) — `export SPIDER_API_BASE=http://localhost:11434/v1`
- **Groq** — `export SPIDER_API_BASE=https://api.groq.com/openai/v1`
- Any OpenAI-compatible endpoint

Add custom providers by calling `llm.RegisterProvider()` in `pkg/llm/registry.go`.

### Security

Three layers that cannot be disabled:

1. **Scope Enforcer** — No tool can read or write files outside the project directory. Symlinks and path traversal are resolved and blocked.
2. **Permission Checker** — `bash` and `write_file` require explicit approval by default. Pattern-based rules (e.g., deny `rm -rf`) can be added.
3. **Human-in-the-Loop** — Before executing a dangerous tool, the user is prompted. External writes require double confirmation.

### Environment variables

| Variable | Default | Description |
|---|---|---|
| `OPENAI_API_KEY` | — | OpenAI API key |
| `ANTHROPIC_API_KEY` | — | Anthropic API key |
| `SPIDER_PROVIDER` | `openai` | Provider name |
| `SPIDER_MODEL` | `gpt-4o` | Model ID |
| `SPIDER_API_BASE` | — | Base URL (for OpenAI-compatible APIs) |
| `SPIDER_MAX_ITERATIONS` | `25` | Max conversation steps |
| `SPIDER_ALLOW_EXTERNAL` | `false` | Allow writes outside project |
| `SPIDER_MEMORY_DIR` | `~/.spider/memory/` | Memory storage directory |

### Architecture

```
              LLM Provider (OpenAI / Anthropic / Ollama …)
                      │
            ┌─────────▼──────────┐
            │  Planner Agent     │  breaks tasks into steps
            │  (orchestrator)    │  and runs them in parallel
            └──┬──────┬──────┬───┘
               │             │
       ┌───────▼──┐   ┌───────▼──┐
       │  Writer  │   │  Runner  │  sub-agents running concurrently
       └───────┬──┘   └───────┬──┘  (max 3 at a time)
               └──────┬───────┘
                      │
              ┌───────▼───────┐
              │    Analyst    │  runs after writer + runner
              └───────┬───────┘
                      │
              ┌───────▼───────┐
              │   Debugger    │  only if tests fail
              └───────────────┘
```

### Development

```sh
make build   # compile
make test    # run tests
make clean   # remove binary

# Build without Makefile
go build -o spider ./cmd/spider
```

### Memory system

Spider automatically compacts conversation history when it exceeds 75% of the model's context window. Summaries are saved as `.md` files in `~/.spider/memory/` and can be searched by agents to recall past sessions.

## License

MIT
