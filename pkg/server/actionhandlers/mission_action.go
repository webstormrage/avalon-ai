package actionhandlers

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"context"
	"fmt"
)

func ApplyMissionAction(ctx context.Context, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(game.TurnsOrder) == 0 {
		return fmt.Errorf("turnsOrder is empty")
	}
	currentTurnPosition := game.TurnsOrder[0]
	speaker, err := store.GetPlayerByPosition(ctx, tx, gameID, currentTurnPosition)
	if err != nil {
		return err
	}
	if speaker == nil {
		return fmt.Errorf("player not found at position %d", currentTurnPosition)
	}
	if action.PlayerID != speaker.ID {
		return fmt.Errorf("action.playerId %d is not current turn player %d", action.PlayerID, speaker.ID)
	}
	if action.Params.Success == nil {
		return fmt.Errorf("params.success must be boolean")
	}

	success, err := actionSuccessToMissionResult(action.Params)
	if err != nil {
		return err
	}
	playerMissionAction := "FAIL"
	if *action.Params.Success {
		playerMissionAction = "SUCCESS"
	}
	if err := store.UpdatePlayerActionFields(ctx, tx, speaker.ID, nil, &playerMissionAction); err != nil {
		return err
	}
	if err := store.CreateEvent(ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_MISSION_RESULT,
		PlayerID: speaker.ID,
		Content:  success,
		Hidden:   true,
	}); err != nil {
		return err
	}
	if err := store.CreateEvent(ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: speaker.ID,
		Content:  getActionMessage(action.Params),
		Hidden:   true,
	}); err != nil {
		return err
	}

	mission, err := findCurrentMission(ctx, tx, game.ID)
	if err != nil {
		return err
	}
	if mission == nil {
		return fmt.Errorf("active mission not found")
	}

	if *action.Params.Success {
		mission.Successes++
	} else {
		mission.Fails++
	}
	mission.Progress++

	game.TurnsOrder = append([]int(nil), game.TurnsOrder[1:]...)
	if len(game.TurnsOrder) > 0 {
		if err := store.UpdateMission(ctx, tx, mission); err != nil {
			return err
		}
		return store.UpdateGame(ctx, tx, game)
	}

	allowedFails := mission.AllowedFails
	if allowedFails <= 0 {
		allowedFails = mission.MaxFails + 1
	}
	missionResult := "COMPLETED"
	if mission.Fails >= allowedFails {
		missionResult = "FAILED"
	}
	mission.Status = missionResult

	votesResult := fmt.Sprintf("Mission result.\nSuccess - %d\nFail - %d\n", mission.Successes, mission.Fails)
	if missionResult == "FAILED" {
		game.Fails++
		votesResult += "Mission failed."
	} else {
		game.Wins++
		votesResult += "Mission succeeded."
	}

	if err := store.UpdateMission(ctx, tx, mission); err != nil {
		return err
	}
	if err := store.ClearPlayersMissionActionByGameID(ctx, tx, game.ID); err != nil {
		return err
	}
	if err := store.CreateEvent(ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_SQUAD_MISSION_RESULT,
		PlayerID: speaker.ID,
		Content:  votesResult,
	}); err != nil {
		return err
	}

	if game.Fails >= 3 {
		game.Phase = constants.STATE_RED_VICTORY
		return store.UpdateGame(ctx, tx, game)
	}
	if game.Wins >= 3 {
		players, err := store.GetPlayersByGameID(ctx, tx, game.ID)
		if err != nil {
			return err
		}
		redsOrder, err := orderedRedsFromAssassin(players)
		if err != nil {
			return err
		}
		game.Phase = constants.STATE_ASSASSIONATION_DISCUSSION
		game.TurnsOrder = redsOrder
		if len(redsOrder) > 0 {
			game.SpeakerPosition = redsOrder[0]
		}
		return store.UpdateGame(ctx, tx, game)
	}

	count, err := store.CountPlayersByGameID(ctx, tx, game.ID)
	if err != nil {
		return err
	}
	nextLeader := mission.Leader + 1
	if nextLeader > count {
		nextLeader = 1
	}
	nextTurnsOrder, err := orderedRingFromPosition(nextLeader, count)
	if err != nil {
		return err
	}

	game.MissionPriority++
	game.Phase = constants.STATE_DISCUSSION
	game.LeaderPosition = nextLeader
	game.SpeakerPosition = nextLeader
	game.TurnsOrder = nextTurnsOrder
	mission.Squad = []int{}
	mission.Leader = nextLeader
	if err := store.UpdateMission(ctx, tx, mission); err != nil {
		return err
	}
	if err := store.ClearPlayersVoteByGameID(ctx, tx, game.ID); err != nil {
		return err
	}
	return store.UpdateGame(ctx, tx, game)
}
