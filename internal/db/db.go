package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var dbpool *pgxpool.Pool

// Define the db connection pool. Don't forget to close !
func Init() error {
	var err error
	connStr := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		"postgres",
		"5432",
		os.Getenv("POSTGRES_DB"),
	)
	fmt.Println(connStr)
	dbpool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = tableUser()
	if err != nil {
		return err
	}
	err = tableProgress()
	if err != nil {
		return err
	}

	return nil
}

func Close() {
	dbpool.Close()
}

type User struct {
	ID    string
	Email string
}

func tableUser() error {
	_, err := dbpool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS users(
			id TEXT PRIMARY KEY,
			email TEXT
		);`,
	)
	return err
}

func InsertUser(user User) error {
	query := `INSERT INTO users (id, email) VALUES (@id, @email) ON CONFLICT (id) DO NOTHING;`
	args := pgx.NamedArgs{
		"id":    user.ID,
		"email": user.Email,
	}
	_, err := dbpool.Exec(context.Background(), query, args)
	return err
}

func QueryUser(id string) (User, error) {
	query := `SELECT id, email FROM users WHERE id = @id;`
	args := pgx.NamedArgs{"id": id}
	row := dbpool.QueryRow(context.Background(), query, args)
	user := User{}
	err := row.Scan(&user.ID, &user.Email)
	return user, err
}

type Progress struct {
	UserID string
	StepID int
	Done   bool
}

func tableProgress() error {
	_, err := dbpool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS progress(
			user_id TEXT,
			step_id INTEGER,
			done BOOLEAN NOT NULL,
			PRIMARY KEY(user_id, step_id)
		);`,
	)
	return err
}
func (u *User) ToggleProgress(stepID int) error {
	query := `INSERT INTO progress (user_id, step_id, done)
	VALUES (@user_id, @step_id, true)
	ON CONFLICT (user_id, step_id)
	DO UPDATE SET done = NOT (SELECT done FROM progress WHERE user_id=@user_id AND step_id=@step_id);`
	args := pgx.NamedArgs{
		"user_id": u.ID,
		"step_id": stepID,
	}
	_, err := dbpool.Exec(context.Background(), query, args)
	return err
}

func (u *User) IsStepDone(stepID int) (bool, error) {
	query := `SELECT COALESCE ((SELECT done FROM progress WHERE user_id = @user_id AND step_id = @step_id), false);`
	args := pgx.NamedArgs{
		"user_id": u.ID,
		"step_id": stepID,
	}
	row := dbpool.QueryRow(context.Background(), query, args)
	var done bool
	err := row.Scan(&done)
	return done, err
}
