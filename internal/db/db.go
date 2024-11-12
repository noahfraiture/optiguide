package db

import (
	"context"
	"fmt"

	"optiguide/internal/parser"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
)

var dbPool *pgxpool.Pool

func GetPool() (*pgxpool.Pool, error) {
	var err error
	if dbPool == nil {
		err = Init()
	}
	return dbPool, err
}

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
	dbPool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = tableCard(dbPool)
	if err != nil {
		return err
	}
	err = deleteCards(dbPool)
	if err != nil {
		return err
	}
	cards, err := parser.Parse("guide.xlsx")
	if err != nil {
		return err
	}
	err = insertCards(dbPool, cards)
	if err != nil {
		return err
	}

	err = tableUser(dbPool)
	if err != nil {
		return err
	}
	err = tableUserClass(dbPool)
	if err != nil {
		return err
	}
	err = tableProgress(dbPool)
	if err != nil {
		return err
	}

	return nil
}

func Close() {
	dbPool.Close()
}
