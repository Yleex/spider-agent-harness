package builtin

import (
	"context"
	"fmt"
	"spider/pkg/memory"
	"spider/pkg/tool"
	"strings"
)

func MemorySearchTool(store *memory.FileStore) tool.Tool {
	return tool.Tool{
		Name:        "memory_search",
		Description: "Busca en la memoria persistente del proyecto (resúmenes de sesiones anteriores)",
		Parameters: map[string]tool.ParameterSchema{
			"query": {
				Type:        "string",
				Description: "Términos de búsqueda",
				Required:    true,
			},
			"tags": {
				Type:        "string",
				Description: "Filtrar por tags separados por coma (opcional)",
				Required:    false,
			},
		},
		PathArgs: nil,
		Execute: func(ctx context.Context, args map[string]any) (any, error) {
			query, _ := args["query"].(string)

			var tags []string
			if tagsRaw, ok := args["tags"].(string); ok && tagsRaw != "" {
				for _, t := range strings.Split(tagsRaw, ",") {
					t = strings.TrimSpace(t)
					if t != "" {
						tags = append(tags, t)
					}
				}
			}

			result, err := store.Search(query, tags, 5)
			if err != nil {
				return nil, fmt.Errorf("buscando memoria: %w", err)
			}

			return result, nil
		},
	}
}

func MemorySaveTool(store *memory.FileStore) tool.Tool {
	return tool.Tool{
		Name:        "memory_save",
		Description: "Guarda un fragmento importante en la memoria persistente del proyecto",
		Parameters: map[string]tool.ParameterSchema{
			"content": {
				Type:        "string",
				Description: "Contenido a guardar en memoria",
				Required:    true,
			},
			"tags": {
				Type:        "string",
				Description: "Tags separados por coma para categorizar la memoria",
				Required:    false,
			},
		},
		PathArgs: nil,
		Execute: func(ctx context.Context, args map[string]any) (any, error) {
			content, _ := args["content"].(string)
			if content == "" {
				return nil, fmt.Errorf("content es requerido")
			}

			var tags []string
			if tagsRaw, ok := args["tags"].(string); ok && tagsRaw != "" {
				for _, t := range strings.Split(tagsRaw, ",") {
					t = strings.TrimSpace(t)
					if t != "" {
						tags = append(tags, t)
					}
				}
			}

			entry := &memory.SummaryEntry{
				Summary:    content,
				TokensSaved: 0,
			}

			path, err := store.Save(entry, "manual", tags)
			if err != nil {
				return nil, fmt.Errorf("guardando memoria: %w", err)
			}

			return fmt.Sprintf("Memoria guardada en %s", path), nil
		},
	}
}
