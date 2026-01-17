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

	switch game.GameState {
	case constants.STATE_DISCUSSION:
		if game.SpeakerPosition == game.LeaderPosition {
			err = h.handleLeaderDiscussion(tx, gameID)
		} else {
			err = h.handleSpeakerDiscussion(tx, gameID)
		}
	}
	return err
}
