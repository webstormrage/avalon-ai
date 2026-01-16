package store

import (
	"avalon/pkg/dto"
	"context"
	"database/sql"
	"fmt"
)

func CreatePrompt(
	ctx context.Context,
	db *sql.DB,
	prompt *dto.Prompt,
) error {

	err := db.QueryRowContext(ctx, `
        INSERT INTO prompts (
            game_id,
            model,
            system_prompt,
            message_prompt
        )
        VALUES ($1, $2, $3, $4)
        RETURNING id
    `,
		prompt.GameID,
		prompt.Model,
		prompt.SystemPrompt,
		prompt.MessagePrompt,
	).Scan(&prompt.ID)

	if err != nil {
		return fmt.Errorf("create prompt: %w", err)
	}

	return nil
}

func GetPromptsNotCompletedByGameID(
	ctx context.Context,
	db *sql.DB,
	gameID int,
) ([]dto.Prompt, error) {

	rows, err := db.QueryContext(ctx, `
        SELECT
            id,
            game_id,
            model,
            system_prompt,
            message_prompt,
            response,
            status
        FROM prompts
        WHERE game_id = $1
          AND status != 'COMPLETED'
        ORDER BY id
    `, gameID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	prompts := make([]dto.Prompt, 0)

	for rows.Next() {
		var p dto.Prompt
		if err := rows.Scan(
			&p.ID,
			&p.GameID,
			&p.Model,
			&p.SystemPrompt,
			&p.MessagePrompt,
			&p.Response,
			&p.Status,
		); err != nil {
			return nil, err
		}
		prompts = append(prompts, p)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prompts, nil
}

func UpdatePrompt(
	ctx context.Context,
	db *sql.DB,
	prompt *dto.Prompt,
) error {

	res, err := db.ExecContext(ctx, `
        UPDATE prompts
        SET
            model = $1,
            system_prompt = $2,
            message_prompt = $3,
            status = $4,
            response = $5
        WHERE id = $6
    `,
		prompt.Model,
		prompt.SystemPrompt,
		prompt.MessagePrompt,
		prompt.Status,
		prompt.Response,
		prompt.ID,
	)
	if err != nil {
		return fmt.Errorf("update prompt: %w", err)
	}

	rows, err := res.RowsAffected()
	if err == nil && rows == 0 {
		return sql.ErrNoRows
	}

	return nil
}
