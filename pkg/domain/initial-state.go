package domain

import (
	"avalon/pkg/dto"
	"math/rand"
)

func GetInitialState(missions []int, players []*dto.Character) *dto.GameState {
	return &dto.GameState{
		Missions:     missions,
		Players:      players,
		MissionIndex: 0,
		LeaderIndex:  rand.Intn(len(players)),
		SkipsCount:   0,
		Wins:         0,
		Fails:        0,
		Logs:         []dto.Action{},
	}
}
