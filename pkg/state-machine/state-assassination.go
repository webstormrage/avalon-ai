package statemachine

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
)

func createAssassinationPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}

	players, err := store.GetPlayersByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}

	missions, err := store.GetMissionsByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}

	assassins, err := store.GetPlayersByRole(h.Ctx, tx, gameID, constants.ROLE_ASSASSIN)
	if err != nil {
		return err
	}
	speaker := assassins[0]
	game.SpeakerPosition = speaker.Position

	err = store.UpdateGame(h.Ctx, tx, game)
	if err != nil {
		return err
	}

	return store.CreatePrompt(h.Ctx, tx, &dto.Prompt{
		GameID: game.ID,
		Model:  speaker.Model,
		SystemPrompt: prompts.GetSystemPrompt(
			prompts.SystemPromptProps{
				Name:     speaker.Name,
				Players:  players,
				Roles:    presets.Roles5, // TODO: РЅР°РґРѕ Р±СЂР°С‚СЊ РёР· СЃРёСЃС‚РµРјРЅС‹С… events
				Role:     speaker.Role,
				Missions: missions,
			},
		),
		MessagePrompt: prompts.AssassinationPrompt,
	})
}

func applyAssassinationPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	prompt := pendingPrompts[0]
	prompt.Status = constants.STATUS_COMPLETED

	err = store.UpdatePrompt(h.Ctx, tx, &prompt)
	if err != nil {
		return err
	}

	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}

	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, gameID, game.SpeakerPosition)
	if err != nil {
		return err
	}

	targetName, _ := prompts.ExtractAssassinationTarget(prompt.Response)

	targets, err := store.FindPlayersByNameLike(h.Ctx, tx, gameID, targetName)

	err = store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_ASSASSINATION,
		PlayerID: speaker.ID,
		Content:  targetName,
	})
	err = store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: speaker.ID,
		Content:  prompt.Response,
	})
	if len(targets) > 0 && targets[0].Role == constants.ROLE_MERLIN {
		game.GameState = constants.STATE_RED_VICTORY
	} else {
		game.GameState = constants.STATE_BLUE_VICTORY
	}

	return store.UpdateGame(h.Ctx, tx, game)
}

func handleAssassination(h *Handler, tx store.QueryRower, gameID int) error {
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(pendingPrompts) == 0 {
		err = createAssassinationPrompt(h, tx, gameID)
	} else {
		switch pendingPrompts[0].Status {
		case constants.STATUS_NOT_STARTED:
			err = sendLlmPrompt(h, tx, gameID)
		case constants.STATUS_HAS_RESPONSE:
			err = applyAssassinationPrompt(h, tx, gameID)
		}
	}
	return err
}
