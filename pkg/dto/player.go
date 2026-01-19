package dto

type PlayerV2 struct {
	ID               int
	Name             string
	Model            string
	Role             string
	Voice            string
	Mood             string
	TtsModel         string
	VoiceTemperature float32
	VoiceStyle       string
	Position         int
	GameID           int
}
