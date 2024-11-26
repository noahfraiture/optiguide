package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func StoreFeedback(dbPool *pgxpool.Pool, ctx context.Context, feedback string, user User) error {
	query := `INSERT INTO feedbacks (content, user_id, created_at) VALUES (@feedback, @user_id, NOW())`
	args := pgx.NamedArgs{"feedback": feedback, "user_id": user.ID}
	_, err := dbPool.Exec(ctx, query, args)
	return err
}
