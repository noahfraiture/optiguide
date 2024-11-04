package db

import (
	"context"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

var dbpool *pgxpool.Pool

// Define the db connection pool. Don't forget to close !
func Init() error {
	_ = godotenv.Load()
	var err error
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", os.Getenv("POSTGRES_USER"), os.Getenv("POSTGRES_PASSWORD"), "localhost", "5432", os.Getenv("POSTGRES_DB"))
	dbpool, err = pgxpool.New(context.Background(), connStr)
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

type Progress struct {
	UserID string
	StepID string
	Done   bool
}

func InsertUser(user User) error {
	query := `INSERT INTO users (id, email) VALUES (@id, @email)`
	args := pgx.NamedArgs{
		"id":    user.ID,
		"email": user.Email,
	}
	_, err := dbpool.Exec(context.Background(), query, args)
	return err
}

func QueryUser(id string) (User, error) {
	query := `SELECT id, email FROM users WHERE id = @id`
	args := pgx.NamedArgs{"id": id}
	row := dbpool.QueryRow(context.Background(), query, args)
	user := User{}
	err := row.Scan(&user.ID, &user.Email)
	return user, err
}
