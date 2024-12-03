package db

import (
	"context"
	"fmt"
	"strings"

	"optiguide/internal/parser"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

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
	password, err := getPassword()
	if err != nil {
		return err
	}
	connStrPgx := fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s",
		os.Getenv("POSTGRES_USER"),
		password,
		os.Getenv("POSTGRES_HOST"),
		"5432",
		os.Getenv("POSTGRES_DB"),
	)

	dbPool, err = pgxpool.New(context.Background(), connStrPgx)
	if err != nil {
		return err
	}

	if err := buildDatabase(password); err != nil && err != migrate.ErrNoChange {
		fmt.Println("error during building of database")
		return err
	}

	if err := updateCards(); err != nil {
		fmt.Println("error during rebuilding of cards")
		return err
	}

	return nil
}

func buildDatabase(password string) error {
	// Build the database from the migrations files
	connStrMigration := fmt.Sprintf(
		"postgresql://%s:%s@%s:%s/%s?sslmode=disable&search_path=public",
		os.Getenv("POSTGRES_USER"),
		password,
		os.Getenv("POSTGRES_HOST"),
		"5432",
		os.Getenv("POSTGRES_DB"),
	)
	m, err := migrate.New(
		"file://migrations",
		connStrMigration,
	)
	if err != nil {
		return err
	}
	return m.Up()
}

func rebuildCards() error {
	dbPool, err := GetPool()
	if err != nil {
		fmt.Println(err)
		return err
	}
	empty, err := IsEmpty(dbPool)
	if err != nil {
		fmt.Println(err)
		return err
	}
	if empty {
		cards, err := parser.Parse("guide.xlsx")
		if err != nil {
			return err
		}
		err = InsertCards(dbPool, cards)
		if err != nil {
			return err
		}
	}
	return nil
}

func updateCards() error {
	dbPool, err := GetPool()
	if err != nil {
		fmt.Println(err)
		return err
	}
	cards, err := parser.Parse("guide.xlsx")
	if err != nil {
		return err
	}
	err = UpdateCards(dbPool, cards)
	if err != nil {
		return err
	}
	return nil
}

func getPassword() (string, error) {
	password, ok := os.LookupEnv("POSTGRES_PASSWORD")
	if !ok {
		passwordFile, ok := os.LookupEnv("POSTGRES_PASSWORD_FILE")
		if !ok {
			return "", fmt.Errorf("No password set")
		}
		data, err := os.ReadFile(passwordFile)
		if err != nil {
			return "", err
		}
		password = strings.TrimSpace(string(data))
	}
	return password, nil
}
