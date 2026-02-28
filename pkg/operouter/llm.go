package operouter

import (
	"avalon/pkg/dto"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model        string        `json:"model"`
	SystemPrompt string        `json:"system_prompt"`
	Messages     []chatMessage `json:"messages"`
	Instruction  string        `json:"instruction"`
}

type chatResponse struct {
	Status   string `json:"status"`
	Model    string `json:"model"`
	Response string `json:"response"`
	ID       string `json:"id"`
}

type errorResponse struct {
	Detail any `json:"detail"`
}

func getMessages(logs []*dto.Event, user string) []chatMessage {
	messages := make([]chatMessage, 0, len(logs))

	for _, log := range logs {
		if log == nil {
			continue
		}

		var role string
		var message string

		if log.Source == user {
			role = "assistant"
			message = log.Content
		} else {
			role = "user"
			message = fmt.Sprintf("[%s]%s: %s\n", log.Source, log.Type, log.Content)
		}

		messages = append(messages, chatMessage{
			Role:    role,
			Content: message,
		})
	}

	return messages
}

type OpenRouterAgent struct {
	ctx     context.Context
	client  *http.Client
	baseURL string
}

func NewAgent(
	ctx context.Context,
	baseURL string,
) (dto.Agent, error) {
	baseURL = strings.TrimSpace(baseURL)
	if baseURL == "" {
		return nil, fmt.Errorf("open router baseURL is empty")
	}

	return &OpenRouterAgent{
		ctx: ctx,
		client: &http.Client{
			Timeout: 10 * time.Minute,
		},
		baseURL: strings.TrimRight(baseURL, "/"),
	}, nil
}

func (a *OpenRouterAgent) Close() error {
	return nil
}

func (a *OpenRouterAgent) Send(
	persona dto.PlayerV2,
	systemPrompt string,
	instruction string,
	logs []*dto.Event,
) (string, error) {

	reqBody := chatRequest{
		Model:        persona.Model,
		SystemPrompt: systemPrompt,
		Messages:     getMessages(logs, persona.Name),
		Instruction:  instruction,
	}

	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal chat request: %w", err)
	}

	req, err := http.NewRequestWithContext(
		a.ctx,
		http.MethodPost,
		a.baseURL+"/chat",
		bytes.NewReader(raw),
	)
	if err != nil {
		return "", fmt.Errorf("build chat request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("send chat request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp errorResponse
		if decodeErr := json.NewDecoder(resp.Body).Decode(&errResp); decodeErr == nil {
			return "", fmt.Errorf("open-router error: status=%d detail=%v", resp.StatusCode, errResp.Detail)
		}
		return "", fmt.Errorf("open-router error: status=%d", resp.StatusCode)
	}

	var out chatResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", fmt.Errorf("decode chat response: %w", err)
	}

	if strings.TrimSpace(out.Response) == "" {
		return "", fmt.Errorf("empty response")
	}

	return out.Response, nil
}
