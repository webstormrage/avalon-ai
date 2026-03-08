package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"fmt"
	"strconv"
	"strings"
)

func applyProposeSquad(h *GameHandler, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	leader, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.LeaderPosition)
	if err != nil {
		return err
	}
	squad, err := squadNumbersToStrings(action.Params.SquadNumbers)
	if err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_SQUAD_DECLARATION,
		PlayerID: leader.ID,
		Content:  strings.Join(squad, ", "),
	}); err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: leader.ID,
		Content:  getActionMessage(action.Params),
	}); err != nil {
		return err
	}
	count, err := store.CountPlayersByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}
	game.SpeakerPosition++
	if game.SpeakerPosition > count {
		game.SpeakerPosition = 1
	}
	return store.UpdateGame(h.Ctx, tx, game)
}

func applyRateSquad(h *GameHandler, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.SpeakerPosition)
	if err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: speaker.ID,
		Content:  getActionMessage(action.Params),
	}); err != nil {
		return err
	}
	count, err := store.CountPlayersByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}
	game.SpeakerPosition++
	if game.SpeakerPosition > count {
		game.SpeakerPosition = 1
	}
	if game.SpeakerPosition == game.LeaderPosition {
		game.GameState = constants.STATE_VOTING
	}
	return store.UpdateGame(h.Ctx, tx, game)
}

func applyAnnounceSquad(h *GameHandler, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	leader, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.LeaderPosition)
	if err != nil {
		return err
	}
	squad, err := squadNumbersToStrings(action.Params.SquadNumbers)
	if err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_SQUAD_STATEMENT,
		PlayerID: leader.ID,
		Content:  strings.Join(squad, ", "),
	}); err != nil {
		return err
	}
	roster := squad
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_SQUAD_ROSTER,
		PlayerID: leader.ID,
		Content:  strings.Join(roster, ", "),
		Hidden:   true,
	}); err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: leader.ID,
		Content:  getActionMessage(action.Params),
	}); err != nil {
		return err
	}
	count, err := store.CountPlayersByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}
	game.SpeakerPosition++
	if game.SpeakerPosition > count {
		game.SpeakerPosition = 1
	}
	return store.UpdateGame(h.Ctx, tx, game)
}

func applyVoteSquad(h *GameHandler, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, game.SpeakerPosition)
	if err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: speaker.ID,
		Content:  getActionMessage(action.Params),
		Hidden:   true,
	}); err != nil {
		return err
	}
	approve, err := actionApproveToVote(action.Params)
	if err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_SQUAD_VOTE,
		PlayerID: speaker.ID,
		Content:  approve,
		Hidden:   true,
	}); err != nil {
		return err
	}
	count, err := store.CountPlayersByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}
	game.SpeakerPosition++
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
		for _, vote := range votes {
			if vote.Content == "ЗА" {
				votesFor++
			} else {
				votesAgainst++
			}
		}
		votesResult := fmt.Sprintf("Итоги голосования.\nЗа - %d\nПротив - %d\n", votesFor, votesAgainst)
		if votesAgainst > votesFor {
			votesResult += "Состав не одобрен."
			game.SkipsCount++
		} else {
			votesResult += "Состав одобрен."
			game.SkipsCount = 0
		}
		if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
			GameID:   game.ID,
			Type:     constants.EVENT_SQUAD_VOTE_RESULT,
			PlayerID: speaker.ID,
			Content:  votesResult,
		}); err != nil {
			return err
		}
		if game.SkipsCount >= 5 {
			game.GameState = constants.STATE_RED_VICTORY
		} else if votesAgainst > votesFor {
			game.GameState = constants.STATE_DISCUSSION
			game.LeaderPosition++
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

func applyMissionAction(h *GameHandler, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, gameID, game.SpeakerPosition)
	if err != nil {
		return err
	}
	success, err := actionSuccessToMissionResult(action.Params)
	if err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_MISSION_RESULT,
		PlayerID: speaker.ID,
		Content:  success,
		Hidden:   true,
	}); err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: speaker.ID,
		Content:  getActionMessage(action.Params),
		Hidden:   true,
	}); err != nil {
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
	}

	votes, err := store.GetEventsByGameIDAndType(h.Ctx, tx, gameID, constants.EVENT_PLAYER_MISSION_RESULT, mission.SquadSize)
	if err != nil {
		return err
	}
	votesFor := 0
	votesAgainst := 0
	for _, vote := range votes {
		if vote.Content == "УСПЕХ" {
			votesFor++
		} else {
			votesAgainst++
		}
	}
	votesResult := fmt.Sprintf("Итоги миссии.\nУспех - %d\nПровал - %d\n", votesFor, votesAgainst)
	if votesAgainst > mission.MaxFails {
		game.Fails++
	} else {
		game.Wins++
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_SQUAD_MISSION_RESULT,
		PlayerID: speaker.ID,
		Content:  votesResult,
	}); err != nil {
		return err
	}
	game.MissionPriority++
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
		game.LeaderPosition++
		if game.LeaderPosition > count {
			game.LeaderPosition = 1
		}
		game.SpeakerPosition = game.LeaderPosition
	}
	return store.UpdateGame(h.Ctx, tx, game)
}

func applyAnnounceAssassination(h *GameHandler, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, gameID, game.SpeakerPosition)
	if err != nil {
		return err
	}
	if action.Params.Target == nil {
		return fmt.Errorf("missing params.target")
	}
	targetPos := *action.Params.Target
	target, err := store.GetPlayerByPosition(h.Ctx, tx, gameID, targetPos)
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("unknown target position: %d", targetPos)
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_ASSASSINATION,
		PlayerID: speaker.ID,
		Content:  target.Name,
	}); err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: speaker.ID,
		Content:  getActionMessage(action.Params),
	}); err != nil {
		return err
	}
	if target.Role == constants.ROLE_MERLIN {
		game.GameState = constants.STATE_RED_VICTORY
	} else {
		game.GameState = constants.STATE_BLUE_VICTORY
	}
	return store.UpdateGame(h.Ctx, tx, game)
}

func getActionMessage(params GameActionParams) string {
	if params.Message == nil {
		return ""
	}
	return *params.Message
}

func squadNumbersToStrings(squadNumbers []int) ([]string, error) {
	if len(squadNumbers) == 0 {
		return nil, fmt.Errorf("params.squadNumbers must be non-empty")
	}
	roster := make([]string, 0, len(squadNumbers))
	for _, number := range squadNumbers {
		if number <= 0 {
			return nil, fmt.Errorf("invalid squad number: %d", number)
		}
		roster = append(roster, strconv.Itoa(number))
	}
	return roster, nil
}

func actionApproveToVote(params GameActionParams) (string, error) {
	if params.Approve == nil {
		return "", fmt.Errorf("params.approve must be boolean")
	}
	if *params.Approve {
		return "Р—Рђ", nil
	}
	return "РџР РћРўРР’", nil
}

func actionSuccessToMissionResult(params GameActionParams) (string, error) {
	if params.Success == nil {
		return "", fmt.Errorf("params.success must be boolean")
	}
	if *params.Success {
		return "РЈРЎРџР•РҐ", nil
	}
	return "РџР РћР’РђР›", nil
}
