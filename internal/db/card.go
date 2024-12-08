package db

import (
	"context"
	"fmt"
	"optiguide/internal/parser"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Achievement struct {
	Name string
	Link string
	Done bool
}

type Card struct {
	parser.Card
	// We create the same achievements but with the field Done
	Achievements []Achievement
	// We add the `boxes` which are the progress of the team on the card
	Boxes []bool
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
	ctx := context.Background()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	for _, card := range cards {
		err := insertCard(tx, card)
		if err != nil {
			return err
		}

		err = insertAchievements(tx, card)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func UpdateCards(dbPool *pgxpool.Pool, cards []parser.Card) error {
	ctx := context.Background()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not start transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	for _, card := range cards {
		id, err := updateCard(tx, card)
		if err != nil {
			return err
		}

		card.ID = id
		err = updateAchievements(tx, card)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}
func insertCard(tx pgx.Tx, card parser.Card) error {
	query := `INSERT INTO cards (id, idx, level, info, task_title_one, task_title_two, task_content_one, task_content_two, dungeon_one, dungeon_two, dungeon_three, spell)
        VALUES (@id, @idx, @level, @info, @task_title_one, @task_title_two, @task_content_one, @task_content_two, @dungeon_one, @dungeon_two, @dungeon_three, @spell)`
	args := pgx.NamedArgs{
		"id":               card.ID,
		"idx":              card.Idx,
		"level":            card.Level,
		"info":             card.Info,
		"task_title_one":   card.TaskTitleOne,
		"task_title_two":   card.TaskTitleTwo,
		"task_content_one": card.TaskContentOne,
		"task_content_two": card.TaskContentTwo,
		"dungeon_one":      strings.Join(card.DungeonOne, "\n"),
		"dungeon_two":      strings.Join(card.DungeonTwo, "\n"),
		"dungeon_three":    strings.Join(card.DungeonThree, "\n"),
		"spell":            card.Spell,
	}
	_, err := tx.Exec(context.Background(), query, args)
	if err != nil {
		return fmt.Errorf("failed to insert card: %v", err)
	}
	return nil
}

func updateCard(tx pgx.Tx, card parser.Card) (uuid.UUID, error) {
	query := `
		WITH ins AS (
	        INSERT INTO cards (id, idx, level, info, task_title_one, task_title_two, task_content_one, task_content_two, dungeon_one, dungeon_two, dungeon_three, spell)
	        VALUES (@id, @idx, @level, @info, @task_title_one, @task_title_two, @task_content_one, @task_content_two, @dungeon_one, @dungeon_two, @dungeon_three, @spell)
	        ON CONFLICT (idx) DO NOTHING
	        RETURNING id
		), existing AS (
			SELECT id FROM cards WHERE idx = $2
		)
		SELECT id FROM ins UNION SELECT id FROM existing;`
	args := pgx.NamedArgs{
		"id":               card.ID,
		"idx":              card.Idx,
		"level":            card.Level,
		"info":             card.Info,
		"task_title_one":   card.TaskTitleOne,
		"task_title_two":   card.TaskTitleTwo,
		"task_content_one": card.TaskContentOne,
		"task_content_two": card.TaskContentTwo,
		"dungeon_one":      strings.Join(card.DungeonOne, "\n"),
		"dungeon_two":      strings.Join(card.DungeonTwo, "\n"),
		"dungeon_three":    strings.Join(card.DungeonThree, "\n"),
		"spell":            card.Spell,
	}
	row := tx.QueryRow(context.Background(), query, args)
	var id uuid.UUID
	err := row.Scan(&id)
	if err != nil {
		return id, fmt.Errorf("failed to insert card: %v", err)
	}
	return id, nil
}

func insertAchievements(tx pgx.Tx, card parser.Card) error {
	for _, achievement := range card.Achievements {
		achievementID := uuid.New()
		_, err := tx.Exec(context.Background(), `
            INSERT INTO achievements (id, name, link, card_id)
            VALUES ($1, $2, $3, $4)`,
			achievementID,
			achievement.Value,
			achievement.Link,
			card.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert achievement: %v", err)
		}
	}
	return nil
}

func updateAchievements(tx pgx.Tx, card parser.Card) error {
	for _, achievement := range card.Achievements {
		achievementID := uuid.New()
		_, err := tx.Exec(context.Background(), `
            INSERT INTO achievements (id, name, link, card_id)
            VALUES ($1, $2, $3, $4)
            ON CONFLICT (name, card_id) DO NOTHING`,
			achievementID,
			achievement.Value,
			achievement.Link,
			card.ID,
		)
		if err != nil {
			return fmt.Errorf("failed to insert achievement: %v", err)
		}
	}
	return nil
}

func GetCards(dbPool *pgxpool.Pool, user User) ([]*Card, error) {
	ctx := context.Background()

	tx, err := dbPool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `SELECT
        cards.idx,
        cards.level,
        cards.info,
        cards.task_title_one,
        cards.task_title_two,
        cards.task_content_one,
        cards.task_content_two,
        cards.dungeon_one,
        cards.dungeon_two,
        cards.dungeon_three,
        cards.spell,
        progress.box_index,
        progress.done,
        achievements.name,
        achievements.link,
        achievements_users.done,
		users.team_size
    FROM cards
    JOIN users
		ON users.id = @user_id
    LEFT JOIN progress
		ON cards.id = progress.card_id
		AND progress.user_id = @user_id
		AND progress.box_index < users.team_size
    LEFT JOIN achievements
		ON cards.id = achievements.card_id
    LEFT JOIN achievements_users
		ON achievements.id = achievements_users.achievement_id
		AND @user_id = achievements_users.user_id
    ORDER BY cards.idx, achievements.name;`

	args := pgx.NamedArgs{
		"user_id": user.ID,
	}

	rows, err := tx.Query(ctx, query, args)
	if err != nil {
		return nil, err
	}

	// A map to associate each card's index to its instance to manage boxes state
	cardsMap := make(map[int]*Card)
	// Since we don't know if we query a page or all, we use 0
	// TODO : may add the number of all cards if known
	cards := make([]*Card, 0)

	// Process rows
	for rows.Next() {
		var (
			idx             int
			dungeonOneStr   string
			dungeonTwoStr   string
			dungeonThreeStr string
			boxIndex        *int
			boxDone         *bool
			achievementName *string
			achievementLink *string
			achievementDone *bool
			teamSize        int
		)

		// Create an inner newCard for the current row
		newCard := Card{}

		// We scan the card. If it's already known, only the idx is relevant
		// We might already know it because the way the query is done, will
		// returns the same card multiple times for each box and achievements possible
		// NOTE : we could optimize that
		err := rows.Scan(
			&idx,
			&newCard.Level,
			&newCard.Info,
			&newCard.TaskTitleOne,
			&newCard.TaskTitleTwo,
			&newCard.TaskContentOne,
			&newCard.TaskContentTwo,
			&dungeonOneStr,
			&dungeonTwoStr,
			&dungeonThreeStr,
			&newCard.Spell,
			&boxIndex,
			&boxDone,
			&achievementName,
			&achievementLink,
			&achievementDone,
			&teamSize,
		)
		if err != nil {
			return nil, err
		}

		existingCard, exists := cardsMap[idx]
		if !exists {
			// We set these once we don't have row for the same card anymore
			newCard.Idx = idx
			newCard.DungeonOne = listFromString(dungeonOneStr)
			newCard.DungeonTwo = listFromString(dungeonTwoStr)
			newCard.DungeonThree = listFromString(dungeonThreeStr)
			newCard.Boxes = make([]bool, teamSize)
			cardsMap[idx] = &newCard
			cards = append(cards, &newCard)
			existingCard = &newCard
		}

		if boxIndex != nil && boxDone != nil {
			existingCard.Boxes[*boxIndex] = *boxDone
		}

		if achievementName != nil {
			var link string
			if achievementLink != nil {
				link = *achievementLink
			}
			var done bool
			if achievementDone != nil {
				done = *achievementDone
			}
			contains := false
			for _, ach := range existingCard.Achievements {
				if ach.Name == *achievementName {
					contains = true
					break
				}
			}
			if !contains {
				existingCard.Achievements = append(existingCard.Achievements, Achievement{
					Name: *achievementName,
					Link: link,
					Done: done,
				})
			}
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
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
