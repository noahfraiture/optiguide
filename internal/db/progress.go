package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ToggleProgress(db *pgxpool.Pool, user User, cardID, boxIndex int) error {
	// Start the transaction
	tx, err := db.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	queryUpsert := `
    WITH upsert AS (
        INSERT INTO progress (user_id, card_id, box_index, done)
        VALUES (@user_id, @card_id, @box_index, true)
        ON CONFLICT (user_id, card_id, box_index)
        DO UPDATE SET done = NOT progress.done
        RETURNING done
    )
    SELECT done FROM upsert;`

	var newDoneValue bool
	err = tx.QueryRow(context.Background(), queryUpsert, pgx.NamedArgs{
		"user_id":   user.ID,
		"card_id":   cardID,
		"box_index": boxIndex,
	}).Scan(&newDoneValue)
	if err != nil {
		return fmt.Errorf("error toggling progress: %w", err)
	}

	queryProgressChange := `
	SELECT CASE
	    WHEN @done THEN 100.0 / (
			SELECT COUNT(*) * users.team_size
			FROM cards, users
			WHERE users.id = @user_id
			GROUP BY users.team_size
	    )
	    ELSE -100.0 / (
			SELECT COUNT(*) * users.team_size
			FROM cards, users
			WHERE users.id = @user_id
			GROUP BY users.team_size
	    )
		END AS change;`

	var progressChange float64
	err = tx.QueryRow(
		context.Background(),
		queryProgressChange,
		pgx.NamedArgs{
			"done":    newDoneValue,
			"user_id": user.ID,
		},
	).
		Scan(&progressChange)

	if err != nil {
		return fmt.Errorf("error calculating progress change: %w", err)
	}

	queryUpdateUser := `UPDATE users SET progress = progress + $1 WHERE id = $2;`
	_, err = tx.Exec(context.Background(), queryUpdateUser, progressChange, user.ID)
	if err != nil {
		return fmt.Errorf("error updating user progress: %w", err)
	}

	err = tx.Commit(context.Background())
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}
