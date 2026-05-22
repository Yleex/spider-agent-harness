package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"spider/pkg/agent"
	"spider/pkg/agents/analyst"
	"spider/pkg/agents/debugger"
	"spider/pkg/agents/integrator"
	"spider/pkg/agents/runner"
	"spider/pkg/agents/writer"
	"spider/pkg/approval"
	"spider/pkg/cli"
	"spider/pkg/config"
	"spider/pkg/llm"
	"spider/pkg/memory"
	"spider/pkg/permission"
	"spider/pkg/scope"
)

func main() {
	appCfg := config.Load()

	if len(os.Args) < 2 {
		cli.RunHelp()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "init":
		cli.RunInit()
		return
	case "list":
		cli.RunList()
		return
	case "help", "--help", "-h":
		cli.RunHelp()
		return
	case "run":
		runAgent(appCfg)
		return
	default:
		cli.Warn("Comando desconocido: " + cmd)
		fmt.Println()
		cli.Info("Ejecuta " + cli.Cyan("spider help") + " para ver los comandos disponibles")
		os.Exit(1)
	}
}

func runAgent(appCfg config.AppConfig) {
	if len(os.Args) < 4 {
		cli.Warn("Uso: spider run <agente> \"<tarea>\"")
		cli.Info("Agentes: " + strings.Join(listAgentNames(), ", "))
		cli.Info("Ejemplo: " + cli.Cyan("spider run writer \"genera tests para pkg/agent\""))
		os.Exit(1)
	}

	agentName := os.Args[2]
	task := os.Args[3]

	projectRoot, err := config.ResolveProjectRoot()
	if err != nil {
		cli.Fatal(fmt.Sprintf("No se pudo determinar el directorio del proyecto: %v", err))
	}

	scopeVal, err := scope.New(projectRoot)
	if err != nil {
		cli.Fatal(fmt.Sprintf("Error de seguridad: %v", err))
	}

	provider, err := llm.GetProvider(appCfg.Agent.Provider)
	if err != nil {
		cli.Fatal(err.Error())
	}

	permCheck := permission.New()
	permCheck.SetDefault("bash", permission.ActionAsk)
	permCheck.SetDefault("write_file", permission.ActionAsk)

	appr := approval.NewConfirmApprover(approval.NewStdinApprover())

	var memStore *memory.FileStore
	if appCfg.MemoryDir != "" || appCfg.CompactCfg != nil {
		store, err := memory.NewFileStore(appCfg.MemoryDir)
		if err != nil {
			cli.Warn(fmt.Sprintf("No se pudo inicializar almacén de memoria: %v", err))
		} else {
			memStore = store
		}
	}

	reg := agent.NewRegistry()
	reg.Register("writer", func(c agent.Config) agent.Agent { return writer.New(c) })
	reg.Register("runner", func(c agent.Config) agent.Agent { return runner.New(c) })
	reg.Register("analyst", func(c agent.Config) agent.Agent { return analyst.New(c) })
	reg.Register("debugger", func(c agent.Config) agent.Agent { return debugger.New(c) })
	reg.Register("integrator", func(c agent.Config) agent.Agent { return integrator.New(c) })

	agentCfg := agent.Config{
		Name:          agentName,
		Provider:      provider,
		MaxIterations: appCfg.Agent.MaxIterations,
		AllowExternal: appCfg.Agent.AllowExternal,
		ScopeVal:      scopeVal,
		PermCheck:     permCheck,
		Approver:      appr,
		CompactCfg:    appCfg.CompactCfg,
		MemoryStore:   memStore,
	}

	a, err := reg.Create(agentName, agentCfg)
	if err != nil {
		cli.Fatal(err.Error())
	}

	cli.Info(fmt.Sprintf("Ejecutando agente %s...", cli.Bold(agentName)))
	cli.Info(fmt.Sprintf("Tarea: %s", task))

	result, err := a.Run(context.Background(), task)
	if err != nil {
		cli.Fatal(fmt.Sprintf("Error: %v", err))
	}

	cli.Done("Agente completado")
	fmt.Println()
	fmt.Println(result)
}

func listAgentNames() []string {
	return []string{"writer", "runner", "analyst", "debugger", "integrator"}
}
