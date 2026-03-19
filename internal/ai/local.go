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

// LocalProvider implements the Provider interface for local LLMs (e.g., Ollama)
type LocalProvider struct {
	endpoint string
	model    string
	client   *http.Client
}

// NewLocalProvider creates a new local LLM provider
func NewLocalProvider(endpoint, model string) *LocalProvider {
	if endpoint == "" {
		endpoint = "http://localhost:11434"
	}
	if model == "" {
		model = "llama3"
	}
	return &LocalProvider{
		endpoint: endpoint,
		model:    model,
		client:   &http.Client{Timeout: 120 * time.Second},
	}
}

func (p *LocalProvider) Name() string {
	return "local"
}

func (p *LocalProvider) Available() bool {
	// Try to connect to the endpoint
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", p.endpoint, nil)
	if err != nil {
		return false
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return false
	}
	resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// Ollama API format
type ollamaRequest struct {
	Model    string          `json:"model"`
	Messages []ollamaMessage `json:"messages"`
	Stream   bool            `json:"stream"`
}

type ollamaMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ollamaResponse struct {
	Message struct {
		Content string `json:"content"`
	} `json:"message"`
	Error string `json:"error,omitempty"`
}

func (p *LocalProvider) GenerateResponse(ctx context.Context, prompt string, systemData models.SystemContext, privacyConfig models.PrivacyConfig) (string, error) {
	fullPrompt := BuildPrompt(prompt, systemData, privacyConfig)

	reqBody := ollamaRequest{
		Model: p.model,
		Messages: []ollamaMessage{
			{Role: "system", Content: GetSystemPrompt()},
			{Role: "user", Content: fullPrompt},
		},
		Stream: false,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := p.endpoint + "/api/chat"

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed (is Ollama running?): %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result ollamaResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != "" {
		return "", fmt.Errorf("API error: %s", result.Error)
	}

	return result.Message.Content, nil
}
