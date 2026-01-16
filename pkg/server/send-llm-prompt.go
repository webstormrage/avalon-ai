package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
)

func (h *GameHandler) sendLlmPrompt(game *dto.GameV2, prompt *dto.Prompt) error {
	speaker, err := store.GetPlayerByPosition(h.Ctx, h.DB, game.ID, game.SpeakerPosition)

	if err != nil {
		return err
	}

	response, err := h.Agent.Send(dto.Persona{ //TODO: передавать спикера на прямую
		Self:      speaker.Name,
		ModelName: speaker.Model,
	},
		prompt.SystemPrompt,
		prompt.MessagePrompt,
		[]dto.Action{}, //TODO: передавать логи!!!
	)
	if err != nil {
		return err
	}
	prompt.Response = response
	prompt.Status = constants.STATUS_HAS_RESPONSE

	return store.UpdatePrompt(h.Ctx, h.DB, prompt)
}
