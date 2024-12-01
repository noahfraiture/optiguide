package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func ToggleProgress(db *pgxpool.Pool, user User, cardIndex, boxIndex int) error {
	// Start the transaction
	ctx := context.Background()
	tx, err := db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("error starting transaction: %w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback(ctx)
		}
	}()

	queryUpsert := `
	WITH card AS (
		SELECT * FROM cards WHERE idx = @card_index
	), upsert AS (
        INSERT INTO progress (user_id, card_id, box_index, done)
        VALUES (@user_id, (SELECT id FROM card), @box_index, true)
        ON CONFLICT (user_id, card_id, box_index)
        DO UPDATE SET done = NOT progress.done
        RETURNING done
    )
    SELECT done FROM upsert;`

	var newDoneValue bool
	err = tx.QueryRow(ctx, queryUpsert, pgx.NamedArgs{
		"user_id":    user.ID,
		"card_index": cardIndex,
		"box_index":  boxIndex,
	}).Scan(&newDoneValue)
	if err != nil {
		return fmt.Errorf("error toggling progress: %w", err)
	}

	queryProgressChange := `
	SELECT CASE
	    WHEN @done THEN 100.0 / (
			SELECT COUNT(*) * users.team_size
			FROM cards, users
			WHERE users.id = @user_id AND cards.idx IS NOT NULL
			GROUP BY users.team_size
	    )
	    ELSE -100.0 / (
			SELECT COUNT(*) * users.team_size
			FROM cards, users
			WHERE users.id = @user_id AND cards.idx IS NOT NULL
			GROUP BY users.team_size
	    )
		END AS change;`

	var progressChange float64
	err = tx.QueryRow(
		ctx,
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
	_, err = tx.Exec(ctx, queryUpdateUser, progressChange, user.ID)
	if err != nil {
		return fmt.Errorf("error updating user progress: %w", err)
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("error committing transaction: %w", err)
	}

	return nil
}

func ToggleAchievement(dbPool *pgxpool.Pool, user User, cardIndex int, achievement string) error {
	query := `
		INSERT INTO achievements (name, card_id, done, user_id)
		VALUES (@name, (SELECT id FROM cards WHERE idx = @idx), true, @user_id)
		ON CONFLICT (name, card_id, user_id)
		DO UPDATE SET done = NOT achievements.done;`

	args := pgx.NamedArgs{
		"name":    achievement,
		"idx":     cardIndex,
		"user_id": user.ID,
	}

	_, err := dbPool.Exec(context.Background(), query, args)
	if err != nil {
		return err
	}
	return nil
}
