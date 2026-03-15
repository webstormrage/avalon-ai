package actionhandlers

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"context"
	"fmt"
)

func ApplyVoteSquad(ctx context.Context, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(game.TurnsOrder) == 0 {
		return fmt.Errorf("turnsOrder is empty")
	}
	currentTurnPosition := game.TurnsOrder[0]
	speaker, err := store.GetPlayerByPosition(ctx, tx, game.ID, currentTurnPosition)
	if err != nil {
		return err
	}
	if speaker == nil {
		return fmt.Errorf("player not found at position %d", currentTurnPosition)
	}
	if action.PlayerID != speaker.ID {
		return fmt.Errorf("action.playerId %d is not current turn player %d", action.PlayerID, speaker.ID)
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
	approve, err := actionApproveToVote(action.Params)
	if err != nil {
		return err
	}
	playerVote := "AGAINST"
	if action.Params.Approve != nil && *action.Params.Approve {
		playerVote = "FOR"
	}
	if err := store.UpdatePlayerActionFields(ctx, tx, speaker.ID, &playerVote, nil); err != nil {
		return err
	}
	if err := store.CreateEvent(ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_SQUAD_VOTE,
		PlayerID: speaker.ID,
		Content:  approve,
		Hidden:   true,
	}); err != nil {
		return err
	}

	game.TurnsOrder = append([]int(nil), game.TurnsOrder[1:]...)
	if len(game.TurnsOrder) > 0 {
		return store.UpdateGame(ctx, tx, game)
	}

	mission, err := findCurrentMission(ctx, tx, game.ID)
	if err != nil {
		return err
	}
	if mission == nil {
		return fmt.Errorf("active mission not found")
	}
	players, err := store.GetPlayersByGameID(ctx, tx, game.ID)
	if err != nil {
		return err
	}
	votesFor := 0
	votesAgainst := 0
	for _, p := range players {
		switch p.Vote {
		case "FOR":
			votesFor++
		case "AGAINST":
			votesAgainst++
		}
	}
	votesResult := fmt.Sprintf("Итоги голосования.\nЗа - %d\nПротив - %d\n", votesFor, votesAgainst)
	if votesAgainst <= votesFor {
		votesResult += "Состав одобрен."
		game.Phase = constants.STATE_MISSION
		game.SkipsCount = 0
		game.TurnsOrder = append([]int(nil), mission.Squad...)
		if len(mission.Squad) > 0 {
			game.SpeakerPosition = mission.Squad[0]
		}
	} else {
		votesResult += "Состав не одобрен."
		mission.Skips++
		game.SkipsCount = mission.Skips
		if mission.Skips < 5 {
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
			game.Phase = constants.STATE_DISCUSSION
			game.TurnsOrder = nextTurnsOrder
			game.LeaderPosition = nextLeader
			game.SpeakerPosition = nextLeader
			mission.Squad = []int{}
			mission.Leader = nextLeader
			if err := store.ClearPlayersVoteByGameID(ctx, tx, game.ID); err != nil {
				return err
			}
		} else {
			game.Phase = constants.STATE_RED_VICTORY
		}
	}
	if err := store.CreateEvent(ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_SQUAD_VOTE_RESULT,
		PlayerID: speaker.ID,
		Content:  votesResult,
	}); err != nil {
		return err
	}
	if err := store.UpdateMission(ctx, tx, mission); err != nil {
		return err
	}
	if game.Phase != constants.STATE_DISCUSSION {
		if err := store.ClearPlayersVoteByGameID(ctx, tx, game.ID); err != nil {
			return err
		}
	}
	return store.UpdateGame(ctx, tx, game)
}
