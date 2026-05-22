package builtin

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"spider/pkg/tool"
	"strings"
)

func ReadFileTool() tool.Tool {
	return tool.Tool{
		Name:        "read_file",
		Description: "Lee el contenido de un archivo",
		Parameters: map[string]tool.ParameterSchema{
			"path": {
				Type:        "string",
				Description: "Ruta del archivo a leer",
				Required:    true,
			},
			"offset": {
				Type:        "number",
				Description: "Línea inicial (1-indexed, opcional)",
				Required:    false,
			},
			"limit": {
				Type:        "number",
				Description: "Máximo de líneas a leer (opcional)",
				Required:    false,
			},
		},
		PathArgs: []string{"path"},
		Execute: func(ctx context.Context, args map[string]any) (any, error) {
			path, _ := args["path"].(string)
			file, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("leyendo archivo: %w", err)
			}
			defer file.Close()

			offset, _ := args["offset"].(float64)
			limit, _ := args["limit"].(float64)

			if offset <= 0 && limit <= 0 {
				data, err := os.ReadFile(path)
				if err != nil {
					return nil, fmt.Errorf("leyendo archivo: %w", err)
				}
				return string(data), nil
			}

			var lines []string
			scanner := bufio.NewScanner(file)
			startLine := int(offset)
			if startLine < 1 {
				startLine = 1
			}
			maxLines := int(limit)

			for lineNum := 1; scanner.Scan(); lineNum++ {
				if lineNum < startLine {
					continue
				}
				if maxLines > 0 && lineNum >= startLine+maxLines {
					break
				}
				lines = append(lines, scanner.Text())
			}

			return strings.Join(lines, "\n"), nil
		},
	}
}

func WriteFileTool() tool.Tool {
	return tool.Tool{
		Name:        "write_file",
		Description: "Escribe contenido en un archivo",
		Parameters: map[string]tool.ParameterSchema{
			"path": {
				Type:        "string",
				Description: "Ruta del archivo a escribir",
				Required:    true,
			},
			"content": {
				Type:        "string",
				Description: "Contenido a escribir",
				Required:    true,
			},
		},
		PathArgs: []string{"path"},
		Execute: func(ctx context.Context, args map[string]any) (any, error) {
			path, _ := args["path"].(string)
			content, _ := args["content"].(string)

			dir := filepath.Dir(path)
			if err := os.MkdirAll(dir, 0755); err != nil {
				return nil, fmt.Errorf("creando directorio: %w", err)
			}

			if err := os.WriteFile(path, []byte(content), 0644); err != nil {
				return nil, fmt.Errorf("escribiendo archivo: %w", err)
			}

			return fmt.Sprintf("archivo escrito: %s (%d bytes)", path, len(content)), nil
		},
	}
}

func ListFilesTool() tool.Tool {
	return tool.Tool{
		Name:        "list_files",
		Description: "Lista archivos en un directorio",
		Parameters: map[string]tool.ParameterSchema{
			"path": {
				Type:        "string",
				Description: "Ruta del directorio",
				Required:    true,
			},
			"pattern": {
				Type:        "string",
				Description: "Glob pattern (ej: *.go)",
				Required:    false,
			},
		},
		PathArgs: []string{"path"},
		Execute: func(ctx context.Context, args map[string]any) (any, error) {
			path, _ := args["path"].(string)
			pattern, _ := args["pattern"].(string)

			if pattern != "" {
				matches, err := filepath.Glob(filepath.Join(path, pattern))
				if err != nil {
					return nil, fmt.Errorf("glob: %w", err)
				}
				return matches, nil
			}

			entries, err := os.ReadDir(path)
			if err != nil {
				return nil, fmt.Errorf("leyendo directorio: %w", err)
			}

			var names []string
			for _, e := range entries {
				names = append(names, e.Name())
			}
			return names, nil
		},
	}
}
