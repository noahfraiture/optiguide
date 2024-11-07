package db

import (
	"context"
	"fmt"
	"optiguide/internal/parser"
	"os"
	"strings"

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
	dbpool, err = pgxpool.New(context.Background(), connStr)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = tableUser()
	if err != nil {
		return err
	}
	err = tableCard()
	if err != nil {
		return err
	}
	err = tableProgress()
	if err != nil {
		return err
	}
	cards, err := parser.Parse("guide.xlsx")
	if err != nil {
		return err
	}
	isFull, err := isCardsFull()
	if err != nil {
		return err
	}
	if !isFull {
		err = insertCards(cards)
		if err != nil {
			return err
		}
	}

	return nil
}

func Close() {
	dbpool.Close()
}

type User struct {
	ID       string
	Email    string
	TeamSize int
}

// We define a user as an id (from auth provider), his email and the number of case to display
func tableUser() error {
	_, err := dbpool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS users(
			id TEXT PRIMARY KEY,
			team_size INTEGER NOT NULL,
			email TEXT
		);`,
	)
	return err
}

func InsertUser(user User) error {
	query := `INSERT INTO users (id, email, team_size) VALUES (@id, @email, 1) ON CONFLICT (id) DO NOTHING;`
	args := pgx.NamedArgs{
		"id":    user.ID,
		"email": user.Email,
	}
	_, err := dbpool.Exec(context.Background(), query, args)
	return err
}

func QueryUser(id string) (User, error) {
	query := `SELECT id, email, team_size FROM users WHERE id = @id;`
	args := pgx.NamedArgs{"id": id}
	row := dbpool.QueryRow(context.Background(), query, args)
	user := User{}
	err := row.Scan(&user.ID, &user.Email, &user.TeamSize)
	return user, err
}

// Done is an integer that show with his binary value the state of completeness of the team :
// 0b1010 means he completed for his second and forth character

type Progress struct {
	UserID string
	CardID int
	Done   int
}

func tableProgress() error {
	_, err := dbpool.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS progress(
			user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
			card_id INTEGER REFERENCES cards(id) ON DELETE CASCADE,
			done INTEGER NOT NULL,
			PRIMARY KEY(user_id, card_id)
		);`,
	)
	return err
}
func (u *User) SetProgress(cardID, box int) error {
	query := `INSERT INTO progress (user_id, card_id, done)
	VALUES (@user_id, @card_id, 1<<@box)
	ON CONFLICT (user_id, card_id)
	DO UPDATE SET done = progress.done # (1<<@box);`
	args := pgx.NamedArgs{
		"user_id": u.ID,
		"card_id": cardID,
		"box":     box,
	}
	_, err := dbpool.Exec(context.Background(), query, args)
	return err
}

func (u *User) GetDoneValue(cardID int) (int, error) {
	query := `SELECT COALESCE ((SELECT done FROM progress WHERE user_id = @user_id AND card_id = @card_id), 0);`
	args := pgx.NamedArgs{
		"user_id": u.ID,
		"card_id": cardID,
	}
	row := dbpool.QueryRow(context.Background(), query, args)
	var done int
	err := row.Scan(&done)
	return done, err
}

func (u *User) PlusTeamSize(value int) error {
	_, err := dbpool.Exec(
		context.Background(),
		`UPDATE users SET team_size = users.team_size + @value WHERE id = @id;`,
		pgx.NamedArgs{
			"id":    u.ID,
			"value": value,
		},
	)
	return err
}

func tableCard() error {
	_, err := dbpool.Exec(context.Background(),
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

func isCardsFull() (bool, error) {
	row := dbpool.QueryRow(
		context.Background(),
		`SELECT CASE WHEN EXISTS(SELECT true FROM cards) THEN true ELSE false END;`,
	)
	var isFull bool
	err := row.Scan(&isFull)
	return isFull, err
}

func insertCards(cards []parser.Card) error {
	rows := [][]any{}
	for _, card := range cards {
		rows = append(rows, []any{
			card.ID,
			card.Level,
			card.Info,
			card.TaskOne,
			card.TaskTwo,
			strings.Join(card.Achievements, "\n"),
			card.DungeonOne,
			card.DungeonTwo,
			card.DungeonThree,
			card.Spell,
		})
	}
	_, err := dbpool.CopyFrom(context.Background(), pgx.Identifier{"cards"}, []string{
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

type CardUser struct {
	Card parser.Card
	Done int
}

func (u *User) GetPage(page int) ([]CardUser, error) {
	const pageSize = 10
	rows, err := dbpool.Query(context.Background(),
		`SELECT
			id,
			level,
			info,
			task_one,
			task_two,
			achievements,
			dungeon_one,
			dungeon_two,
			dungeon_three,
			spell,
			COALESCE(progress.done, 0) AS done
		FROM cards
		LEFT JOIN progress ON card_id = id AND user_id = @user_id
		WHERE id >= @min AND id < @max
		ORDER BY id;`,
		pgx.NamedArgs{"min": page * pageSize, "max": (page + 1) * pageSize, "user_id": u.ID},
	)
	if err != nil {
		return nil, err
	}
	cards := make([]CardUser, 0, pageSize)
	for rows.Next() {
		card := parser.Card{}
		var achievementsStr string
		var done int
		err = rows.Scan(&card.ID,
			&card.Level,
			&card.Info,
			&card.TaskOne,
			&card.TaskTwo,
			&achievementsStr,
			&card.DungeonOne,
			&card.DungeonTwo,
			&card.DungeonThree,
			&card.Spell,
			&done,
		)
		card.Achievements = strings.Split(achievementsStr, "\n")
		if err != nil {
			return nil, err
		}
		cards = append(cards, CardUser{Card: card, Done: done})
	}
	return cards, nil

}