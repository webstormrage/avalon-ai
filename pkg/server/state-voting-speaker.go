package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
	"fmt"
)

func (h *GameHandler) createSpeakerVotingPrompt(tx store.QueryRower, gameID int) error {
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

	squadEvent, err := store.GetLastEventByGameIDAndType(h.Ctx, tx, gameID, constants.EVENT_SQUAD_STATEMENT)
	if err != nil {
		return err
	}
	return store.CreatePrompt(h.Ctx, tx, &dto.Prompt{
		GameID: game.ID,
		Model:  speaker.Model,
		SystemPrompt: prompts.GetSystemPrompt(
			prompts.SystemPromptProps{
				Name:     speaker.Name,
				Mood:     speaker.Mood,
				Players:  players,
				Roles:    presets.Roles5, // TODO: надо брать из системных events
				Role:     speaker.Role,
				Missions: missions,
			},
		),
		MessagePrompt: prompts.RenderVotePrompt(prompts.VoteProps{
			Mission: *mission,
			Leader:  leader.Name,
			Team:    squadEvent.Content,
		}),
	})
}

func (h *GameHandler) applySpeakerVotingPrompt(tx store.QueryRower, gameID int) error {
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
		GameID:  game.ID,
		Type:    constants.EVENT_PLAYER_SPEECH,
		Source:  speaker.Name,
		Content: prompt.Response,
		Hidden:  true,
	})
	if err != nil {
		return err
	}

	err = store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:  game.ID,
		Type:    constants.EVENT_SQUAD_VOTE,
		Source:  speaker.Name,
		Content: prompts.ExtractVote(prompt.Response),
		Hidden:  true,
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
		votes, err := store.GetEventsByGameIDAndType(h.Ctx, tx, gameID, constants.EVENT_SQUAD_VOTE)
		if err != nil {
			return err
		}
		votesFor := 0
		votesAgainst := 0
		votesResult := ""
		for _, vote := range votes {
			if vote.Content == "ЗА" {
				votesFor += 1
				votesResult += vote.Source + " проголосовал ЗА\n"
			} else {
				votesAgainst += 1
				votesResult += vote.Source + " проголосовал ПРОТИВ\n"
			}
		}
		votesResult += fmt.Sprintf("Итоги голосования.\nЗа - %d\nПротив - %d\n", votesFor, votesAgainst)
		if votesAgainst > votesFor {
			votesResult += "Состав не одобрен."
			game.SkipsCount += 1
		} else {
			votesResult += "Состав одобрен."
			game.SkipsCount = 0
		}
		err = store.CreateEvent(h.Ctx, tx, &dto.Event{
			GameID:  game.ID,
			Type:    constants.EVENT_SQUAD_VOTE_RESULT,
			Source:  speaker.Name,
			Content: votesResult,
		})
		if err != nil {
			return err
		}

		game.GameState = constants.STATE_MISSION
	}

	return store.UpdateGame(h.Ctx, tx, game)
}

func (h *GameHandler) handleSpeakerVoting(tx store.QueryRower, gameID int) error {
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(pendingPrompts) == 0 {
		err = h.createSpeakerVotingPrompt(tx, gameID)
	} else {
		switch pendingPrompts[0].Status {
		case constants.STATUS_NOT_STARTED:
			err = h.sendLlmPrompt(tx, gameID)
		case constants.STATUS_HAS_RESPONSE:
			err = h.applySpeakerVotingPrompt(tx, gameID)
		}
	}
	return err
}
