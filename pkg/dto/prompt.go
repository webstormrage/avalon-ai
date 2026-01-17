package dto

type Prompt struct {
	ID            int
	GameID        int
	Model         string
	SystemPrompt  string
	MessagePrompt string
	Response      string
	Status        string
}
