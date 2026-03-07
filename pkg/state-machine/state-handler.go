package statemachine

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/server"
	"avalon/pkg/store"
)

func getState(h *Handler, tx store.QueryRower, gameID int) (*server.GameState, error) {
	game, err := store.GetGame(h.Ctx, tx, gameID)

	if err != nil {
		return nil, err
	}
	activePrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return nil, err
	}
	var prompt *dto.Prompt
	if len(activePrompts) > 0 {
		prompt = &activePrompts[0]
	}
	event := "NO_EVENT"

	players, err := store.GetPlayersByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return nil, err
	}
	requiredAction := getRequiredAction(*game, players)

	return &server.GameState{
		Game:           *game,
		Prompt:         prompt,
		Players:        players,
		CurrentEvent:   event,
		RequiredAction: requiredAction,
	}, nil
}

func getRequiredAction(game dto.GameV2, players []dto.PlayerV2) *server.RequiredAction {
	speakerID, isSpeakerLeader := getSpeakerAndLeaderInfo(game, players)
	if speakerID == 0 {
		return nil
	}

	switch game.GameState {
	case constants.STATE_DISCUSSION:
		if isSpeakerLeader {
			return &server.RequiredAction{
				PlayerID: speakerID,
				Name:     "propose_squad",
				ParamsDef: map[string]string{
					"message": "Р В­РЎвЂљР С• Р С—РЎС“Р В±Р В»Р С‘РЎвЂЎР Р…Р С•Р Вµ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р Вµ Р Т‘Р В»РЎРЏ Р Р†РЎРѓР ВµРЎвЂ¦ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р†",
					"squad":   "РЎРѓР С—Р С‘РЎРѓР С•Р С” Р Р…Р С•Р СР ВµРЎР‚Р С•Р Р† Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р† РЎРѓР С•РЎРѓРЎвЂљР В°Р Р†Р В°",
				},
			}
		}
		return &server.RequiredAction{
			PlayerID: speakerID,
			Name:     "rate_squad",
			ParamsDef: map[string]string{
				"message": "Р С—РЎС“Р В±Р В»Р С‘РЎвЂЎР Р…Р С•Р Вµ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р Вµ Р Т‘Р В»РЎРЏ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р†",
			},
		}
	case constants.STATE_VOTING:
		if isSpeakerLeader {
			return &server.RequiredAction{
				PlayerID: speakerID,
				Name:     "announce_squad",
				ParamsDef: map[string]string{
					"message": "Р В­РЎвЂљР С• Р С—РЎС“Р В±Р В»Р С‘РЎвЂЎР Р…Р С•Р Вµ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р Вµ Р Т‘Р В»РЎРЏ Р Р†РЎРѓР ВµРЎвЂ¦ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р†",
					"squad":   "РЎРѓР С—Р С‘РЎРѓР С•Р С” Р Р…Р С•Р СР ВµРЎР‚Р С•Р Р† Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р† РЎРѓР С•РЎРѓРЎвЂљР В°Р Р†Р В°",
				},
			}
		}
		return &server.RequiredAction{
			PlayerID: speakerID,
			Name:     "vote_squad",
			ParamsDef: map[string]string{
				"approve": "Р вЂРЎС“Р В»Р ВµР В°Р Р…, Р С–Р С•Р В»Р С•РЎРѓ Р В·Р В° Р С‘Р В»Р С‘ Р С—РЎР‚Р С•РЎвЂљР С‘Р Р† РЎРѓР С•РЎРѓРЎвЂљР В°Р Р†Р В°",
			},
		}
	case constants.STATE_MISSION:
		return &server.RequiredAction{
			PlayerID: speakerID,
			Name:     "mission_action",
			ParamsDef: map[string]string{
				"success": "Р вЂРЎС“Р В»Р ВµР В°Р Р…. РЎС“РЎРѓР С—Р ВµРЎвЂ¦ Р С‘Р В»Р С‘ Р С—РЎР‚Р С•Р Р†Р В°Р В» Р СР С‘РЎРѓРЎРѓР С‘Р С‘",
			},
		}
	case constants.STATE_ASSASSINATION_DISCUSSION, constants.STATE_ASSASSIONATION_DISCUSSION:
		if isSpeakerLeader {
			return &server.RequiredAction{
				PlayerID: speakerID,
				Name:     "propose_assassination",
				ParamsDef: map[string]string{
					"message": "Р С—РЎС“Р В±Р В»Р С‘РЎвЂЎР Р…Р С•Р Вµ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р Вµ Р Т‘Р В»РЎРЏ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р†",
					"target":  "Р СњР С•Р СР ВµРЎР‚ Р В¶Р ВµРЎР‚РЎвЂљР Р†РЎвЂ№ РЎвЂ Р ВµР В»Р С•Р Вµ РЎвЂЎР С‘РЎРѓР В»Р С•",
				},
			}
		}
		return &server.RequiredAction{
			PlayerID: speakerID,
			Name:     "rate_assassination",
			ParamsDef: map[string]string{
				"message": "Р СџРЎС“Р В±Р В»Р С‘РЎвЂЎР Р…Р С•Р Вµ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р Вµ Р Т‘Р В»РЎРЏ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р†",
			},
		}
	case constants.STATE_ASSASSINATION, constants.STATE_ASSASSIONATION:
		return &server.RequiredAction{
			PlayerID: speakerID,
			Name:     "announce_assassination",
			ParamsDef: map[string]string{
				"message": "Р С—РЎС“Р В±Р В»Р С‘РЎвЂЎР Р…Р С•Р Вµ РЎРѓР С•Р С•Р В±РЎвЂ°Р ВµР Р…Р С‘Р Вµ Р Т‘Р В»РЎРЏ Р С‘Р С–РЎР‚Р С•Р С”Р С•Р Р†",
				"target":  "Р СњР С•Р СР ВµРЎР‚ Р В¶Р ВµРЎР‚РЎвЂљР Р†РЎвЂ№ РЎвЂ Р ВµР В»Р С•Р Вµ РЎвЂЎР С‘РЎРѓР В»Р С•",
			},
		}
	default:
		return nil
	}
}

func getSpeakerAndLeaderInfo(game dto.GameV2, players []dto.PlayerV2) (speakerID int, isSpeakerLeader bool) {
	leaderID := 0
	for _, player := range players {
		if player.Position == game.SpeakerPosition {
			speakerID = player.ID
		}
		if player.Position == game.LeaderPosition {
			leaderID = player.ID
		}
	}
	return speakerID, speakerID != 0 && speakerID == leaderID
}

func handleNextState(h *Handler, gameID int) (*server.GameState, error) {
	tx, err := h.DB.BeginTx(h.Ctx, nil)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	game, err := store.GetGame(h.Ctx, tx, gameID)

	if err != nil {
		return nil, err
	}

	var isLeader bool = game.SpeakerPosition == game.LeaderPosition

	if err != nil {
		return nil, err
	}
	prevState := game.GameState
	prevWins := game.Wins
	prevFails := game.Fails
	prevSkips := game.SkipsCount
	switch game.GameState {
	case constants.STATE_DISCUSSION:
		if isLeader {
			err = handleLeaderDiscussion(h, tx, gameID)
		} else {
			err = handleSpeakerDiscussion(h, tx, gameID)
		}
	case constants.STATE_VOTING:
		if isLeader {
			err = handleLeaderVoting(h, tx, gameID)
		} else {
			err = handleSpeakerVoting(h, tx, gameID)
		}
	case constants.STATE_MISSION:
		err = handleMission(h, tx, gameID)
	case constants.STATE_ASSASSINATION, constants.STATE_ASSASSIONATION:
		err = handleAssassination(h, tx, gameID)
	case constants.STATE_RED_VICTORY:
	case constants.STATE_BLUE_VICTORY:
	}
	if err != nil {
		return nil, err
	}
	state, err := getState(h, tx, gameID)
	if state.Game.GameState == constants.STATE_VOTING && prevState != constants.STATE_VOTING {
		state.CurrentEvent = "VOTING_STARTED"
	} else if state.Game.GameState == constants.STATE_MISSION && prevState != constants.STATE_MISSION {
		state.CurrentEvent = "MISSION_STARTED"
	} else if (state.Game.GameState == constants.STATE_ASSASSIONATION || state.Game.GameState == constants.STATE_ASSASSINATION) &&
		(prevState != constants.STATE_ASSASSIONATION && prevState != constants.STATE_ASSASSINATION) {
		state.CurrentEvent = "ASSASSION_STARTED"
	} else if state.Game.GameState == constants.STATE_BLUE_VICTORY && prevState != constants.STATE_BLUE_VICTORY {
		state.CurrentEvent = "BLUE_WON"
	} else if state.Game.GameState == constants.STATE_RED_VICTORY && prevState != constants.STATE_RED_VICTORY {
		state.CurrentEvent = "BLUE_LOST"
	} else if prevState == constants.STATE_MISSION && state.Game.Wins > prevWins {
		state.CurrentEvent = "MISSION_COMPLETED"
	} else if prevState == constants.STATE_MISSION && state.Game.Fails > prevFails {
		state.CurrentEvent = "MISSION_FAILED"
	} else if prevState == constants.STATE_VOTING && state.Game.SkipsCount > prevSkips {
		state.CurrentEvent = "LEADER_SKIPPED"
	}
	if err != nil {
		return nil, err
	}
	return state, tx.Commit()
}
