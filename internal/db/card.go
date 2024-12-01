package db

import (
	"context"
	"optiguide/internal/parser"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func InsertCards(dbPool *pgxpool.Pool, cards []parser.Card) error {
	rows := [][]any{}
	for _, card := range cards {
		rows = append(rows, []any{
			card.ID,
			card.Idx,
			card.Level,
			card.Info,
			card.TaskTitleOne,
			card.TaskTitleTwo,
			card.TaskContentOne,
			card.TaskContentTwo,
			// TODO : change this system, this isn't the best but its ok
			strings.Join(card.Achievements, "\n"),
			strings.Join(card.DungeonOne, "\n"),
			strings.Join(card.DungeonTwo, "\n"),
			strings.Join(card.DungeonThree, "\n"),
			card.Spell,
		})
	}
	_, err := dbPool.CopyFrom(context.Background(), pgx.Identifier{"cards"}, []string{
		"id",
		"idx",
		"level",
		"info",
		"task_title_one",
		"task_title_two",
		"task_content_one",
		"task_content_two",
		"achievements",
		"dungeon_one",
		"dungeon_two",
		"dungeon_three",
		"spell",
	}, pgx.CopyFromRows(rows))
	return err
}

func GetCards(db *pgxpool.Pool, page int) ([]parser.Card, error) {
	const PAGESIZE = 10
	query := `SELECT
		idx,
		level,
		info,
		task_title_one,
		task_title_two,
		task_content_one,
		task_content_two,
		achievements,
		dungeon_one,
		dungeon_two,
		dungeon_three,
		spell
		FROM cards
		WHERE idx >= @lo AND idx < @hi
		ORDER BY idx;`
	args := pgx.NamedArgs{
		"lo": page * PAGESIZE,
		"hi": (page + 1) * PAGESIZE,
	}
	cards := make([]parser.Card, 0, PAGESIZE)
	rows, err := db.Query(context.Background(), query, args)
	if err != nil {
		return nil, err
	}
	var achievementsStr string
	var dungeonOneStr string
	var dungeonTwoStr string
	var dungeonThreeStr string
	for rows.Next() {
		card := parser.Card{}
		err := rows.Scan(
			&card.Idx,
			&card.Level,
			&card.Info,
			&card.TaskTitleOne,
			&card.TaskTitleTwo,
			&card.TaskContentOne,
			&card.TaskContentTwo,
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
