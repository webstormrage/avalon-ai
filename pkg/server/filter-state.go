package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
)

func FilterStateByRequester(state *GameState, requester dto.PlayerV2) {
	if state == nil || len(state.Players) == 0 {
		return
	}

	if requester.Role == constants.ROLE_GAME_MASTER {
		return
	}

	visibleRoles := make(map[int]bool)
	visibleRoles[requester.ID] = true

	switch requester.Role {
	case constants.ROLE_ARTHURS_LOYAL:
		visibleRoles[requester.ID] = true
	case constants.ROLE_MERLIN:
		visibleRoles[requester.ID] = true
		for _, player := range state.Players {
			if player.Role == constants.ROLE_ASSASSIN || player.Role == constants.ROLE_MORDRED_MINION {
				visibleRoles[player.ID] = true
			}
		}
	case constants.ROLE_ASSASSIN, constants.ROLE_MORDRED_MINION:
		for _, player := range state.Players {
			if player.Role == constants.ROLE_ASSASSIN || player.Role == constants.ROLE_MORDRED_MINION {
				visibleRoles[player.ID] = true
			}
		}
	}

	filteredPlayers := make([]dto.PlayerV2, 0, len(state.Players))
	for i := range state.Players {
		if state.Players[i].Role == constants.ROLE_GAME_MASTER {
			continue
		}
		if _, ok := visibleRoles[state.Players[i].ID]; !ok {
			state.Players[i].Role = ""
		}
		state.Players[i].ID = 0
		filteredPlayers = append(filteredPlayers, state.Players[i])
	}
	state.Players = filteredPlayers
}
