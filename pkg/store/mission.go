package store

import (
	"avalon/pkg/dto"
	"context"
	"database/sql"
	"fmt"

	"github.com/lib/pq"
)

func CreateMission(
	ctx context.Context,
	tx QueryRower,
	mission *dto.MissionV2,
) error {

	err := tx.QueryRowContext(ctx, `
        INSERT INTO missions (name, max_fails, priority, squad_size, squad, progress, fails, successes, skips, votes, game_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING id
    `,
		mission.Name,
		mission.MaxFails,
		mission.Priority,
		mission.SquadSize,
		pq.Array(intsToInt64s(mission.Squad)),
		mission.Progress,
		mission.Fails,
		mission.Successes,
		mission.Skips,
		jsonOrEmptyArray(mission.Votes),
		mission.GameID,
	).Scan(&mission.ID)

	if err != nil {
		return fmt.Errorf("create mission: %w", err)
	}

	return nil
}

func GetMissionByPriority(
	ctx context.Context,
	db QueryRower,
	gameID int,
	priority int,
) (*dto.MissionV2, error) {

	var m dto.MissionV2
	var squad []int64

	err := db.QueryRowContext(ctx, `
        SELECT
            id,
            name,
            max_fails,
            priority,
            squad_size,
            squad,
            progress,
            fails,
            successes,
            skips,
            votes,
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
		pq.Array(&squad),
		&m.Progress,
		&m.Fails,
		&m.Successes,
		&m.Skips,
		&m.Votes,
		&m.GameID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // или domain error
		}
		return nil, err
	}
	m.Squad = int64sToInts(squad)

	return &m, nil
}

func GetMissionsByGameID(
	ctx context.Context,
	db QueryRower,
	gameID int,
) ([]dto.MissionV2, error) {

	rows, err := db.QueryContext(ctx, `
        SELECT
            id,
            name,
            max_fails,
            priority,
            squad_size,
            squad,
            progress,
            fails,
            successes,
            skips,
            votes,
            game_id
        FROM missions
        WHERE game_id = $1
        ORDER BY priority ASC
    `, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	missions := make([]dto.MissionV2, 0)

	for rows.Next() {
		var m dto.MissionV2
		var squad []int64
		if err := rows.Scan(
			&m.ID,
			&m.Name,
			&m.MaxFails,
			&m.Priority,
			&m.SquadSize,
			pq.Array(&squad),
			&m.Progress,
			&m.Fails,
			&m.Successes,
			&m.Skips,
			&m.Votes,
			&m.GameID,
		); err != nil {
			return nil, err
		}
		m.Squad = int64sToInts(squad)
		missions = append(missions, m)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return missions, nil
}

func UpdateMission(
	ctx context.Context,
	tx QueryRower,
	mission *dto.MissionV2,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE missions
		SET
			name = $1,
			max_fails = $2,
			priority = $3,
			squad_size = $4,
			squad = $5,
			progress = $6,
			fails = $7,
			successes = $8,
			skips = $9,
			votes = $10
		WHERE id = $11
	`,
		mission.Name,
		mission.MaxFails,
		mission.Priority,
		mission.SquadSize,
		pq.Array(intsToInt64s(mission.Squad)),
		mission.Progress,
		mission.Fails,
		mission.Successes,
		mission.Skips,
		jsonOrEmptyArray(mission.Votes),
		mission.ID,
	)
	if err != nil {
		return fmt.Errorf("update mission: %w", err)
	}

	return nil
}

func intsToInt64s(values []int) []int64 {
	if len(values) == 0 {
		return nil
	}

	result := make([]int64, len(values))
	for i, v := range values {
		result[i] = int64(v)
	}

	return result
}

func int64sToInts(values []int64) []int {
	if len(values) == 0 {
		return nil
	}

	result := make([]int, len(values))
	for i, v := range values {
		result[i] = int(v)
	}

	return result
}

func jsonOrEmptyArray(raw []byte) []byte {
	if len(raw) == 0 {
		return []byte("[]")
	}
	return raw
}
