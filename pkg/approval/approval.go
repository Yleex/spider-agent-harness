package approval

import "context"

type Approver interface {
	RequestApproval(ctx context.Context, toolName string, args map[string]any) (bool, error)
	ConfirmExternal(ctx context.Context, toolName string, args map[string]any) (bool, error)
}
