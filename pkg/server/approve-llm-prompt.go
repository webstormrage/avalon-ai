package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
)

func (h *GameHandler) approveLlmPrompt(_ *dto.GameV2, prompt *dto.Prompt) error {
	prompt.Status = constants.STATUS_APPROVED

	return store.UpdatePrompt(h.Ctx, h.DB, prompt)
}
