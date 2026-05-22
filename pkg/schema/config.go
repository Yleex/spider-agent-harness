package schema

type ProviderConfig struct {
	Provider string  `yaml:"provider" json:"provider"`
	Model    string  `yaml:"model" json:"model"`
	APIKey   string  `yaml:"api_key" json:"-"`
	Temp     float64 `yaml:"temperature" json:"temperature"`
	MaxTokens int   `yaml:"max_tokens" json:"max_tokens"`
}

type AgentConfig struct {
	Name         string            `yaml:"name" json:"name"`
	SystemPrompt string            `yaml:"system_prompt" json:"system_prompt"`
	Provider     ProviderConfig    `yaml:"provider" json:"provider"`
	MaxIterations int              `yaml:"max_iterations" json:"max_iterations"`
	AllowExternal bool             `yaml:"allow_external" json:"allow_external"`
}
