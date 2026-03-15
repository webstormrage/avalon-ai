package actionhandlers

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"avalon/pkg/store"
	"context"
	"fmt"
)

func ApplyAnnounceAssassination(ctx context.Context, tx store.QueryRower, gameID int, action GameAction) error {
	game, err := store.GetGame(ctx, tx, gameID)
	if err != nil {
		return err
	}
	if action.Params.Target == nil {
		return fmt.Errorf("missing params.target")
	}
	targetPos := *action.Params.Target
	target, err := store.GetPlayerByPosition(ctx, tx, gameID, targetPos)
	if err != nil {
		return err
	}
	if target == nil {
		return fmt.Errorf("unknown target position: %d", targetPos)
	}
	if err := store.CreateEvent(ctx, tx, &dto.Event{
		GameID:   game.ID,
		Type:     constants.EVENT_ASSASSINATION,
		PlayerID: target.ID,
		Content:  getActionMessage(action.Params),
	}); err != nil {
		return err
	}
	if target.Role == constants.ROLE_MERLIN {
		game.Phase = constants.STATE_RED_VICTORY
	} else {
		game.Phase = constants.STATE_BLUE_VICTORY
	}
	return store.UpdateGame(ctx, tx, game)
}
