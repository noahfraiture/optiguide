package db

import (
	"context"
	"fmt"
	"optiguide/internal/parser"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Achievement struct {
	Name string
	Done bool
}

type Card struct {
	parser.Card
	Achievements []Achievement
	Boxes        []bool
}

func IsEmpty(dbPool *pgxpool.Pool) (bool, error) {
	ctx := context.Background()
	query := `SELECT EXISTS(SELECT 1 FROM cards LIMIT 1);`

	var exists bool
	err := dbPool.QueryRow(ctx, query).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("error checking if cards table is empty: %w", err)
	}

	// `exists` will be true if there is at least one row, otherwise it will be false
	return !exists, nil
}

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

func GetCards(dbPool *pgxpool.Pool, user User, page int) ([]Card, error) {
	const PAGESIZE = 10
	ctx := context.Background()

	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err != nil {
			tx.Rollback(context.Background())
		}
	}()

	query := `SELECT
        cards.idx,
        cards.level,
        cards.info,
        cards.task_title_one,
        cards.task_title_two,
        cards.task_content_one,
        cards.task_content_two,
        cards.achievements,
        cards.dungeon_one,
        cards.dungeon_two,
        cards.dungeon_three,
        cards.spell,
        progress.box_index,
        progress.done,
        achievements.name,
        achievements.done,
		users.team_size
    FROM cards
    JOIN users ON users.id = @user_id
    LEFT JOIN progress ON cards.id = progress.card_id AND progress.user_id = @user_id
    LEFT JOIN achievements ON cards.id = achievements.card_id AND achievements.user_id = @user_id
    WHERE cards.idx >= @lo AND cards.idx < @hi
    ORDER BY cards.idx;`

	args := pgx.NamedArgs{
		"user_id": user.ID,
		"lo":      page * PAGESIZE,
		"hi":      (page + 1) * PAGESIZE,
	}

	rows, err := tx.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}

	// A map to associate each card's index to its instance, for managing boxes state
	cardsMap := make(map[int]*Card)
	cards := make([]Card, 0, PAGESIZE)

	// Process rows
	for rows.Next() {
		var (
			idx             int
			achievementStr  string
			dungeonOneStr   string
			dungeonTwoStr   string
			dungeonThreeStr string
			boxIndex        *int
			boxDone         *bool
			achievementName *string
			achievementDone *bool
			teamSize        int
		)

		// Create an inner card for the current row
		card := Card{}

		// We scan the card. If it's already known, only the idx is relevant
		// We might already know it because the way the query is done, will
		// returns the same card multiple times for each box and achievements possible
		// NOTE : we could optimize that
		err := rows.Scan(
			&idx,
			&card.Level,
			&card.Info,
			&card.TaskTitleOne,
			&card.TaskTitleTwo,
			&card.TaskContentOne,
			&card.TaskContentTwo,
			&achievementStr,
			&dungeonOneStr,
			&dungeonTwoStr,
			&dungeonThreeStr,
			&card.Spell,
			&boxIndex,
			&boxDone,
			&achievementName,
			&achievementDone,
			&teamSize,
		)
		if err != nil {
			return nil, err
		}

		existingCard, exists := cardsMap[idx]
		if !exists {
			card.Idx = idx
			card.Achievements = parseAchievement(achievementStr)
			card.DungeonOne = listFromString(dungeonOneStr)
			card.DungeonTwo = listFromString(dungeonTwoStr)
			card.DungeonThree = listFromString(dungeonThreeStr)
			card.Boxes = make([]bool, teamSize)
			cardsMap[idx] = &card
			cards = append(cards, card)
			existingCard = &card
		} else {
			card = *existingCard
		}

		if boxIndex != nil && boxDone != nil {
			existingCard.Boxes[*boxIndex] = *boxDone
		}

		if achievementName != nil && achievementDone != nil {
			for i, achievement := range card.Achievements {
				if achievement.Name == *achievementName {
					achievement.Done = *achievementDone
					card.Achievements[i] = achievement
					break
				}
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return cards, nil
}

func parseAchievement(value string) []Achievement {
	names := strings.Split(value, "\n")
	achievements := make([]Achievement, 0, len(names))
	for _, name := range names {
		achievements = append(achievements, Achievement{Name: name, Done: false})
	}
	return achievements
}

func listFromString(value string) []string {
	if len(strings.TrimSpace(value)) == 0 {
		return []string{}
	} else {
		return strings.Split(strings.TrimSpace(value), "\n")
	}
}

type BoxesState map[int]bool

func GetCardBoxes(dbPool *pgxpool.Pool, user User) (map[int]BoxesState, error) {
	ctx := context.Background()

	query := `
        SELECT
            progress.card_id,
            progress.box_index,
            progress.done
        FROM progress
        WHERE progress.user_id = @user_id
        ORDER BY progress.card_id, progress.box_index;
    `
	args := pgx.NamedArgs{
		"user_id": user.ID,
	}

	rows, err := dbPool.Query(ctx, query, args)
	if err != nil {
		return nil, fmt.Errorf("error executing query to get card boxes: %w", err)
	}
	defer rows.Close()

	cardBoxes := make(map[int]BoxesState)

	for rows.Next() {
		var (
			cardId   int
			boxIndex int
			done     bool
		)

		if err := rows.Scan(&cardId, &boxIndex, &done); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		// If this card doesn't exist in the map yet, initialize its boxes slice
		if _, ok := cardBoxes[cardId]; !ok {
			cardBoxes[cardId] = make(BoxesState, user.TeamSize)
		}

		// Assign the box state
		if boxIndex < len(cardBoxes[cardId]) {
			cardBoxes[cardId][boxIndex] = done
		} else {
			// Handle the case where box_index might exceed the expected team size
			return nil, fmt.Errorf("box_index %d exceeds team size %d for card ID %d", boxIndex, user.TeamSize, cardId)
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during iteration over rows: %w", err)
	}

	return cardBoxes, nil
}
