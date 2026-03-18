package services

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"sysmind/internal/models"
)

// ConfigService manages application configuration
type ConfigService struct {
	configPath string
	config     models.AIConfig
	mu         sync.RWMutex
}

// NewConfigService creates a new config service
func NewConfigService() (*ConfigService, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		configDir = "."
	}

	appConfigDir := filepath.Join(configDir, "sysmind")
	if err := os.MkdirAll(appConfigDir, 0755); err != nil {
		return nil, err
	}

	configPath := filepath.Join(appConfigDir, "config.json")

	cs := &ConfigService{
		configPath: configPath,
		config:     getDefaultConfig(),
	}

	// Load existing config if it exists
	cs.loadConfig()

	return cs, nil
}

func getDefaultConfig() models.AIConfig {
	return models.AIConfig{
		Provider:      "openai",
		Model:         "gpt-4o-mini",
		LocalEndpoint: "http://localhost:11434",
	}
}

func (cs *ConfigService) loadConfig() {
	data, err := os.ReadFile(cs.configPath)
	if err != nil {
		return
	}

	var config models.AIConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return
	}

	cs.config = config
}

// GetConfig returns the current configuration
func (cs *ConfigService) GetConfig() models.AIConfig {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	return cs.config
}

// SetConfig updates and saves the configuration
func (cs *ConfigService) SetConfig(config models.AIConfig) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	cs.config = config

	data, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(cs.configPath, data, 0600)
}

// GetAvailableProviders returns list of available AI providers
func (cs *ConfigService) GetAvailableProviders() []ProviderInfo {
	return []ProviderInfo{
		// OpenAI
		{
			ID:   "openai",
			Name: "OpenAI",
			Models: []ModelInfo{
				{ID: "gpt-4o", Name: "GPT-4o (Latest)"},
				{ID: "gpt-4o-mini", Name: "GPT-4o Mini (Fast)"},
				{ID: "gpt-4-turbo", Name: "GPT-4 Turbo"},
				{ID: "gpt-4", Name: "GPT-4"},
				{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo"},
				{ID: "o1-preview", Name: "o1 Preview (Reasoning)"},
				{ID: "o1-mini", Name: "o1 Mini (Reasoning)"},
			},
			RequiresAPIKey: true,
		},
		// Anthropic (Claude)
		{
			ID:   "anthropic",
			Name: "Anthropic (Claude)",
			Models: []ModelInfo{
				{ID: "claude-3-5-sonnet-20241022", Name: "Claude 3.5 Sonnet (Latest)"},
				{ID: "claude-3-5-haiku-20241022", Name: "Claude 3.5 Haiku (Fast)"},
				{ID: "claude-3-opus-20240229", Name: "Claude 3 Opus"},
				{ID: "claude-3-sonnet-20240229", Name: "Claude 3 Sonnet"},
				{ID: "claude-3-haiku-20240307", Name: "Claude 3 Haiku"},
			},
			RequiresAPIKey: true,
		},
		// Kimi (Moonshot)
		{
			ID:   "kimi",
			Name: "Kimi (Moonshot AI)",
			Models: []ModelInfo{
				{ID: "moonshot-v1-8k", Name: "Moonshot v1 8K"},
				{ID: "moonshot-v1-32k", Name: "Moonshot v1 32K"},
				{ID: "moonshot-v1-128k", Name: "Moonshot v1 128K"},
			},
			RequiresAPIKey: true,
		},
		// GLM (Zhipu)
		{
			ID:   "glm",
			Name: "GLM (Zhipu AI)",
			Models: []ModelInfo{
				{ID: "glm-4-plus", Name: "GLM-4 Plus"},
				{ID: "glm-4-flash", Name: "GLM-4 Flash (Fast)"},
				{ID: "glm-4-air", Name: "GLM-4 Air"},
				{ID: "glm-4-airx", Name: "GLM-4 AirX"},
				{ID: "glm-4-long", Name: "GLM-4 Long (1M Context)"},
				{ID: "glm-4v-plus", Name: "GLM-4V Plus (Vision)"},
				{ID: "glm-4v", Name: "GLM-4V (Vision)"},
			},
			RequiresAPIKey: true,
		},
		// GitHub Copilot
		{
			ID:   "copilot",
			Name: "GitHub Copilot",
			Models: []ModelInfo{
				{ID: "gpt-4o", Name: "GPT-4o"},
				{ID: "gpt-4", Name: "GPT-4"},
				{ID: "gpt-3.5-turbo", Name: "GPT-3.5 Turbo"},
				{ID: "claude-3.5-sonnet", Name: "Claude 3.5 Sonnet"},
			},
			RequiresAPIKey: true,
		},
		// Cloudflare Workers AI
		{
			ID:   "cloudflare",
			Name: "Cloudflare Workers AI",
			Models: []ModelInfo{
				{ID: "@cf/meta/llama-3.3-70b-instruct-fp8-fast", Name: "Llama 3.3 70B Instruct"},
				{ID: "@cf/meta/llama-3.2-11b-vision-instruct", Name: "Llama 3.2 11B Vision"},
				{ID: "@cf/meta/llama-3.2-3b-instruct", Name: "Llama 3.2 3B Instruct"},
				{ID: "@cf/meta/llama-3.2-1b-instruct", Name: "Llama 3.2 1B Instruct"},
				{ID: "@cf/meta/llama-3.1-70b-instruct", Name: "Llama 3.1 70B Instruct"},
				{ID: "@cf/meta/llama-3.1-8b-instruct", Name: "Llama 3.1 8B Instruct"},
				{ID: "@cf/meta/llama-3-8b-instruct", Name: "Llama 3 8B Instruct"},
				{ID: "@cf/mistral/mistral-7b-instruct-v0.2", Name: "Mistral 7B Instruct v0.2"},
				{ID: "@hf/thebloke/deepseek-coder-6.7b-instruct-awq", Name: "DeepSeek Coder 6.7B"},
				{ID: "@cf/qwen/qwen1.5-14b-chat-awq", Name: "Qwen 1.5 14B Chat"},
				{ID: "@cf/google/gemma-7b-it", Name: "Gemma 7B IT"},
			},
			RequiresAPIKey: true,
			RequiresAcctID: true,
		},
		// Local LLM (Ollama)
		{
			ID:   "local",
			Name: "Local LLM (Ollama)",
			Models: []ModelInfo{
				{ID: "llama3.2", Name: "Llama 3.2"},
				{ID: "llama3.1", Name: "Llama 3.1"},
				{ID: "llama3", Name: "Llama 3"},
				{ID: "llama2", Name: "Llama 2"},
				{ID: "mistral", Name: "Mistral"},
				{ID: "mixtral", Name: "Mixtral"},
				{ID: "codellama", Name: "Code Llama"},
				{ID: "deepseek-coder", Name: "DeepSeek Coder"},
				{ID: "phi3", Name: "Phi-3"},
				{ID: "gemma2", Name: "Gemma 2"},
				{ID: "qwen2", Name: "Qwen 2"},
				{ID: "yi", Name: "Yi"},
			},
			RequiresEndpoint: true,
		},
	}
}

// ProviderInfo contains information about an AI provider
type ProviderInfo struct {
	ID               string      `json:"id"`
	Name             string      `json:"name"`
	Models           []ModelInfo `json:"models"`
	RequiresAPIKey   bool        `json:"requiresApiKey"`
	RequiresAcctID   bool        `json:"requiresAcctId"`
	RequiresEndpoint bool        `json:"requiresEndpoint"`
}

// ModelInfo contains information about a model
type ModelInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
