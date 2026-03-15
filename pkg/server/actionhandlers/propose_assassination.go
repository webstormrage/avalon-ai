package actionhandlers

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"context"
	"fmt"
)

func ApplyProposeAssassination(ctx context.Context, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(game.TurnsOrder) == 0 {
		return fmt.Errorf("turnsOrder is empty")
	}
	currentTurnPosition := game.TurnsOrder[0]
	currentSpeaker, err := store.GetPlayerByPosition(ctx, tx, game.ID, currentTurnPosition)
	if err != nil {
		return err
	}
	if currentSpeaker == nil {
		return fmt.Errorf("player not found at position %d", currentTurnPosition)
	}
	if action.PlayerID != currentSpeaker.ID {
		return fmt.Errorf("action.playerId %d is not current turn player %d", action.PlayerID, currentSpeaker.ID)
	}
	if action.Params.Target == nil {
		return fmt.Errorf("missing params.target")
	}
	targetPos := *action.Params.Target
	target, err := store.GetPlayerByPosition(ctx, tx, game.ID, targetPos)
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("unknown target position: %d", targetPos)
	}
	if err := store.CreateEvent(ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: target.ID,
		Content:  getActionMessage(action.Params),
	}); err != nil {
		return err
	}
	game.TurnsOrder = append([]int(nil), game.TurnsOrder[1:]...)
	return store.UpdateGame(ctx, tx, game)
}
