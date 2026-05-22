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
	"spider/pkg/orchestrator"
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

func buildInfra(appCfg config.AppConfig) (*scope.Validator, llm.Provider, *permission.Checker, permission.Approver, *memory.FileStore) {
	projectRoot, err := config.ResolveProjectRoot()
	if err != nil {
		cli.Fatal("No se pudo determinar el directorio del proyecto: " + err.Error())
	}

	scopeVal, err := scope.New(projectRoot)
	if err != nil {
		cli.Fatal("Error de seguridad: " + err.Error())
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
			cli.Warn("No se pudo inicializar almacén de memoria: " + err.Error())
		} else {
			memStore = store
		}
	}

	return scopeVal, provider, permCheck, appr, memStore
}

func buildAgentCfg(name string, appCfg config.AppConfig, provider llm.Provider, scopeVal *scope.Validator, permCheck *permission.Checker, appr permission.Approver, memStore *memory.FileStore) agent.Config {
	return agent.Config{
		Name:          name,
		Provider:      provider,
		MaxIterations: appCfg.Agent.MaxIterations,
		AllowExternal: appCfg.Agent.AllowExternal,
		ScopeVal:      scopeVal,
		PermCheck:     permCheck,
		Approver:      appr,
		CompactCfg:    appCfg.CompactCfg,
		MemoryStore:   memStore,
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

	scopeVal, provider, permCheck, appr, memStore := buildInfra(appCfg)

	if agentName == "planner" {
		runPlanner(task, appCfg, provider, scopeVal, permCheck, appr, memStore)
		return
	}

	reg := agent.NewRegistry()
	reg.Register("writer", func(c agent.Config) agent.Agent { return writer.New(c) })
	reg.Register("runner", func(c agent.Config) agent.Agent { return runner.New(c) })
	reg.Register("analyst", func(c agent.Config) agent.Agent { return analyst.New(c) })
	reg.Register("debugger", func(c agent.Config) agent.Agent { return debugger.New(c) })
	reg.Register("integrator", func(c agent.Config) agent.Agent { return integrator.New(c) })

	agentCfg := buildAgentCfg(agentName, appCfg, provider, scopeVal, permCheck, appr, memStore)

	a, err := reg.Create(agentName, agentCfg)
	if err != nil {
		cli.Fatal(err.Error())
	}

	cli.Info("Ejecutando agente " + cli.Bold(agentName) + "...")
	cli.Info("Tarea: " + task)

	result, err := a.Run(context.Background(), task)
	if err != nil {
		cli.Fatal("Error: " + err.Error())
	}

	cli.Done("Agente completado")
	fmt.Println()
	fmt.Println(result)
}

func runPlanner(task string, appCfg config.AppConfig, provider llm.Provider, scopeVal *scope.Validator, permCheck *permission.Checker, appr permission.Approver, memStore *memory.FileStore) {
	store := memory.NewSharedResultStore()
	pool := orchestrator.NewPool(3)

	baseCfg := buildAgentCfg("planner", appCfg, provider, scopeVal, permCheck, appr, memStore)
	baseCfg.SharedStore = store

	pool.Register("writer", func(c agent.Config) agent.Agent { return writer.New(c) })
	pool.Register("runner", func(c agent.Config) agent.Agent { return runner.New(c) })
	pool.Register("analyst", func(c agent.Config) agent.Agent { return analyst.New(c) })
	pool.Register("debugger", func(c agent.Config) agent.Agent { return debugger.New(c) })
	pool.Register("integrator", func(c agent.Config) agent.Agent { return integrator.New(c) })

	planner := orchestrator.NewPlanner(baseCfg, pool, store)

	cli.Info("Ejecutando agente " + cli.Bold("planner") + " (orquestador con paralelismo)...")
	cli.Info("Tarea: " + task)

	result, err := planner.Run(context.Background(), task)
	if err != nil {
		cli.Fatal("Error: " + err.Error())
	}

	cli.Done("Plan completado")
	fmt.Println()
	fmt.Println(result)
}

func listAgentNames() []string {
	return []string{"writer", "runner", "analyst", "debugger", "integrator", "planner"}
}
