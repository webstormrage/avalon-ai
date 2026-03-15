package actionhandlers

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"context"
	"fmt"
)

func ApplyProposeSquad(ctx context.Context, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(ctx, tx, gameID)
	if err != nil {
		return err
	}
	if len(game.TurnsOrder) == 0 {
		return fmt.Errorf("turnsOrder is empty")
	}
	currentTurnPosition := game.TurnsOrder[0]
	currentTurnPlayer, err := store.GetPlayerByPosition(ctx, tx, game.ID, currentTurnPosition)
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
	if err := store.CreateEvent(ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_PLAYER_SPEECH,
		PlayerID: currentTurnPlayer.ID,
		Content:  getActionMessage(action.Params),
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
	mission.Squad = action.Params.SquadNumbers
	if err := store.UpdateMission(ctx, tx, mission); err != nil {
		return err
	}
	game.TurnsOrder = append([]int(nil), game.TurnsOrder[1:]...)
	return store.UpdateGame(ctx, tx, game)
}
