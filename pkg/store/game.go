package store

import (
	"avalon/pkg/dto"
	"context"
	"database/sql"
	"fmt"
)

type QueryRower interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func CreateGame(
	ctx context.Context,
	tx QueryRower,
	game *dto.GameV2,
) error {

	err := tx.QueryRowContext(ctx, `
        INSERT INTO games (
            mission_priority,
            leader_position,
            skips_count,               
            wins,
            fails,
            game_state               
        )
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id
    `,
		game.MissionPriority,
		game.LeaderPosition,
		game.SkipsCount,
		game.Wins,
		game.Fails,
		game.GameState,
	).Scan(&game.ID)

	if err != nil {
		return fmt.Errorf("create game: %w", err)
	}

	return nil
}

func CreateGameTransaction(
	ctx context.Context,
	db *sql.DB,
	game *dto.GameV2,
	missions []*dto.MissionV2,
	players []*dto.PlayerV2,
) (int, error) {

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	// 1. create game
	if err := CreateGame(ctx, tx, game); err != nil {
		return 0, err
	}

	gameID := int(game.ID)

	// 2. create missions
	for _, m := range missions {
		m.GameID = gameID
		if err := CreateMission(ctx, tx, m); err != nil {
			return 0, err
		}
	}

	// 3. create players
	for _, p := range players {
		p.GameID = gameID
		if err := CreatePlayer(ctx, tx, p); err != nil {
			return 0, err
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}

	return gameID, nil
}

func GetGame(
	ctx context.Context,
	db QueryRower,
	gameID int,
) (*dto.GameV2, error) {

	var game dto.GameV2

	err := db.QueryRowContext(ctx, `
        SELECT
            id,
            mission_priority,
            leader_position,
            speaker_position,
            skips_count,
            wins,
            fails,
            game_state
        FROM games
        WHERE id = $1
    `, gameID).Scan(
		&game.ID,
		&game.MissionPriority,
		&game.SpeakerPosition,
		&game.LeaderPosition,
		&game.SkipsCount,
		&game.Wins,
		&game.Fails,
		&game.GameState,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // или domain error
		}
		return nil, err
	}

	return &game, nil
}
