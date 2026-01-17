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
            content,
            hidden
        )
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `,
		event.GameID,
		event.Source,
		event.Type,
		event.Content,
		event.Hidden,
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
			hidden
		FROM events
		WHERE game_id = $1
		  AND hidden = FALSE
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
			&event.Hidden,
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
			hidden
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
		&event.Hidden,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // событий такого type нет
		}
		return nil, fmt.Errorf("get last event by game id and type: %w", err)
	}

	return &event, nil
}

func GetEventsByGameIDAndType(
	ctx context.Context,
	tx QueryRower,
	gameID int,
	eventType string,
	limit int,
) ([]*dto.Event, error) {

	rows, err := tx.QueryContext(ctx, `
		SELECT
			id,
			game_id,
			source,
			type,
			content,
			hidden
		FROM (
			SELECT
				id,
				game_id,
				source,
				type,
				content,
				hidden,
				created_at
			FROM events
			WHERE game_id = $1
			  AND type = $2
			ORDER BY created_at DESC
			LIMIT $3
		) t
		ORDER BY created_at ASC
	`, gameID, eventType, limit)
	if err != nil {
		return nil, fmt.Errorf("get last n events by game id and type: %w", err)
	}
	defer rows.Close()

	var events []*dto.Event

	for rows.Next() {
		var event dto.Event
		if err := rows.Scan(
			&event.ID,
			&event.GameID,
			&event.Source,
			&event.Type,
			&event.Content,
			&event.Hidden,
		); err != nil {
			return nil, fmt.Errorf("scan event: %w", err)
		}

		events = append(events, &event)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate events: %w", err)
	}

	return events, nil
}
