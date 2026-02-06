package store

import (
	"avalon/pkg/dto"
	"context"
	"database/sql"
)

func CreatePlayer(
	ctx context.Context,
	tx QueryRower,
	p *dto.PlayerV2,
) error {
	err := tx.QueryRowContext(ctx, `
        INSERT INTO players (
            name, model, tts_model, role, voice, 
            voice_temperature, voice_style, mood, position, game_id
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING id
    `,
		p.Name,
		p.Model,
		p.TtsModel,
		p.Role,
		p.Voice,
		p.VoiceTemperature,
		p.VoiceStyle,
		p.Mood,
		p.Position,
		p.GameID,
	).Scan(&p.ID)

	return err
}

func GetPlayerByPosition(
	ctx context.Context,
	db QueryRower,
	gameID int,
	position int,
) (*dto.PlayerV2, error) {
	var p dto.PlayerV2

	err := db.QueryRowContext(ctx, `
        SELECT
            id, name, model, tts_model, role, voice, 
            voice_temperature, voice_style, mood, position, game_id
        FROM players
        WHERE game_id = $1 AND position = $2
    `, gameID, position).Scan(
		&p.ID,
		&p.Name,
		&p.Model,
		&p.TtsModel,
		&p.Role,
		&p.Voice,
		&p.VoiceTemperature,
		&p.VoiceStyle,
		&p.Mood,
		&p.Position,
		&p.GameID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}

func FindPlayersByNameLike(
	ctx context.Context,
	db QueryRower,
	gameID int,
	namePart string,
) ([]dto.PlayerV2, error) {
	rows, err := db.QueryContext(ctx, `
        SELECT
            id, name, model, tts_model, role, voice, 
            voice_temperature, voice_style, mood, position, game_id
        FROM players
        WHERE game_id = $1 AND name ILIKE '%' || $2 || '%'
        ORDER BY position
    `, gameID, namePart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []dto.PlayerV2
	for rows.Next() {
		var p dto.PlayerV2
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Model, &p.TtsModel, &p.Role, &p.Voice,
			&p.VoiceTemperature, &p.VoiceStyle, &p.Mood, &p.Position, &p.GameID,
		); err != nil {
			return nil, err
		}
		players = append(players, p)
	}

	return players, rows.Err()
}

func GetPlayersByRole(
	ctx context.Context,
	db QueryRower,
	gameID int,
	role string,
) ([]dto.PlayerV2, error) {
	rows, err := db.QueryContext(ctx, `
        SELECT
            id, name, model, tts_model, role, voice, 
            voice_temperature, voice_style, mood, position, game_id
        FROM players
        WHERE game_id = $1 AND role = $2
        ORDER BY position
    `, gameID, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]dto.PlayerV2, 0)
	for rows.Next() {
		var p dto.PlayerV2
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Model, &p.TtsModel, &p.Role, &p.Voice,
			&p.VoiceTemperature, &p.VoiceStyle, &p.Mood, &p.Position, &p.GameID,
		); err != nil {
			return nil, err
		}
		players = append(players, p)
	}

	return players, rows.Err()
}

func CountPlayersByGameID(
	ctx context.Context,
	db QueryRower,
	gameID int,
) (int, error) {
	var count int
	err := db.QueryRowContext(ctx, `
        SELECT COUNT(*) FROM players WHERE game_id = $1
    `, gameID).Scan(&count)
	return count, err
}

func GetPlayersByGameID(
	ctx context.Context,
	db QueryRower,
	gameID int,
) ([]dto.PlayerV2, error) {
	rows, err := db.QueryContext(ctx, `
        SELECT
            id, name, model, tts_model, role, voice, 
            voice_temperature, voice_style, mood, position, game_id
        FROM players
        WHERE game_id = $1
        ORDER BY position ASC
    `, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	players := make([]dto.PlayerV2, 0)
	for rows.Next() {
		var p dto.PlayerV2
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Model, &p.TtsModel, &p.Role, &p.Voice,
			&p.VoiceTemperature, &p.VoiceStyle, &p.Mood, &p.Position, &p.GameID,
		); err != nil {
			return nil, err
		}
		players = append(players, p)
	}

	return players, rows.Err()
}

func GetPlayerByID(
	ctx context.Context,
	db QueryRower,
	id int,
) (*dto.PlayerV2, error) {

	var p dto.PlayerV2

	err := db.QueryRowContext(ctx, `
        SELECT
            id, name, model, tts_model, role, voice, 
            voice_temperature, voice_style, mood, position, game_id
        FROM players
        WHERE id = $1
    `, id).Scan(
		&p.ID,
		&p.Name,
		&p.Model,
		&p.TtsModel,
		&p.Role,
		&p.Voice,
		&p.VoiceTemperature,
		&p.VoiceStyle,
		&p.Mood,
		&p.Position,
		&p.GameID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return &p, nil
}
