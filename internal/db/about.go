package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func StoreFeedback(dbPool *pgxpool.Pool, ctx context.Context, feedback string, userID string) error {
	query := `INSERT INTO feedbacks (content, user_id, created_at) VALUES (@feedback, @user_id, NOW())`
	args := pgx.NamedArgs{"feedback": feedback, "user_id": userID}
	_, err := dbPool.Exec(ctx, query, args)
	return err
}
