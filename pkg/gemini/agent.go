package gemini

import (
	"avalon/pkg/dto"
	"context"
	"fmt"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"strings"
)

func f32(v float32) *float32 {
	return &v
}

func i32(v int32) *int32 {
	return &v
}

func getContents(logs []dto.Action, user string) []*genai.Content {
	var contents []*genai.Content

	for _, log := range logs {
		var role string
		var message string

		if log.User == user {
			role = "model"
			message = log.Message
		} else {
			role = "user"
			message = log.User + ": " + log.Message
		}

		contents = append(contents, &genai.Content{
			Role: role,
			Parts: []genai.Part{
				genai.Text(message),
			},
		})
	}

	return contents
}

type GeminiAgent struct {
	client *genai.Client
	ctx    context.Context
}

func NewAgent(
	ctx context.Context,
	apiKey string,
) (dto.Agent, error) {

	client, err := genai.NewClient(
		ctx,
		option.WithAPIKey(apiKey),
	)
	if err != nil {
		return nil, err
	}

	return &GeminiAgent{
		ctx:    ctx,
		client: client,
	}, nil
}

func (a *GeminiAgent) Close() error {
	return a.client.Close()
}

func (a *GeminiAgent) Send(
	persona dto.Persona,
	systemPrompt string,
	instruction string,
	logs []dto.Action,
) (string, error) {

	model := a.client.GenerativeModel(persona.ModelName)

	if systemPrompt != "" {
		model.SystemInstruction = &genai.Content{
			Role: "system",
			Parts: []genai.Part{
				genai.Text(systemPrompt),
			},
		}
	}

	model.Temperature = f32(0.4)
	model.TopP = f32(0.95)
	model.TopK = i32(40)

	chat := model.StartChat()
	chat.History = getContents(logs, persona.Self)

	resp, err := chat.SendMessage(a.ctx, genai.Text(instruction))
	if err != nil {
		return "", err
	}

	if len(resp.Candidates) == 0 ||
		len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response")
	}

	var result strings.Builder
	for _, part := range resp.Candidates[0].Content.Parts {
		if t, ok := part.(genai.Text); ok {
			result.WriteString(string(t))
		}
	}

	return result.String(), nil
}
