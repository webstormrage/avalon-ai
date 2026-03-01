package dto

type GameV2 struct {
	ID              int    `json:"id"`
	MissionPriority int    `json:"missionPriority"`
	LeaderPosition  int    `json:"leaderPosition"`
	SpeakerPosition int    `json:"speakerPosition"`
	SkipsCount      int    `json:"skipsCount"`
	Wins            int    `json:"wins"`
	Fails           int    `json:"fails"`
	GameState       string `json:"gameState"`
}
