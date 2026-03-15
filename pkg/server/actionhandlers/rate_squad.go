package actionhandlers

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"context"
	"fmt"
)

func ApplyRateSquad(ctx context.Context, tx store.QueryRower, gameID int, action GameAction) error {
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
	}); err != nil {
		return err
	}
	game.TurnsOrder = append([]int(nil), game.TurnsOrder[1:]...)
	if len(game.TurnsOrder) > 0 {
		return store.UpdateGame(ctx, tx, game)
	}

	count, err := store.CountPlayersByGameID(ctx, tx, game.ID)
	if err != nil {
		return err
	}
	nextLeaderPos := currentTurnPosition + 1
	if nextLeaderPos > count {
		nextLeaderPos = 1
	}
	game.Phase = constants.STATE_VOTING
	nextTurnsOrder, err := orderedRingFromPosition(nextLeaderPos, count)
	if err != nil {
		return err
	}
	game.TurnsOrder = nextTurnsOrder
	return store.UpdateGame(ctx, tx, game)
}
