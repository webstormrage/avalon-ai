package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/store"
	"fmt"
)

func (h *GameHandler) getState(tx store.QueryRower, gameID int) (string, error) {
	game, err := store.GetGame(h.Ctx, tx, gameID)

	if err != nil {
		return "", err
	}
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return "", err
	}

	promptStatus := ""
	if len(pendingPrompts) > 0 {
		promptStatus = pendingPrompts[0].Status
	}
	return fmt.Sprintf("%s %d:%d {%d} Leader#%d Speaker#%d %s", game.GameState, game.Wins, game.Fails, game.SkipsCount, game.LeaderPosition, game.SpeakerPosition, promptStatus), nil
}

func (h *GameHandler) handleNextState(gameID int) (string, error) {
	tx, err := h.DB.BeginTx(h.Ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	game, err := store.GetGame(h.Ctx, tx, gameID)

	if err != nil {
		return "", err
	}

	var isLeader bool = game.SpeakerPosition == game.LeaderPosition

	initialState, err := h.getState(tx, gameID)
	if err != nil {
		return "", err
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
		return initialState, nil
	case constants.STATE_BLUE_VICTORY:
		return initialState, nil
	}
	if err != nil {
		return "", err
	}

	nextState, err := h.getState(tx, gameID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("\n%s ========> %s\n", initialState, nextState), tx.Commit()
}
