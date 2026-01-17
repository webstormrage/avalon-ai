package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
	"strings"
)

func (h *GameHandler) createLeaderDiscussionPrompt(tx store.QueryRower, gameID int) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	leader, err := store.GetPlayerByPosition(h.Ctx, tx, gameID, game.LeaderPosition)
	if err != nil {
		return err
	}
	players, err := store.GetPlayersByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	mission, err := store.GetMissionByPriority(h.Ctx, tx, gameID, game.MissionPriority)
	if err != nil {
		return err
	}
	missions, err := store.GetMissionsByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}

	return store.CreatePrompt(h.Ctx, tx, &dto.Prompt{
		GameID: gameID,
		Model:  leader.Model,
		SystemPrompt: prompts.GetSystemPrompt(
			prompts.SystemPromptProps{
				Name:     leader.Name,
				Mood:     leader.Mood,
				Players:  players,
				Roles:    presets.Roles5, // TODO: надо брать из базы
				Role:     leader.Role,
				Missions: missions,
			},
		),
		MessagePrompt: prompts.RenderProposalPrompt(prompts.StatementProps{
			Resume: prompts.ResumeProps{
				Wins:       game.Wins,
				Fails:      game.Fails,
				SkipsCount: game.SkipsCount,
			},
			Mission: *mission,
		}),
	})

}

func (h *GameHandler) applyLeaderDiscussionPrompt(tx store.QueryRower, gameID int) error {
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

	leader, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.LeaderPosition)
	if err != nil {
		return err
	}
	squad, _ := prompts.ExtractTeam(prompt.Response)
	err = store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:  game.ID,
		Type:    constants.EVENT_SQUAD_DECLARATION,
		Source:  leader.Name,
		Content: strings.Join(squad, ", "),
	})
	if err != nil {
		return err
	}
	err = store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:  game.ID,
		Type:    constants.EVENT_PLAYER_SPEECH,
		Source:  leader.Name,
		Content: prompt.Response,
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

	return store.UpdateGame(h.Ctx, tx, game)
}

func (h *GameHandler) handleLeaderDiscussion(tx store.QueryRower, gameID int) error {
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(pendingPrompts) == 0 {
		err = h.createLeaderDiscussionPrompt(tx, gameID)
	} else {
		switch pendingPrompts[0].Status {
		case constants.STATUS_NOT_STARTED:
			err = h.sendLlmPrompt(tx, gameID)
		case constants.STATUS_HAS_RESPONSE:
			err = h.applyLeaderDiscussionPrompt(tx, gameID)
		}
	}
	return err
}
