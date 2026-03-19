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

// CopilotProvider implements the Provider interface for GitHub Copilot
// Note: This uses the Copilot Chat API which requires a valid GitHub Copilot subscription
type CopilotProvider struct {
	token  string
	model  string
	client *http.Client
}

// NewCopilotProvider creates a new GitHub Copilot provider
func NewCopilotProvider(token, model string) *CopilotProvider {
	if model == "" {
		model = "gpt-4o"
	}
	return &CopilotProvider{
		token:  token,
		model:  model,
		client: &http.Client{Timeout: 90 * time.Second},
	}
}

func (p *CopilotProvider) Name() string {
	return "copilot"
}

func (p *CopilotProvider) Available() bool {
	return p.token != ""
}

// Copilot uses OpenAI-compatible API format
type copilotRequest struct {
	Model     string           `json:"model"`
	Messages  []copilotMessage `json:"messages"`
	MaxTokens int              `json:"max_tokens,omitempty"`
}

type copilotMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type copilotResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (p *CopilotProvider) GenerateResponse(ctx context.Context, prompt string, systemData models.SystemContext, privacyConfig models.PrivacyConfig) (string, error) {
	if !p.Available() {
		return "", fmt.Errorf("GitHub token not configured")
	}

	fullPrompt := BuildPrompt(prompt, systemData, privacyConfig)

	reqBody := copilotRequest{
		Model: p.model,
		Messages: []copilotMessage{
			{Role: "system", Content: GetSystemPrompt()},
			{Role: "user", Content: fullPrompt},
		},
		MaxTokens: 2048,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	// GitHub Copilot API endpoint
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.githubcopilot.com/chat/completions", bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.token)
	req.Header.Set("Editor-Version", "SysMind/1.0")
	req.Header.Set("Copilot-Integration-Id", "sysmind-desktop")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API error (HTTP %d): %s", resp.StatusCode, string(body))
	}

	var result copilotResponse
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
