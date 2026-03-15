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
        INSERT INTO missions (name, status, max_fails, allowed_fails, priority, squad_size, squad, progress, fails, successes, skips, leader, votes, game_id)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
        RETURNING id
    `,
		mission.Name,
		mission.Status,
		mission.MaxFails,
		effectiveAllowedFails(mission),
		mission.Priority,
		mission.SquadSize,
		pq.Array(intsToInt64s(mission.Squad)),
		mission.Progress,
		mission.Fails,
		mission.Successes,
		mission.Skips,
		mission.Leader,
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
            status,
            max_fails,
            allowed_fails,
            priority,
            squad_size,
            squad,
            progress,
            fails,
            successes,
            skips,
            leader,
            votes,
            game_id
        FROM missions
        WHERE game_id = $1
          AND priority = $2
    `, gameID, priority).Scan(
		&m.ID,
		&m.Name,
		&m.Status,
		&m.MaxFails,
		&m.AllowedFails,
		&m.Priority,
		&m.SquadSize,
		pq.Array(&squad),
		&m.Progress,
		&m.Fails,
		&m.Successes,
		&m.Skips,
		&m.Leader,
		&m.Votes,
		&m.GameID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // или domain error
		}
		return nil, err
	}
	if m.MaxFails == 0 && m.AllowedFails > 0 {
		m.MaxFails = m.AllowedFails
	}
	if m.AllowedFails == 0 && m.MaxFails > 0 {
		m.AllowedFails = m.MaxFails
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
            status,
            max_fails,
            allowed_fails,
            priority,
            squad_size,
            squad,
            progress,
            fails,
            successes,
            skips,
            leader,
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
			&m.Status,
			&m.MaxFails,
			&m.AllowedFails,
			&m.Priority,
			&m.SquadSize,
			pq.Array(&squad),
			&m.Progress,
			&m.Fails,
			&m.Successes,
			&m.Skips,
			&m.Leader,
			&m.Votes,
			&m.GameID,
		); err != nil {
			return nil, err
		}
		if m.MaxFails == 0 && m.AllowedFails > 0 {
			m.MaxFails = m.AllowedFails
		}
		if m.AllowedFails == 0 && m.MaxFails > 0 {
			m.AllowedFails = m.MaxFails
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
			status = $2,
			max_fails = $3,
			allowed_fails = $4,
			priority = $5,
			squad_size = $6,
			squad = $7,
			progress = $8,
			fails = $9,
			successes = $10,
			skips = $11,
			leader = $12,
			votes = $13
		WHERE id = $14
	`,
		mission.Name,
		mission.Status,
		mission.MaxFails,
		effectiveAllowedFails(mission),
		mission.Priority,
		mission.SquadSize,
		pq.Array(intsToInt64s(mission.Squad)),
		mission.Progress,
		mission.Fails,
		mission.Successes,
		mission.Skips,
		mission.Leader,
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

func effectiveAllowedFails(mission *dto.MissionV2) int {
	if mission.AllowedFails > 0 {
		return mission.AllowedFails
	}
	return mission.MaxFails
}
