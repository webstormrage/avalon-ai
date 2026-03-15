package store

import (
	"avalon/pkg/constants"
	"avalon/pkg/dto"
	"context"
	"database/sql"
)

const playerSelect = `
	id, name, model, role, character_type, position, vote, mission_action, game_id
`

// ------------------------------------------------
// CREATE
// ------------------------------------------------

func CreatePlayer(
	ctx context.Context,
	tx QueryRower,
	p *dto.PlayerV2,
) error {

	return tx.QueryRowContext(ctx, `
        INSERT INTO players (
            name, model, role, character_type, position, vote, mission_action, game_id
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING id
    `,
		p.Name,
		p.Model,
		p.Role,
		p.CharacterType,
		p.Position,
		p.Vote,
		p.MissionAction,
		p.GameID,
	).Scan(&p.ID)
}

// ------------------------------------------------
// GET BY POSITION
// ------------------------------------------------

func GetPlayerByPosition(
	ctx context.Context,
	db QueryRower,
	gameID int,
	position int,
) (*dto.PlayerV2, error) {

	var p dto.PlayerV2
	var vote sql.NullString
	var missionAction sql.NullString

	err := db.QueryRowContext(ctx, `
        SELECT `+playerSelect+`
        FROM players
        WHERE game_id = $1 AND position = $2
    `, gameID, position).Scan(
		&p.ID,
		&p.Name,
		&p.Model,
		&p.Role,
		&p.CharacterType,
		&p.Position,
		&vote,
		&missionAction,
		&p.GameID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	assignNullableString(&p.Vote, vote)
	assignNullableString(&p.MissionAction, missionAction)

	return &p, nil
}

// ------------------------------------------------
// FIND BY NAME LIKE
// ------------------------------------------------

func FindPlayersByNameLike(
	ctx context.Context,
	db QueryRower,
	gameID int,
	namePart string,
) ([]dto.PlayerV2, error) {

	rows, err := db.QueryContext(ctx, `
        SELECT `+playerSelect+`
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
		var vote sql.NullString
		var missionAction sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Model,
			&p.Role,
			&p.CharacterType,
			&p.Position,
			&vote,
			&missionAction,
			&p.GameID,
		); err != nil {
			return nil, err
		}
		assignNullableString(&p.Vote, vote)
		assignNullableString(&p.MissionAction, missionAction)
		players = append(players, p)
	}

	return players, rows.Err()
}

// ------------------------------------------------
// GET BY ROLE
// ------------------------------------------------

func GetPlayersByRole(
	ctx context.Context,
	db QueryRower,
	gameID int,
	role string,
) ([]dto.PlayerV2, error) {

	rows, err := db.QueryContext(ctx, `
        SELECT `+playerSelect+`
        FROM players
        WHERE game_id = $1 AND role = $2
        ORDER BY position
    `, gameID, role)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []dto.PlayerV2

	for rows.Next() {
		var p dto.PlayerV2
		var vote sql.NullString
		var missionAction sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Model,
			&p.Role,
			&p.CharacterType,
			&p.Position,
			&vote,
			&missionAction,
			&p.GameID,
		); err != nil {
			return nil, err
		}
		assignNullableString(&p.Vote, vote)
		assignNullableString(&p.MissionAction, missionAction)
		players = append(players, p)
	}

	return players, rows.Err()
}

// ------------------------------------------------
// COUNT
// ------------------------------------------------

func CountPlayersByGameID(
	ctx context.Context,
	db QueryRower,
	gameID int,
) (int, error) {

	var count int

	err := db.QueryRowContext(ctx, `
        SELECT COUNT(*)
        FROM players
        WHERE game_id = $1
          AND LOWER(role) <> LOWER($2)
    `, gameID, constants.ROLE_GAME_MASTER).Scan(&count)

	return count, err
}

// ------------------------------------------------
// GET ALL BY GAME
// ------------------------------------------------

func GetPlayersByGameID(
	ctx context.Context,
	db QueryRower,
	gameID int,
) ([]dto.PlayerV2, error) {

	rows, err := db.QueryContext(ctx, `
        SELECT `+playerSelect+`
        FROM players
        WHERE game_id = $1
        ORDER BY position ASC
    `, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var players []dto.PlayerV2

	for rows.Next() {
		var p dto.PlayerV2
		var vote sql.NullString
		var missionAction sql.NullString
		if err := rows.Scan(
			&p.ID,
			&p.Name,
			&p.Model,
			&p.Role,
			&p.CharacterType,
			&p.Position,
			&vote,
			&missionAction,
			&p.GameID,
		); err != nil {
			return nil, err
		}
		assignNullableString(&p.Vote, vote)
		assignNullableString(&p.MissionAction, missionAction)
		players = append(players, p)
	}

	return players, rows.Err()
}

// ------------------------------------------------
// GET BY ID
// ------------------------------------------------

func GetPlayerByID(
	ctx context.Context,
	db QueryRower,
	id int,
) (*dto.PlayerV2, error) {

	var p dto.PlayerV2
	var vote sql.NullString
	var missionAction sql.NullString

	err := db.QueryRowContext(ctx, `
        SELECT `+playerSelect+`
        FROM players
        WHERE id = $1
    `, id).Scan(
		&p.ID,
		&p.Name,
		&p.Model,
		&p.Role,
		&p.CharacterType,
		&p.Position,
		&vote,
		&missionAction,
		&p.GameID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	assignNullableString(&p.Vote, vote)
	assignNullableString(&p.MissionAction, missionAction)

	return &p, nil
}

func UpdatePlayerActionFields(
	ctx context.Context,
	tx QueryRower,
	playerID int,
	vote *string,
	missionAction *string,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE players
		SET vote = COALESCE($1, ''), mission_action = COALESCE($2, '')
		WHERE id = $3
	`, vote, missionAction, playerID)
	return err
}

func ClearPlayersVoteByGameID(
	ctx context.Context,
	tx QueryRower,
	gameID int,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE players
		SET vote = ''
		WHERE game_id = $1
	`, gameID)
	return err
}

func ClearPlayersMissionActionByGameID(
	ctx context.Context,
	tx QueryRower,
	gameID int,
) error {
	_, err := tx.ExecContext(ctx, `
		UPDATE players
		SET mission_action = ''
		WHERE game_id = $1
	`, gameID)
	return err
}

func assignNullableString(dst *string, src sql.NullString) {
	if !src.Valid {
		*dst = ""
		return
	}
	*dst = src.String
}
