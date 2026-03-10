package server

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"encoding/json"
	"fmt"
	"strconv"
)

func applyProposeSquad(h *GameHandler, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(game.TurnsOrder) == 0 {
		return fmt.Errorf("turnsOrder is empty")
	}
	currentTurnPosition := game.TurnsOrder[0]
	currentTurnPlayer, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, currentTurnPosition)
	if err != nil {
		return err
	}
	if currentTurnPlayer == nil {
		return fmt.Errorf("player not found at position %d", currentTurnPosition)
	}
	if action.PlayerID != currentTurnPlayer.ID {
		return fmt.Errorf("action.playerId %d is not current turn player %d", action.PlayerID, currentTurnPlayer.ID)
	}
	if _, err := squadNumbersToStrings(action.Params.SquadNumbers); err != nil {
		return err
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: currentTurnPlayer.ID,
		Content:  getActionMessage(action.Params),
	}); err != nil {
		return err
	}
	mission, err := store.GetMissionByPriority(h.Ctx, tx, game.ID, game.MissionPriority)
	if err != nil {
		return err
	}
	if mission == nil {
		return fmt.Errorf("mission not found for priority %d", game.MissionPriority)
	}
	mission.Squad = action.Params.SquadNumbers
	if err := store.UpdateMission(h.Ctx, tx, mission); err != nil {
		return err
	}
	game.TurnsOrder = append([]int(nil), game.TurnsOrder[1:]...)
	return store.UpdateGame(h.Ctx, tx, game)
}

func applyRateSquad(h *GameHandler, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(h.Ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(game.TurnsOrder) == 0 {
		return fmt.Errorf("turnsOrder is empty")
	}
	currentTurnPosition := game.TurnsOrder[0]
	speaker, err := store.GetPlayerByPosition(h.Ctx, tx, game.ID, currentTurnPosition)
	if err != nil {
		return err
	}
	if speaker == nil {
		return fmt.Errorf("player not found at position %d", currentTurnPosition)
	}
	if action.PlayerID != speaker.ID {
		return fmt.Errorf("action.playerId %d is not current turn player %d", action.PlayerID, speaker.ID)
	}
	if err := store.CreateEvent(h.Ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: speaker.ID,
		Content:  getActionMessage(action.Params),
	}); err != nil {
		return err
	}
	game.TurnsOrder = append([]int(nil), game.TurnsOrder[1:]...)
	if len(game.TurnsOrder) > 0 {
		return store.UpdateGame(h.Ctx, tx, game)
	}

	count, err := store.CountPlayersByGameID(h.Ctx, tx, game.ID)
	if err != nil {
		return err
	}
	nextLeaderPos := currentTurnPosition + 1
	if nextLeaderPos > count {
		nextLeaderPos = 1
	}
	game.LeaderPosition = nextLeaderPos
	game.SpeakerPosition = nextLeaderPos
	game.GameState = constants.STATE_VOTING
	return persistGameWithTurnsOrder(h, tx, game)
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
	if _, err := squadNumbersToStrings(action.Params.SquadNumbers); err != nil {
		return err
	}
	mission, err := store.GetMissionByPriority(h.Ctx, tx, game.ID, game.MissionPriority)
	if err != nil {
		return err
	}
	if mission == nil {
		return fmt.Errorf("mission not found for priority %d", game.MissionPriority)
	}
	mission.Squad = action.Params.SquadNumbers
	mission.Progress = 0
	mission.Fails = 0
	mission.Successes = 0
	mission.Votes = []byte("[]")
	if err := store.UpdateMission(h.Ctx, tx, mission); err != nil {
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
	return persistGameWithTurnsOrder(h, tx, game)
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
	playerVote := "AGAINST"
	if action.Params.Approve != nil && *action.Params.Approve {
		playerVote = "FOR"
	}
	if err := store.UpdatePlayerActionFields(h.Ctx, tx, speaker.ID, &playerVote, nil); err != nil {
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
		mission, err := store.GetMissionByPriority(h.Ctx, tx, game.ID, game.MissionPriority)
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
			if mission != nil {
				mission.Skips++
				if err := store.UpdateMission(h.Ctx, tx, mission); err != nil {
					return err
				}
			}
		} else {
			votesResult += "Состав одобрен."
			game.SkipsCount = 0
			if mission != nil && len(mission.Squad) > 0 {
				game.SpeakerPosition = mission.Squad[0]
			}
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
		if err := store.ClearPlayersVoteByGameID(h.Ctx, tx, game.ID); err != nil {
			return err
		}
	}
	return persistGameWithTurnsOrder(h, tx, game)
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
	playerMissionAction := "APPROVE"
	if action.Params.Success != nil && *action.Params.Success {
		playerMissionAction = "SUCCESS"
	}
	if err := store.UpdatePlayerActionFields(h.Ctx, tx, speaker.ID, nil, &playerMissionAction); err != nil {
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
	if mission == nil {
		return fmt.Errorf("mission not found for priority %d", game.MissionPriority)
	}
	if mission.Progress >= len(mission.Squad) {
		return fmt.Errorf("mission squad exhausted")
	}
	if speaker.Position != mission.Squad[mission.Progress] {
		return fmt.Errorf("current speaker is not in mission squad order")
	}

	type missionVote struct {
		PlayerID int  `json:"playerId"`
		Success  bool `json:"success"`
	}
	votesHistory := make([]missionVote, 0)
	if len(mission.Votes) > 0 {
		_ = json.Unmarshal(mission.Votes, &votesHistory)
	}
	if action.Params.Success == nil {
		return fmt.Errorf("params.success must be boolean")
	}
	if *action.Params.Success {
		mission.Successes++
	} else {
		mission.Fails++
	}
	votesHistory = append(votesHistory, missionVote{
		PlayerID: speaker.ID,
		Success:  *action.Params.Success,
	})
	votesRaw, err := json.Marshal(votesHistory)
	if err != nil {
		return err
	}
	mission.Votes = votesRaw
	mission.Progress++

	if mission.Progress < len(mission.Squad) {
		game.SpeakerPosition = mission.Squad[mission.Progress]
		if err := store.UpdateMission(h.Ctx, tx, mission); err != nil {
			return err
		}
		return persistGameWithTurnsOrder(h, tx, game)
	}

	votesResult := fmt.Sprintf("Итоги миссии.\nУспех - %d\nПровал - %d\n", mission.Successes, mission.Fails)
	if mission.Fails > mission.MaxFails {
		game.Fails++
	} else {
		game.Wins++
	}
	if err := store.UpdateMission(h.Ctx, tx, mission); err != nil {
		return err
	}
	if err := store.ClearPlayersMissionActionByGameID(h.Ctx, tx, game.ID); err != nil {
		return err
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
	return persistGameWithTurnsOrder(h, tx, game)
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
	return persistGameWithTurnsOrder(h, tx, game)
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
