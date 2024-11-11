package db

import (
	"context"
	"optiguide/internal/parser"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Card struct {
	parser.Card
	Achievements []string // NOTE : override achievements in slice
}

func tableCard(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS cards(
			id INTEGER PRIMARY KEY,
			level INTEGER NOT NULL,
			info TEXT,
			task_one TEXT,
			task_two TEXT,
			achievements TEXT,
			dungeon_one TEXT,
			dungeon_two TEXT,
			dungeon_three TEXT,
			spell TEXT
		);`,
	)
	return err
}

func deleteCards(db *pgxpool.Pool) error {
	query := "DELETE FROM cards;"
	_, err := db.Exec(context.Background(), query)
	return err
}

func insertCards(db *pgxpool.Pool, cards []parser.Card) error {
	rows := [][]any{}
	for _, card := range cards {
		rows = append(rows, []any{
			card.ID,
			card.Level,
			card.Info,
			card.TaskOne,
			card.TaskTwo,
			card.Achievements,
			card.DungeonOne,
			card.DungeonTwo,
			card.DungeonThree,
			card.Spell,
		})
	}
	_, err := db.CopyFrom(context.Background(), pgx.Identifier{"cards"}, []string{
		"id",
		"level",
		"info",
		"task_one",
		"task_two",
		"achievements",
		"dungeon_one",
		"dungeon_two",
		"dungeon_three",
		"spell",
	}, pgx.CopyFromRows(rows))
	return err
}

func GetCards(db *pgxpool.Pool, page int) ([]Card, error) {
	const PAGESIZE = 10
	query := `SELECT
		id,
		level,
		info,
		task_one,
		task_two,
		achievements,
		dungeon_one,
		dungeon_two,
		dungeon_three,
		spell
		FROM cards
		WHERE id >= @lo AND id < @hi
		ORDER BY id;`
	args := pgx.NamedArgs{
		"lo": page * PAGESIZE,
		"hi": (page + 1) * PAGESIZE,
	}
	cards := make([]Card, 0, PAGESIZE)
	rows, err := db.Query(context.Background(), query, args)
	if err != nil {
		return nil, err
	}
	var achievementsStr string
	for rows.Next() {
		card := Card{}
		err := rows.Scan(
			&card.ID,
			&card.Level,
			&card.Info,
			&card.TaskOne,
			&card.TaskTwo,
			&achievementsStr,
			&card.DungeonOne,
			&card.DungeonTwo,
			&card.DungeonThree,
			&card.Spell,
		)
		card.Achievements = strings.Split(achievementsStr, "\n")
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}
