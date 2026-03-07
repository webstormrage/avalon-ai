package statemachine

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
	"fmt"
	"strconv"
)

func createSpeakerVotingPrompt(h *Handler, tx store.QueryRower, gameID int) error {
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
				Players:  players,
				Roles:    presets.Roles5, // TODO: РЅР°РґРѕ Р±СЂР°С‚СЊ РёР· СЃРёСЃС‚РµРјРЅС‹С… events
				Role:     speaker.Role,
				Missions: missions,
			},
		),
		MessagePrompt: prompts.RenderVoteSquadPrompt(prompts.VoteProps{
			Mission: *mission,
			Leader:  leader.Name,
			Team:    squadEvent.Content,
		}),
	})
}

func applySpeakerVotingPrompt(h *Handler, tx store.QueryRower, gameID int) error {
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
		Hidden:   true,
	})
	if err != nil {
		return err
	}

	err = store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_SQUAD_VOTE,
		PlayerID: speaker.ID,
		Content:  prompts.ExtractVote(prompt.Response),
		Hidden:   true,
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
		votes, err := store.GetEventsByGameIDAndType(h.Ctx, tx, gameID, constants.EVENT_SQUAD_VOTE, count-1)
		if err != nil {
			return err
		}
		votesFor := 0
		votesAgainst := 0
		votesResult := ""
		for _, vote := range votes {
			voterName := vote.PlayerName
			if voterName == "" {
				voterName = "player#" + strconv.Itoa(vote.PlayerID)
			}
			if vote.Content == "Р—Рђ" {
				votesFor += 1
				votesResult += voterName + " РїСЂРѕРіРѕР»РѕСЃРѕРІР°Р» Р—Рђ\n"
			} else {
				votesAgainst += 1
				votesResult += voterName + " РїСЂРѕРіРѕР»РѕСЃРѕРІР°Р» РџР РћРўРР’\n"
			}
		}
		votesResult += fmt.Sprintf("РС‚РѕРіРё РіРѕР»РѕСЃРѕРІР°РЅРёСЏ.\nР—Р° - %d\nРџСЂРѕС‚РёРІ - %d\n", votesFor, votesAgainst)
		if votesAgainst > votesFor {
			votesResult += "РЎРѕСЃС‚Р°РІ РЅРµ РѕРґРѕР±СЂРµРЅ."
			game.SkipsCount += 1
		} else {
			votesResult += "РЎРѕСЃС‚Р°РІ РѕРґРѕР±СЂРµРЅ."
			game.SkipsCount = 0
		}
		err = store.CreateEvent(h.Ctx, tx, &dto.Event{
			GameID:   game.ID,
			Type:     constants.EVENT_SQUAD_VOTE_RESULT,
			PlayerID: speaker.ID,
			Content:  votesResult,
		})
		if err != nil {
			return err
		}

		if game.SkipsCount >= 5 {
			game.GameState = constants.STATE_RED_VICTORY
		} else if votesAgainst > votesFor {
			game.GameState = constants.STATE_DISCUSSION
			game.LeaderPosition += 1
			if game.LeaderPosition > count {
				game.LeaderPosition = 1
			}
			game.SpeakerPosition = game.LeaderPosition
		} else {
			game.GameState = constants.STATE_MISSION
		}
	}

	return store.UpdateGame(h.Ctx, tx, game)
}

func handleSpeakerVoting(h *Handler, tx store.QueryRower, gameID int) error {
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(pendingPrompts) == 0 {
		err = createSpeakerVotingPrompt(h, tx, gameID)
	} else {
		switch pendingPrompts[0].Status {
		case constants.STATUS_NOT_STARTED:
			err = sendLlmPrompt(h, tx, gameID)
		case constants.STATUS_HAS_RESPONSE:
			err = applySpeakerVotingPrompt(h, tx, gameID)
		}
	}
	return err
}
