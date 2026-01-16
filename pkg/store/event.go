package store

import (
	"avalon/pkg/dto"
	"context"
	"fmt"
)

func CreateEvent(
	ctx context.Context,
	tx QueryRower,
	event *dto.Event,
) error {

	err := tx.QueryRowContext(ctx, `
        INSERT INTO events (
            game_id,
            source,
            type,               
            content             
        )
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `,
		event.GameID,
		event.Source,
		event.Type,
		event.Content,
	).Scan(&event.ID)

	if err != nil {
		return fmt.Errorf("create game: %w", err)
	}

	return nil
}
