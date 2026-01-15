package dto

import (
	"avalon/pkg/action"
	"avalon/pkg/gemini"
)

type GameState struct {
	Missions     []int
	Players      []*gemini.Character
	MissionIndex int
	LeaderIndex  int
	SkipsCount   int
	Wins         int
	Fails        int
	Logs         []action.Action
}
