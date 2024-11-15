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

// NOTE : insert the user and set its email and team_size
func SetUser(db *pgxpool.Pool, user *User) error {
	query := `WITH inputs(id, email, team_size) AS (
		VALUES (@id::text, @email::text, 1::integer)
	), ins_user AS (
		INSERT INTO users(id, email, team_size)
		SELECT * FROM inputs
		ON CONFLICT (id) DO NOTHING
		RETURNING id, email, team_size
	), ins_team AS (
		INSERT INTO user_box(user_id, box_index, class)
		SELECT id, 0, 0 FROM ins_user
		ON CONFLICT (user_id, box_index) DO NOTHING
	), existing AS (
		SELECT id, email, team_size
		FROM users
		WHERE id = @id
	)
	SELECT id, email, team_size FROM ins_user UNION SELECT id, email, team_size FROM existing;`

	args := pgx.NamedArgs{"id": user.ID, "email": user.Email}
	row := db.QueryRow(context.Background(), query, args)
	err := row.Scan(&user.ID, &user.Email, &user.TeamSize)
	return err
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
	TOUS Class = iota
	ECAFLIP
	ENIRIPSA
	IOP
	CRA
	FECA
	SACRIEUR
	SADIDA
	OSAMODAS
	ENUTROF
	SRAM
	XELOR
	PANDAWA
	ROUBLARD
	ZOBAL
	STEAMER
	ELIOTROPE
	HUPPERMAGE
	OUGINAK
	FORGELANCE

	NB_CLASS
)

var ClassToName = map[Class]string{
	TOUS:       "TOUS",
	ECAFLIP:    "ECAFLIP",
	ENIRIPSA:   "ENIRIPSA",
	IOP:        "IOP",
	CRA:        "CRA",
	FECA:       "FECA",
	SACRIEUR:   "SACRIEUR",
	SADIDA:     "SADIDA",
	OSAMODAS:   "OSAMODAS",
	ENUTROF:    "ENUTROF",
	SRAM:       "SRAM",
	XELOR:      "XELOR",
	PANDAWA:    "PANDAWA",
	ROUBLARD:   "ROUBLARD",
	ZOBAL:      "ZOBAL",
	STEAMER:    "STEAMER",
	ELIOTROPE:  "ELIOTROPE",
	HUPPERMAGE: "HUPPERMAGE",
	OUGINAK:    "OUGINAK",
	FORGELANCE: "FORGELANCE",
}

type TeamBox struct {
	userID   string
	BoxIndex int
	Class    Class
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

	query :=
		`INSERT INTO user_box(user_id, box_index, class)
		VALUES (@user_id, @box_index, @class)
		ON CONFLICT (user_id, box_index)
		DO UPDATE SET class = @class`
	_, err := db.Exec(context.Background(), query, pgx.NamedArgs{
		"user_id":   userID,
		"box_index": boxIndex,
		"class":     class,
	})
	return err
}

func InsertClass(db *pgxpool.Pool, userID string, boxIndex int, class Class) (TeamBox, error) {
	query :=
		`WITH inputs(user_id, box_index, class) AS (
	        VALUES (@user_id::text, @box_index::integer, @class::integer)
	    ), ins AS (
	    	INSERT INTO user_box (user_id, box_index, class)
	    	SELECT * FROM inputs
	    	ON CONFLICT (user_id, box_index) DO NOTHING
	    	RETURNING class, box_index
		), existing AS (
			SELECT class, box_index
			FROM user_box
			WHERE user_id = @user_id AND box_index = @box_index
		)
		SELECT class, box_index FROM ins UNION SELECT class, box_index FROM existing;`

	teamBox := TeamBox{userID: userID}

	row := db.QueryRow(context.Background(), query, pgx.NamedArgs{
		"user_id":   userID,
		"box_index": boxIndex,
		"class":     class,
	})
	err := row.Scan(&teamBox.Class, &teamBox.BoxIndex)
	return teamBox, err
}

func GetClasses(db *pgxpool.Pool, userID string) ([]TeamBox, error) {
	query :=
		`SELECT user_id, box_index, class
		FROM user_box
		JOIN users ON id = user_id
		WHERE user_id = @user_id AND box_index < team_size
		ORDER BY box_index;`
	args := pgx.NamedArgs{"user_id": userID}
	rows, err := db.Query(context.Background(), query, args)
	if err != nil {
		return nil, err
	}

	boxes := make([]TeamBox, 0)
	for rows.Next() {
		box := TeamBox{}
		err := rows.Scan(&box.userID, &box.BoxIndex, &box.Class)
		if err != nil {
			return nil, err
		}
		boxes = append(boxes, box)
	}
	return boxes, nil
}

// Not a table in the database but used for the html render
// Box state by index of the box
type BoxesState map[int]bool

// Merge between Progress and TeamBox
func GetRenderBoxByCards(db *pgxpool.Pool, userID string) (map[int]BoxesState, error) {
	// We can't make a LEFT JOIN to have `done` as false for a default value
	// Because we can't know `card_id`
	query :=
		`SELECT card_id, box_index, done
		FROM progress
		WHERE user_id = @user_id;`
	args := pgx.NamedArgs{
		"user_id": userID,
	}
	rows, err := db.Query(context.Background(), query, args)
	if err != nil {
		return nil, err
	}

	cardBoxes := make(map[int]BoxesState, 0)
	for rows.Next() {
		var done bool
		var box_index int
		var card_id int
		err = rows.Scan(&card_id, &box_index, &done)
		if err != nil {
			return nil, err
		}
		if _, ok := cardBoxes[card_id]; !ok {
			cardBoxes[card_id] = make(BoxesState)
		}
		cardBoxes[card_id][box_index] = done
	}
	return cardBoxes, nil
}
