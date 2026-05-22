package tool

import (
	"context"
	"fmt"
	"spider/pkg/permission"
	"spider/pkg/scope"
)

type ParameterSchema struct {
	Type        string   `json:"type"`
	Description string   `json:"description,omitempty"`
	Required    bool     `json:"required,omitempty"`
	Enum        []string `json:"enum,omitempty"`
}

type Tool struct {
	Name        string
	Description string
	Parameters  map[string]ParameterSchema
	PathArgs    []string
	Execute     func(ctx context.Context, args map[string]any) (any, error)
}

type ToolInstance struct {
	Spec          Tool
	ScopeVal      *scope.Validator
	PermCheck     *permission.Checker
	Approver      permission.Approver
	AllowExternal bool
}

var ErrPermissionDenied = fmt.Errorf("permission denied")

func (t *ToolInstance) Call(ctx context.Context, args map[string]any) (any, error) {
	args = shallowCopy(args)

	if err := t.enforceScope(args); err != nil {
		return nil, err
	}

	if t.PermCheck != nil {
		action, err := t.PermCheck.Evaluate(t.Spec.Name, args)
		if err != nil {
			return nil, err
		}
		if action == permission.ActionDeny {
			return nil, ErrPermissionDenied
		}
		if action == permission.ActionAsk && t.Approver != nil {
			approved, err := t.Approver.RequestApproval(ctx, t.Spec.Name, args)
			if err != nil {
				return nil, err
			}
			if !approved {
				return nil, ErrPermissionDenied
			}
		}
	}

	return t.Spec.Execute(ctx, args)
}

func shallowCopy(orig map[string]any) map[string]any {
	cp := make(map[string]any, len(orig))
	for k, v := range orig {
		cp[k] = v
	}
	return cp
}

func (t *ToolInstance) enforceScope(args map[string]any) error {
	for _, pathArg := range t.Spec.PathArgs {
		raw, ok := args[pathArg]
		if !ok {
			continue
		}
		pathStr, ok := raw.(string)
		if !ok || pathStr == "" {
			continue
		}

		resolved, err := t.ScopeVal.Safe(pathStr)
		if err == nil {
			args[pathArg] = resolved
			continue
		}

		if !t.AllowExternal {
			return fmt.Errorf("SECURITY: acceso denegado — el path queda fuera del proyecto: %s", pathStr)
		}

		if t.Approver == nil {
			return ErrPermissionDenied
		}

		approved, err := t.Approver.RequestApproval(context.Background(), t.Spec.Name, args)
		if err != nil {
			return err
		}
		if !approved {
			return ErrPermissionDenied
		}

		confirmed, err := t.Approver.ConfirmExternal(context.Background(), t.Spec.Name, args)
		if err != nil {
			return err
		}
		if !confirmed {
			return ErrPermissionDenied
		}

		args[pathArg] = pathStr
	}
	return nil
}
