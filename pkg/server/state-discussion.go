package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
)

func (h *GameHandler) createDiscussionPrompt(game *dto.GameV2) error {
	leader, err := store.GetPlayerByPosition(h.Ctx, h.DB, game.ID, game.LeaderPosition)
	if err != nil {
		return err
	}
	players, err := store.GetPlayersByGameID(h.Ctx, h.DB, game.ID)
	if err != nil {
		return err
	}
	mission, err := store.GetMissionByPriority(h.Ctx, h.DB, game.ID, game.MissionPriority)
	if err != nil {
		return err
	}
	missions, err := store.GetMissionsByGameID(h.Ctx, h.DB, game.ID)
	if err != nil {
		return err
	}

	if game.SpeakerPosition == game.LeaderPosition {
		return store.CreatePrompt(h.Ctx, h.DB, &dto.Prompt{
			GameID: game.ID,
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
	} else {
		speaker, err := store.GetPlayerByPosition(h.Ctx, h.DB, game.ID, game.SpeakerPosition)

		if err != nil {
			return err
		}
		err = store.CreatePrompt(h.Ctx, h.DB, &dto.Prompt{
			GameID: game.ID,
			Model:  speaker.Model,
			SystemPrompt: prompts.GetSystemPrompt(
				prompts.SystemPromptProps{
					Name:     speaker.Name,
					Mood:     speaker.Mood,
					Players:  players,
					Roles:    presets.Roles5, // TODO: надо брать из базы
					Role:     speaker.Role,
					Missions: missions,
				},
			),
			MessagePrompt: prompts.RenderCommentPrompt(prompts.VoteProps{
				Mission: *mission,
				Leader:  leader.Model,
				Team:    "", //TODO: extract
			}),
		})
	}
	return err
}

func (h *GameHandler) applyDiscussionPrompt(game *dto.GameV2, prompt *dto.Prompt) error {
	prompt.Status = constants.STATUS_COMPLETED
	err := store.UpdatePrompt(h.Ctx, h.DB, prompt)
	if err != nil {
		return err
	}
	//TODO: добавить лог
	return nil
}

func (h *GameHandler) handleDiscussion(game *dto.GameV2, prompts []dto.Prompt) error {
	var err error
	if len(prompts) == 0 {
		err = h.createDiscussionPrompt(game)
	} else {
		switch prompts[0].Status {
		case constants.STATUS_NOT_STARTED:
			err = h.sendLlmPrompt(game, &prompts[0])
		case constants.STATUS_HAS_RESPONSE:
			err = h.approveLlmPrompt(game, &prompts[0])
		case constants.STATUS_APPROVED:
			err = h.applyDiscussionPrompt(game, prompts[0])
		}
	}
	return err
}
