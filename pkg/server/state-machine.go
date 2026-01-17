package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/store"
	"fmt"
)

func (h *GameHandler) stateMachine(gameID int) (string, error) {
	tx, err := h.DB.BeginTx(h.Ctx, nil)
	if err != nil {
		return "", err
	}
	defer tx.Rollback()

	game, err := store.GetGame(h.Ctx, tx, gameID)

	if err != nil {
		return "", err
	}
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return "", err
	}

	var isLeader bool = game.SpeakerPosition == game.LeaderPosition
	promptStatus := ""
	if len(pendingPrompts) > 0 {
		promptStatus = pendingPrompts[0].Status
	}
	initialState := fmt.Sprintf("%s %d:%d {%d} Leader#%d Speaker#%d %s", game.GameState, game.Wins, game.Fails, game.SkipsCount, game.LeaderPosition, game.SpeakerPosition, promptStatus)
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
	}
	if err != nil {
		return "", err
	}

	game, err = store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return "", err
	}

	pendingPrompts, err = store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return "", err
	}
	promptStatus = ""
	if len(pendingPrompts) > 0 {
		promptStatus = pendingPrompts[0].Status
	}
	nextState := fmt.Sprintf("%s %d:%d {%d} Leader#%d Speaker#%d %s", game.GameState, game.Wins, game.Fails, game.SkipsCount, game.LeaderPosition, game.SpeakerPosition, promptStatus)
	return fmt.Sprintf("\n%s ========> %s\n", initialState, nextState), tx.Commit()
}
