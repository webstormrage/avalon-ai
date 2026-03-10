package dto

import "encoding/json"

type MissionV2 struct {
	ID        int
	Name      string
	MaxFails  int
	SquadSize int
	Priority  int
	Squad     []int
	Progress  int
	Fails     int
	Successes int
	Skips     int
	Votes     json.RawMessage
	GameID    int
}
