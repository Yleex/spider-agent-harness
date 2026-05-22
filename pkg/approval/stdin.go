package approval

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
)

type StdinApprover struct {
	Prompt func(msg string) (string, error)
}

func NewStdinApprover() *StdinApprover {
	return &StdinApprover{
		Prompt: promptFunc,
	}
}

func (a *StdinApprover) RequestApproval(ctx context.Context, toolName string, args map[string]any) (bool, error) {
	fmt.Printf("\n── Tool: %s ──\n", toolName)
	for k, v := range args {
		fmt.Printf("  %s: %v\n", k, v)
	}
	return a.ask("¿Permitir ejecución? (y/N): ")
}

func (a *StdinApprover) ConfirmExternal(ctx context.Context, toolName string, args map[string]any) (bool, error) {
	fmt.Printf("\n⚠  MODIFICACIÓN FUERA DEL PROYECTO\n")
	fmt.Printf("  Tool: %s\n", toolName)
	for k, v := range args {
		fmt.Printf("  %s: %v\n", k, v)
	}
	return a.ask("¿Confirmar esta modificación externa? (y/N): ")
}

func (a *StdinApprover) ask(msg string) (bool, error) {
	resp, err := a.Prompt(msg)
	if err != nil {
		return false, err
	}
	resp = strings.TrimSpace(strings.ToLower(resp))
	return resp == "y" || resp == "yes", nil
}

func promptFunc(msg string) (string, error) {
	fmt.Print(msg)
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}
