package config

import (
	"fmt"
	"os"
	"spider/pkg/schema"
	"strconv"
)

func LoadFromEnv() schema.AgentConfig {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		apiKey = os.Getenv("ANTHROPIC_API_KEY")
	}

	provider := getEnv("SPIDER_PROVIDER", "openai")
	model := getEnv("SPIDER_MODEL", "gpt-4o")
	maxIter, _ := strconv.Atoi(getEnv("SPIDER_MAX_ITERATIONS", "25"))
	allowExternal, _ := strconv.ParseBool(getEnv("SPIDER_ALLOW_EXTERNAL", "false"))

	return schema.AgentConfig{
		Name:          getEnv("SPIDER_AGENT", "default"),
		MaxIterations: maxIter,
		AllowExternal: allowExternal,
		Provider: schema.ProviderConfig{
			Provider:   provider,
			Model:      model,
			APIKey:     apiKey,
			Temp:       0.7,
			MaxTokens:  4096,
		},
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
