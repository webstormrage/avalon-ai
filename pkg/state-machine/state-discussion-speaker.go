package statemachine

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
)

func createSpeakerDiscussionPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	leader, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.LeaderPosition)
	if err != nil {
		return err
	}
	players, err := store.GetPlayersByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}
	mission, err := store.GetMissionByPriority(h.Ctx, tx, game.ID, game.MissionPriority)
	if err != nil {
		return err
	}
	missions, err := store.GetMissionsByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}

	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.SpeakerPosition)
	if err != nil {
		return err
	}

	squadEvent, err := store.GetLastEventByGameIDAndType(h.Ctx, tx, gameID, constants.EVENT_SQUAD_DECLARATION)
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
		MessagePrompt: prompts.RenderCommentPrompt(prompts.VoteProps{
			Mission: *mission,
			Leader:  leader.Name,
			Team:    squadEvent.Content,
		}),
	})
}

func applySpeakerDiscussionPrompt(h *Handler, tx store.QueryRower, gameID int) error {
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

	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.SpeakerPosition)
	if err != nil {
		return err
	}

	err = store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: speaker.ID,
		Content:  prompt.Response,
	})

	if err != nil {
		return err
	}

	count, err := store.CountPlayersByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}

	game.SpeakerPosition += 1
	if game.SpeakerPosition > count {
		game.SpeakerPosition = 1
	}
	if game.SpeakerPosition == game.LeaderPosition {
		// РљСЂСѓРі Р·Р°РјРєРЅСѓР»СЃСЏ
		game.GameState = constants.STATE_VOTING
	}

	return store.UpdateGame(h.Ctx, tx, game)
}

func handleSpeakerDiscussion(h *Handler, tx store.QueryRower, gameID int) error {
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(pendingPrompts) == 0 {
		err = createSpeakerDiscussionPrompt(h, tx, gameID)
	} else {
		switch pendingPrompts[0].Status {
		case constants.STATUS_NOT_STARTED:
			err = sendLlmPrompt(h, tx, gameID)
		case constants.STATUS_HAS_RESPONSE:
			err = applySpeakerDiscussionPrompt(h, tx, gameID)
		}
	}
	return err
}
