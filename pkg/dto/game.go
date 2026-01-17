package dto

type GameV2 struct {
	ID              int
	MissionPriority int
	LeaderPosition  int
	SpeakerPosition int
	SkipsCount      int
	Wins            int
	Fails           int
	GameState       string
}
