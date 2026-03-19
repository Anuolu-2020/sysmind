package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"sysmind/internal/models"
)

// KimiProvider implements the Provider interface for Moonshot AI (Kimi)
type KimiProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewKimiProvider creates a new Kimi/Moonshot provider
func NewKimiProvider(apiKey, model string) *KimiProvider {
	if model == "" {
		model = "moonshot-v1-8k"
	}
	return &KimiProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 90 * time.Second},
	}
}

func (p *KimiProvider) Name() string {
	return "kimi"
}

func (p *KimiProvider) Available() bool {
	return p.apiKey != ""
}

// Kimi uses OpenAI-compatible API format
type kimiRequest struct {
	Model       string        `json:"model"`
	Messages    []kimiMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type kimiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type kimiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

func (p *KimiProvider) GenerateResponse(ctx context.Context, prompt string, systemData models.SystemContext, privacyConfig models.PrivacyConfig) (string, error) {
	if !p.Available() {
		return "", fmt.Errorf("Moonshot (Kimi) API key not configured")
	}

	fullPrompt := BuildPrompt(prompt, systemData, privacyConfig)

	reqBody := kimiRequest{
		Model: p.model,
		Messages: []kimiMessage{
			{Role: "system", Content: GetSystemPrompt()},
			{Role: "user", Content: fullPrompt},
		},
		MaxTokens:   2048,
		Temperature: 0.7,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.moonshot.cn/v1/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result kimiResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return result.Choices[0].Message.Content, nil
}
