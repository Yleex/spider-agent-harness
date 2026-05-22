package llm

import (
	"fmt"
	"spider/pkg/schema"
)

type ProviderFactory func(cfg schema.ProviderConfig) Provider

var registry = map[string]ProviderFactory{}

func RegisterProvider(name string, factory ProviderFactory) {
	registry[name] = factory
}

func GetProvider(cfg schema.ProviderConfig) (Provider, error) {
	fn, ok := registry[cfg.Provider]
	if !ok {
		return nil, fmt.Errorf("proveedor no soportado: %s (disponibles: %v)", cfg.Provider, RegisteredProviders())
	}
	return fn(cfg), nil
}

func RegisteredProviders() []string {
	var names []string
	for n := range registry {
		names = append(names, n)
	}
	return names
}

func init() {
	RegisterProvider("openai", func(cfg schema.ProviderConfig) Provider {
		return NewOpenAI(cfg.APIKey, cfg.Model, cfg.APIBase)
	})
	RegisterProvider("anthropic", func(cfg schema.ProviderConfig) Provider {
		return NewAnthropic(cfg.APIKey, cfg.Model)
	})
}
