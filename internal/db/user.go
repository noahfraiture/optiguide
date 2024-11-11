package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type User struct {
	ID       string
	Email    string
	TeamSize int
}

// We define a user as an id (from auth provider), his email and the number of case to display
func tableUser(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS users(
			id TEXT PRIMARY KEY,
			team_size INTEGER NOT NULL,
			email TEXT
		);`,
	)
	return err
}

func InsertUser(db *pgxpool.Pool, user User) error {
	query :=
		`INSERT INTO users(id, email, team_size)
		VALUES (@id, @email, 1)
		ON CONFLICT (id) DO NOTHING;`
	args := pgx.NamedArgs{
		"id":    user.ID,
		"email": user.Email,
	}
	_, err := db.Exec(context.Background(), query, args)
	if err != nil {
		return err
	}
	err = InsertClass(db, user.ID, 0, NONE)
	return err
}

func QueryUser(db *pgxpool.Pool, id string) (User, error) {
	query := `SELECT id, email, team_size FROM users WHERE id = @id;`
	args := pgx.NamedArgs{"id": id}
	row := db.QueryRow(context.Background(), query, args)
	user := User{}
	err := row.Scan(&user.ID, &user.Email, &user.TeamSize)
	return user, err
}

func PlusTeamSize(db *pgxpool.Pool, userID string, value int) error {
	_, err := db.Exec(
		context.Background(),
		`UPDATE users SET team_size = users.team_size + @value WHERE id = @id;`,
		pgx.NamedArgs{
			"id":    userID,
			"value": value,
		},
	)
	return err
}

// The order is the order of release of the Class that can be found
// https://www.dofus.com/fr/mmorpg/encyclopedie/classes
type Class int

const (
	NONE Class = iota
	ECAFLIP
	ENIRIPSA
	IOP
	CRA
	FECA
	SACRIEUR
	SADIDAS
	OSAMODAS
	ENUTROF
	SRAM
	XELOR
	PANDA
	ROUBLARD
	ZOBAL
	STEAMER
	ELIOTROPE
	HUPPERMAGE
	OUGINAK
	FORGELANCE
)

type UserBox struct {
	userID   string
	boxIndex int
	class    Class
}

func tableUserClass(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS user_box(
			user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
			box_index INTEGER NOT NULL,
			class INTEGER NOT NULL,
			PRIMARY KEY(user_id, box_index)
		);`,
	)
	return err
}

func UpdateClass(db *pgxpool.Pool, userID string, boxIndex int, class Class) error {

	query := `UPDATE user_box SET class = @class
		WHERE user_id = @user_id AND box_index = @box_index;`
	_, err := db.Exec(context.Background(), query, pgx.NamedArgs{
		"user_id":   userID,
		"box_index": boxIndex,
		"class":     class,
	})
	return err
}

func InsertClass(db *pgxpool.Pool, userID string, boxIndex int, class Class) error {
	query :=
		`INSERT INTO user_box(user_id, box_index, class)
		VALUES (@user_id, @box_index, @class)
		ON CONFLICT DO NOTHING;`
	_, err := db.Exec(context.Background(), query, pgx.NamedArgs{
		"user_id":   userID,
		"box_index": boxIndex,
		"class":     class,
	})
	return err
}

func GetBoxes(db *pgxpool.Pool, userID string) ([]UserBox, error) {
	query :=
		`SELECT user_id, box_index, class
		FROM user_box
		WHERE user_id = @user_id
		ORDER BY box_index;`
	args := pgx.NamedArgs{"user_id": userID}
	rows, err := db.Query(context.Background(), query, args)
	if err != nil {
		return nil, err
	}

	boxes := make([]UserBox, 0)
	for rows.Next() {
		box := UserBox{}
		err := rows.Scan(&box.userID, &box.boxIndex, &box.class)
		if err != nil {
			return nil, err
		}
		boxes = append(boxes, box)
	}
	return boxes, nil
}

// Not a table in the database but used for the html render
type Box struct {
	Done  bool
	Class Class
}
type Boxes map[int]Box // Box state by index

// Merge between Progress and UserBox
func GetRenderBoxByCards(db *pgxpool.Pool, userID string) (map[int]Boxes, error) {
	query := `SELECT
		progress.card_id, user_box.box_index, progress.done, user_box.class
		FROM user_box
		JOIN progress on progress.box_index = user_box.box_index
		WHERE user_box.user_id = @user_id;`
	args := pgx.NamedArgs{
		"user_id": userID,
	}
	rows, err := db.Query(context.Background(), query, args)
	if err != nil {
		return nil, err
	}

	cardBoxes := make(map[int]Boxes, 0)
	for rows.Next() {
		var box Box
		var box_index int
		var card_id int
		err = rows.Scan(&card_id, &box_index, &box.Done, &box.Class)
		if err != nil {
			return nil, err
		}
		if _, ok := cardBoxes[card_id]; !ok {
			cardBoxes[card_id] = make(Boxes)
		}
		cardBoxes[card_id][box_index] = box
	}
	return cardBoxes, nil
}
