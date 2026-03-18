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

// CloudflareProvider implements the Provider interface for Cloudflare Workers AI
type CloudflareProvider struct {
	apiToken  string
	accountID string
	model     string
	client    *http.Client
}

// NewCloudflareProvider creates a new Cloudflare Workers AI provider
func NewCloudflareProvider(apiToken, accountID, model string) *CloudflareProvider {
	if model == "" {
		model = "@cf/meta/llama-3-8b-instruct"
	}
	return &CloudflareProvider{
		apiToken:  apiToken,
		accountID: accountID,
		model:     model,
		client:    &http.Client{Timeout: 60 * time.Second},
	}
}

func (p *CloudflareProvider) Name() string {
	return "cloudflare"
}

func (p *CloudflareProvider) Available() bool {
	return p.apiToken != "" && p.accountID != ""
}

type cfRequest struct {
	Messages []cfMessage `json:"messages"`
}

type cfMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type cfResponse struct {
	Result struct {
		Response string `json:"response"`
	} `json:"result"`
	Success bool `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

func (p *CloudflareProvider) GenerateResponse(ctx context.Context, prompt string, systemData models.SystemContext) (string, error) {
	if !p.Available() {
		return "", fmt.Errorf("Cloudflare API token or account ID not configured")
	}

	fullPrompt := BuildPrompt(prompt, systemData)

	reqBody := cfRequest{
		Messages: []cfMessage{
			{Role: "system", Content: GetSystemPrompt()},
			{Role: "user", Content: fullPrompt},
		},
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s", p.accountID, p.model)

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+p.apiToken)

	resp, err := p.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result cfResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if !result.Success {
		errMsg := "unknown error"
		if len(result.Errors) > 0 {
			errMsg = result.Errors[0].Message
		}
		return "", fmt.Errorf("API error: %s", errMsg)
	}

	return result.Result.Response, nil
}
