package approval

import (
	"context"
	"fmt"
	"strings"
)

type ConfirmApprover struct {
	wrapped Approver
}

var _ Approver = (*ConfirmApprover)(nil)

func NewConfirmApprover(wrapped Approver) *ConfirmApprover {
	return &ConfirmApprover{wrapped: wrapped}
}

func (c *ConfirmApprover) RequestApproval(ctx context.Context, toolName string, args map[string]any) (bool, error) {
	approved, err := c.wrapped.RequestApproval(ctx, toolName, args)
	if err != nil {
		return false, err
	}
	if !approved {
		return false, nil
	}

	fmt.Printf("\n⚠  SEGUNDA CONFIRMACIÓN\n")
	fmt.Printf("  Tool: %s\n", toolName)
	for k, v := range args {
		fmt.Printf("  %s: %v\n", k, v)
	}
	return c.ask("¿Confirmar definitivamente? (y/N): ")
}

func (c *ConfirmApprover) ConfirmExternal(ctx context.Context, toolName string, args map[string]any) (bool, error) {
	return c.wrapped.ConfirmExternal(ctx, toolName, args)
}

func (c *ConfirmApprover) ask(msg string) (bool, error) {
	resp, err := promptFunc(msg)
	if err != nil {
		return false, err
	}
	resp = strings.TrimSpace(strings.ToLower(resp))
	return resp == "y" || resp == "yes", nil
}
