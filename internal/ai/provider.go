package ai

import (
	"context"

	"sysmind/internal/models"
)

// Provider defines the interface for AI providers
type Provider interface {
	// GenerateResponse generates a response based on the prompt and system context
	GenerateResponse(ctx context.Context, prompt string, systemData models.SystemContext) (string, error)
	// Name returns the provider name
	Name() string
	// Available checks if the provider is configured and available
	Available() bool
}

// NewProvider creates an AI provider based on the configuration
func NewProvider(config models.AIConfig) Provider {
	switch config.Provider {
	case "openai":
		return NewOpenAIProvider(config.APIKey, config.Model)
	case "anthropic":
		return NewAnthropicProvider(config.APIKey, config.Model)
	case "kimi":
		return NewKimiProvider(config.APIKey, config.Model)
	case "glm":
		return NewGLMProvider(config.APIKey, config.Model)
	case "copilot":
		return NewCopilotProvider(config.APIKey, config.Model)
	case "cloudflare":
		return NewCloudflareProvider(config.APIKey, config.CloudflareAcct, config.Model)
	case "local":
		return NewLocalProvider(config.LocalEndpoint, config.Model)
	default:
		return NewOpenAIProvider(config.APIKey, config.Model)
	}
}
