package dto

type Prompt struct {
	ID            int    `json:"id"`
	GameID        int    `json:"gameId"`
	Model         string `json:"model"`
	SystemPrompt  string `json:"systemPrompt"`
	MessagePrompt string `json:"messagePrompt"`
	Response      string `json:"response"`
	Status        string `json:"status"`
}
