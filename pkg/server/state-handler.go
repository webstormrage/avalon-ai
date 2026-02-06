package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
)

func (h *GameHandler) getState(tx store.QueryRower, gameID int) (*GameState, error) {
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

	return &GameState{
		Game:    *game,
		Prompt:  prompt,
		Players: players,
	}, nil
}

type GameState struct {
	Game    dto.GameV2     `json:"game"`
	Prompt  *dto.Prompt    `json:"prompt,omitempty"`
	Players []dto.PlayerV2 `json:"players,omitempty"`
}

func (h *GameHandler) handleNextState(gameID int) (*GameState, error) {
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
	switch game.GameState {
	case constants.STATE_DISCUSSION:
		if isLeader {
			err = h.handleLeaderDiscussion(tx, gameID)
		} else {
			err = h.handleSpeakerDiscussion(tx, gameID)
		}
	case constants.STATE_VOTING:
		if isLeader {
			err = h.handleLeaderVoting(tx, gameID)
		} else {
			err = h.handleSpeakerVoting(tx, gameID)
		}
	case constants.STATE_MISSION:
		err = h.handleMission(tx, gameID)
	case constants.STATE_ASSASSIONATION:
		err = h.handleAssassination(tx, gameID)
	case constants.STATE_RED_VICTORY:
	case constants.STATE_BLUE_VICTORY:
	}
	if err != nil {
		return nil, err
	}
	state, err := h.getState(tx, gameID)
	if err != nil {
		return nil, err
	}
	return state, tx.Commit()
}
