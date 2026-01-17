package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/store"
)

func (h *GameHandler) stateMachine(gameID int) error {
	tx, err := h.DB.BeginTx(h.Ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	game, err := store.GetGame(h.Ctx, tx, gameID)

	if err != nil {
		return err
	}

	var isLeader bool = game.SpeakerPosition == game.LeaderPosition
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
		}
	}
	if err != nil {
		return err
	}
	return tx.Commit()
}
