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

// GLMProvider implements the Provider interface for Z.AI (Zhipu AI)
type GLMProvider struct {
	apiKey string
	model  string
	client *http.Client
}

// NewGLMProvider creates a new Z.AI/GLM provider
func NewGLMProvider(apiKey, model string) *GLMProvider {
	if model == "" {
		model = "glm-5"
	}
	return &GLMProvider{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{Timeout: 90 * time.Second},
	}
}

func (p *GLMProvider) Name() string {
	return "glm"
}

func (p *GLMProvider) Available() bool {
	return p.apiKey != ""
}

// Z.AI uses standard OpenAI-compatible API format
type glmRequest struct {
	Model       string       `json:"model"`
	Messages    []glmMessage `json:"messages"`
	MaxTokens   int          `json:"max_tokens,omitempty"`
	Temperature float64      `json:"temperature,omitempty"`
}

type glmMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type glmResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

func (p *GLMProvider) GenerateResponse(ctx context.Context, prompt string, systemData models.SystemContext, privacyConfig models.PrivacyConfig) (string, error) {
	if !p.Available() {
		return "", fmt.Errorf("Z.AI (GLM) API key not configured")
	}

	fullPrompt := BuildPrompt(prompt, systemData, privacyConfig)

	reqBody := glmRequest{
		Model: p.model,
		Messages: []glmMessage{
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

	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.z.ai/api/paas/v4/chat/completions", bytes.NewBuffer(jsonBody))
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

	var result glmResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("API error: %s (code: %s)", result.Error.Message, result.Error.Code)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return result.Choices[0].Message.Content, nil
}
