package dto

type MissionV2 struct {
	ID        int
	Name      string
	MaxFails  int
	SquadSize int
	Priority  int
	GameID    int
}
