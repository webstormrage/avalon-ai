package domain

import (
	"avalon/pkg/action"
	"avalon/pkg/dto"
	"avalon/pkg/gemini"
	"math/rand"
)

func GetInitialState(missions []int, players []*gemini.Character) *dto.GameState {
	return &dto.GameState{
		Missions:     missions,
		Players:      players,
		MissionIndex: 0,
		LeaderIndex:  rand.Intn(len(players)),
		SkipsCount:   0,
		Wins:         0,
		Fails:        0,
		Logs:         []action.Action{},
	}
}
