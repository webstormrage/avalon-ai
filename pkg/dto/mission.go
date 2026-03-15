package dto

import "encoding/json"

type MissionV2 struct {
	ID           int
	Name         string
	Status       string
	MaxFails     int
	AllowedFails int
	SquadSize    int
	Priority     int
	Squad        []int
	Skips        int
	Leader       int
	Progress     int
	Fails        int
	Successes    int
	Votes        json.RawMessage
	GameID       int
}
