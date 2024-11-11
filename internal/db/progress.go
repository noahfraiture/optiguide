package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func tableProgress(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS progress(
			user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
			card_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
			box_index INTEGER NOT NULL,
			done BOOLEAN NOT NULL,
			PRIMARY KEY(user_id, card_id, box_index)
		);`,
	)
	return err
}
func ToggleProgress(db *pgxpool.Pool, userID string, cardID, boxIndex int) error {
	query :=
		`INSERT INTO progress (user_id, card_id, box_index, done)
		VALUES (@user_id, @card_id, @box_index, true)
		ON CONFLICT (user_id, card_id, box_index)
		DO UPDATE SET done = NOT progress.done;`
	args := pgx.NamedArgs{
		"user_id":   userID,
		"card_id":   cardID,
		"box_index": boxIndex,
	}
	_, err := db.Exec(context.Background(), query, args)
	return err
}
