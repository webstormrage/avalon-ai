package store

import (
	"avalon/pkg/dto"
	"context"
	"database/sql"
	"fmt"
)

func CreateMission(
	ctx context.Context,
	tx *sql.Tx,
	mission *dto.MissionV2,
) error {

	err := tx.QueryRowContext(ctx, `
        INSERT INTO missions (name, max_fails, priority, squad_size, game_id)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id
    `,
		mission.Name,
		mission.MaxFails,
		mission.Priority,
		mission.SquadSize,
		mission.GameID,
	).Scan(&mission.ID)

	if err != nil {
		return fmt.Errorf("create mission: %w", err)
	}

	return nil
}

func GetMissionByPriority(
	ctx context.Context,
	db *sql.DB,
	gameID int,
	priority int,
) (*dto.MissionV2, error) {

	var m dto.MissionV2

	err := db.QueryRowContext(ctx, `
        SELECT
            id,
            name,
            max_fails,
            priority,
            squad_size,
            game_id
        FROM missions
        WHERE game_id = $1
          AND priority = $2
    `, gameID, priority).Scan(
		&m.ID,
		&m.Name,
		&m.MaxFails,
		&m.Priority,
		&m.SquadSize,
		&m.GameID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // или domain error
		}
		return nil, err
	}

	return &m, nil
}
