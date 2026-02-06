package dto

type PlayerV2 struct {
	ID               int     `json:"id"`
	Name             string  `json:"name"`
	Model            string  `json:"model"`
	Role             string  `json:"role"`
	Voice            string  `json:"voice"`
	Mood             string  `json:"mood"`
	TtsModel         string  `json:"ttsModel"`
	VoiceTemperature float32 `json:"voiceTemperature"`
	VoiceStyle       string  `json:"voiceStyle"`
	Position         int     `json:"position"`
	GameID           int     `json:"gameId"`
}
