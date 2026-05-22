package config

import (
	"fmt"
	"os"
	"spider/pkg/memory"
	"spider/pkg/schema"
	"strconv"
)

type AppConfig struct {
	Agent       schema.AgentConfig
	MemoryDir   string
	CompactCfg  *memory.CompactConfig
}

func Load() AppConfig {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	provider := getEnv("SPIDER_PROVIDER", "openai")
	model := getEnv("SPIDER_MODEL", "gpt-4o")
	apiBase := getEnv("SPIDER_API_BASE", "")
	maxIter, _ := strconv.Atoi(getEnv("SPIDER_MAX_ITERATIONS", "25"))
	allowExternal, _ := strconv.ParseBool(getEnv("SPIDER_ALLOW_EXTERNAL", "false"))
	memoryDir := getEnv("SPIDER_MEMORY_DIR", "")

	contextLimit, _ := strconv.Atoi(getEnv("SPIDER_CONTEXT_LIMIT", "128000"))
	threshold, _ := strconv.ParseFloat(getEnv("SPIDER_COMPACT_THRESHOLD", "0.75"), 64)
	reserveExchanges, _ := strconv.Atoi(getEnv("SPIDER_RESERVE_EXCHANGES", "5"))

	var compactCfg *memory.CompactConfig
	compactEnabled, _ := strconv.ParseBool(getEnv("SPIDER_COMPACT_ENABLED", "true"))
	if compactEnabled {
		compactCfg = &memory.CompactConfig{
			ContextLimit:     contextLimit,
			Threshold:        threshold,
			ReserveExchanges: reserveExchanges,
		}
	}

	return AppConfig{
		Agent: schema.AgentConfig{
			Name:          getEnv("SPIDER_AGENT", "default"),
			MaxIterations: maxIter,
			AllowExternal: allowExternal,
			Provider: schema.ProviderConfig{
				Provider:  provider,
				Model:     model,
				APIKey:    apiKey,
				APIBase:   apiBase,
				Temp:      0.7,
				MaxTokens: 4096,
			},
		},
		MemoryDir:  memoryDir,
		CompactCfg: compactCfg,
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func ResolveProjectRoot() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("obteniendo working directory: %w", err)
	}
	return wd, nil
}
