package selectors

import (
	"avalon/pkg/dto"
	"avalon/pkg/prompts"
	"slices"
)

func GetLeader(state *dto.GameState) *dto.Character {
	return state.Players[state.LeaderIndex]
}

func GetMission(state *dto.GameState) prompts.MissionProps {
	return prompts.MissionProps{
		Index: state.MissionIndex,
		Size:  state.Missions[state.MissionIndex],
	}
}

func GetPlayerByRole(state *dto.GameState, role string) *dto.Character {
	idx := slices.IndexFunc(state.Players, func(p *dto.Character) bool {
		return p.Persona.Role == role
	})
	return state.Players[idx]
}
