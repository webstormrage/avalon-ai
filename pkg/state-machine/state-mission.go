package statemachine

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/presets"
	"avalon/pkg/prompts"
	"avalon/pkg/store"
	"fmt"
	"strconv"
	"strings"
)

func createMissionPrompt(h *Handler, tx store.QueryRower, gameID int) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
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

	rosterEvent, err := store.GetLastEventByGameIDAndType(h.Ctx, tx, gameID, constants.EVENT_SQUAD_ROSTER)
	if err != nil {
		return err
	}
	squadNumbers := strings.Split(rosterEvent.Content, ", ")
	game.SpeakerPosition, err = strconv.Atoi(squadNumbers[0])
	if err != nil {
		return err
	}

	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, gameID, game.SpeakerPosition)
	if err != nil {
		return err
	}

	err = store.UpdateGame(h.Ctx, tx, game)

	squadEvent, err := store.GetLastEventByGameIDAndType(h.Ctx, tx, gameID, constants.EVENT_SQUAD_STATEMENT)
	if err != nil {
		return err
	}
	leaderName := squadEvent.PlayerName
	if leaderName == "" {
		leader := "player#" + strconv.Itoa(squadEvent.PlayerID)
		leaderName = leader
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
		MessagePrompt: prompts.RenderMissionActionPrompt(prompts.VoteProps{
			Mission: *mission,
			Leader:  leaderName,
			Team:    squadEvent.Content,
		}),
	})
}

func applyMissionPrompt(h *Handler, tx store.QueryRower, gameID int) error {
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

	err = store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_MISSION_RESULT,
		PlayerID: speaker.ID,
		Content:  prompts.ExtractMissionResult(prompt.Response),
		Hidden:   true,
	})
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

	mission, err := store.GetMissionByPriority(h.Ctx, tx, game.ID, game.MissionPriority)
	if err != nil {
		return err
	}

	rosterEvent, err := store.GetLastEventByGameIDAndType(h.Ctx, tx, gameID, constants.EVENT_SQUAD_ROSTER)
	if err != nil {
		return err
	}
	squadNumbers := strings.Split(rosterEvent.Content, ", ")

	if len(squadNumbers) > 1 {
		return store.CreateEvent(h.Ctx, tx, &dto.Event{
			GameID:   game.ID,
			Type:     constants.EVENT_SQUAD_ROSTER,
			PlayerID: rosterEvent.PlayerID,
			Content:  strings.Join(squadNumbers[1:], ", "),
			Hidden:   true,
		})
	} else {
		votes, err := store.GetEventsByGameIDAndType(h.Ctx, tx, gameID, constants.EVENT_PLAYER_MISSION_RESULT, mission.SquadSize)
		if err != nil {
			return err
		}
		votesFor := 0
		votesAgainst := 0
		votesResult := ""
		for _, vote := range votes {
			if vote.Content == "РЈРЎРџР•РҐ" {
				votesFor += 1
			} else {
				votesAgainst += 1
			}
		}
		votesResult += fmt.Sprintf("РС‚РѕРіРё РјРёСЃСЃРёРё.\nРЈСЃРїРµС… - %d\nРџСЂРѕРІР°Р» - %d\n", votesFor, votesAgainst)
		if votesAgainst > mission.MaxFails {
			game.Fails++
		} else {
			game.Wins++
		}
		err = store.CreateEvent(h.Ctx, tx, &dto.Event{
			GameID:   game.ID,
			Type:     constants.EVENT_SQUAD_MISSION_RESULT,
			PlayerID: speaker.ID,
			Content:  votesResult,
		})
		game.MissionPriority++
		if err != nil {
			return err
		}

		if game.Wins >= 3 {
			game.GameState = constants.STATE_ASSASSIONATION_DISCUSSION
		} else if game.Fails >= 3 {
			game.GameState = constants.STATE_RED_VICTORY
		} else {
			count, err := store.CountPlayersByGameID(h.Ctx, tx, game.ID)
			if err != nil {
				return err
			}
			game.GameState = constants.STATE_DISCUSSION
			game.LeaderPosition += 1
			if game.LeaderPosition > count {
				game.LeaderPosition = 1
			}
			game.SpeakerPosition = game.LeaderPosition
		}
		return store.UpdateGame(h.Ctx, tx, game)
	}
}

func handleMission(h *Handler, tx store.QueryRower, gameID int) error {
	pendingPrompts, err := store.GetPromptsNotCompletedByGameID(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(pendingPrompts) == 0 {
		err = createMissionPrompt(h, tx, gameID)
	} else {
		switch pendingPrompts[0].Status {
		case constants.STATUS_NOT_STARTED:
			err = sendLlmPrompt(h, tx, gameID)
		case constants.STATUS_HAS_RESPONSE:
			err = applyMissionPrompt(h, tx, gameID)
		}
	}
	return err
}
