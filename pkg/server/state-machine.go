package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
)

func (h *GameHandler) stateMachine(game *dto.GameV2, prompts []dto.Prompt) error {
	var err error
	switch game.GameState {
	case constants.STATE_DISCUSSION:
		err = h.handleDiscussion(game, prompts)
	}
	return err
}
