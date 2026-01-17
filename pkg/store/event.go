package store

import (
	"avalon/pkg/dto"
	"context"
	"database/sql"
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

func GetEventsByGameID(
	ctx context.Context,
	tx QueryRower,
	gameID int,
) ([]*dto.Event, error) {

	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			game_id,
			source,
			type,
			content,
			created_at
		FROM events
		WHERE game_id = $1
		ORDER BY created_at ASC
	`, gameID)
	if err != nil {
		return nil, fmt.Errorf("get events by game id: %w", err)
	}
	defer rows.Close()

	var events []*dto.Event

	for rows.Next() {
		var event dto.Event
		err := rows.Scan(
			&event.ID,
			&event.GameID,
			&event.Source,
			&event.Type,
			&event.Content,
		)
		if err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return events, nil
}

func GetLastEventByGameIDAndType(
	ctx context.Context,
	tx QueryRower,
	gameID int,
	eventType string,
) (*dto.Event, error) {

	var event dto.Event

	err := tx.QueryRowContext(ctx, `
		SELECT
			id,
			game_id,
			source,
			type,
			content,
			created_at
		FROM events
		WHERE game_id = $1
		  AND type = $2
		ORDER BY created_at DESC
		LIMIT 1
	`, gameID, eventType).Scan(
		&event.ID,
		&event.GameID,
		&event.Source,
		&event.Type,
		&event.Content,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // событий такого type нет
		}
		return nil, fmt.Errorf("get last event by game id and type: %w", err)
	}

	return &event, nil
}
