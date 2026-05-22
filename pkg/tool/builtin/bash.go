package builtin

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"spider/pkg/tool"
	"strings"
)

func BashTool() tool.Tool {
	return tool.Tool{
		Name:        "bash",
		Description: "Ejecuta un comando en la terminal del sistema",
		Parameters: map[string]tool.ParameterSchema{
			"command": {
				Type:        "string",
				Description: "Comando a ejecutar",
				Required:    true,
			},
		},
		PathArgs: nil,
		Execute: func(ctx context.Context, args map[string]any) (any, error) {
			cmdStr, _ := args["command"].(string)
			if cmdStr == "" {
				return nil, fmt.Errorf("command es requerido")
			}

			var cmd *exec.Cmd
			if runtime.GOOS == "windows" {
				cmd = exec.CommandContext(ctx, "powershell", "-Command", cmdStr)
			} else {
				cmd = exec.CommandContext(ctx, "sh", "-c", cmdStr)
			}

			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr

			err := cmd.Run()
			outStr := stdout.String()
			errStr := stderr.String()

			if err != nil {
				return nil, fmt.Errorf("ejecutando comando: %w\nstderr: %s\nstdout: %s", err, errStr, outStr)
			}

			if errStr != "" {
				outStr += "\nstderr:\n" + errStr
			}

			return strings.TrimSpace(outStr), nil
		},
	}
}
