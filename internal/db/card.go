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
	// We override property modify them. string -> []string
	Achievements []string
	DungeonOne   []string
	DungeonTwo   []string
	DungeonThree []string
}

func insertCards(db *pgxpool.Pool, cards []parser.Card) error {
	rows := [][]any{}
	for _, card := range cards {
		rows = append(rows, []any{
			card.ID,
			card.Idx,
			card.Level,
			card.Info,
			card.TaskMerge,
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
		"idx",
		"level",
		"info",
		"task_merge",
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
		idx,
		level,
		info,
		task_merge,
		task_one,
		task_two,
		achievements,
		dungeon_one,
		dungeon_two,
		dungeon_three,
		spell
		FROM cards
		WHERE id >= @lo AND id < @hi
		ORDER BY idx;`
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
	var dungeonOneStr string
	var dungeonTwoStr string
	var dungeonThreeStr string
	for rows.Next() {
		card := Card{}
		err := rows.Scan(
			&card.ID,
			&card.Idx,
			&card.Level,
			&card.Info,
			&card.TaskMerge,
			&card.TaskOne,
			&card.TaskTwo,
			&achievementsStr,
			&dungeonOneStr,
			&dungeonTwoStr,
			&dungeonThreeStr,
			&card.Spell,
		)

		card.Achievements = listFromString(achievementsStr)
		card.DungeonOne = listFromString(dungeonOneStr)
		card.DungeonTwo = listFromString(dungeonTwoStr)
		card.DungeonThree = listFromString(dungeonThreeStr)
		if err != nil {
			return nil, err
		}
		cards = append(cards, card)
	}
	return cards, nil
}

func listFromString(value string) []string {
	if len(strings.TrimSpace(value)) == 0 {
		return []string{}
	} else {
		return strings.Split(strings.TrimSpace(value), "\n")
	}
}
