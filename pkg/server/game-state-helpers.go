package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
)

func getState(h *GameHandler, tx store.QueryRower, gameID int) (*GameState, error) {
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
	players, err := store.GetPlayersByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return nil, err
	}
	return &GameState{
		Game:           *game,
		Prompt:         prompt,
		Players:        players,
		CurrentEvent:   "NO_EVENT",
		RequiredAction: getRequiredAction(*game, players),
	}, nil
}

func getRequiredAction(game dto.GameV2, players []dto.PlayerV2) *RequiredAction {
	speakerID, isSpeakerLeader := getSpeakerAndLeaderInfo(game, players)
	if speakerID == 0 {
		return nil
	}

	switch game.Phase {
	case constants.STATE_DISCUSSION:
		if isSpeakerLeader {
			return &RequiredAction{
				PlayerID: speakerID,
				Name:     "propose_squad",
				ParamsDef: map[string]string{
					"message":      "Это публичное сообщение для всех игроков",
					"squadNumbers": "список номеров игроков состава",
				},
			}
		}
		return &RequiredAction{
			PlayerID: speakerID,
			Name:     "rate_squad",
			ParamsDef: map[string]string{
				"message": "Публичное сообщение для игроков",
			},
		}
	case constants.STATE_VOTING:
		if isSpeakerLeader {
			return &RequiredAction{
				PlayerID: speakerID,
				Name:     "announce_squad",
				ParamsDef: map[string]string{
					"message":      "Это публичное сообщение для всех игроков",
					"squadNumbers": "список номеров игроков состава",
				},
			}
		}
		return &RequiredAction{
			PlayerID: speakerID,
			Name:     "vote_squad",
			ParamsDef: map[string]string{
				"approve": "Булеан, голос за или против состава",
			},
		}
	case constants.STATE_MISSION:
		return &RequiredAction{
			PlayerID: speakerID,
			Name:     "mission_action",
			ParamsDef: map[string]string{
				"success": "Булеан. успех или провал миссии",
			},
		}
	case constants.STATE_ASSASSINATION_DISCUSSION, constants.STATE_ASSASSIONATION_DISCUSSION:
		if isSpeakerLeader {
			return &RequiredAction{
				PlayerID: speakerID,
				Name:     "propose_assassination",
				ParamsDef: map[string]string{
					"message": "Публичное сообщение для игроков",
					"target":  "Номер жертвы целое число",
				},
			}
		}
		return &RequiredAction{
			PlayerID: speakerID,
			Name:     "rate_assassination",
			ParamsDef: map[string]string{
				"message": "Публичное сообщение для игроков",
			},
		}
	case constants.STATE_ASSASSINATION, constants.STATE_ASSASSIONATION:
		return &RequiredAction{
			PlayerID: speakerID,
			Name:     "announce_assassination",
			ParamsDef: map[string]string{
				"message": "Публичное сообщение для игроков",
				"target":  "Номер жертвы целое число",
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
