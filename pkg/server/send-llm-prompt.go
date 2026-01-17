package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
)

func (h *GameHandler) sendLlmPrompt(tx store.QueryRower, gameID int) error {
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	prompt := pendingPrompts[0]

	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}

	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, gameID, game.SpeakerPosition)

	if err != nil {
		return err
	}

	events, err := store.GetEventsByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}

	response, err := h.Agent.Send(dto.Persona{ //TODO: передавать спикера на прямую
		Self:      speaker.Name,
		ModelName: speaker.Model,
	},
		prompt.SystemPrompt,
		prompt.MessagePrompt,
		events,
	)
	if err != nil {
		return err
	}
	prompt.Response = response
	prompt.Status = constants.STATUS_HAS_RESPONSE

	return store.UpdatePrompt(h.Ctx, tx, &prompt)
}
