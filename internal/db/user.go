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

func GetUser(db *pgxpool.Pool, userID string) (User, error) {
	user := User{ID: userID}
	err := SetUser(dbPool, &user)
	return user, err
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
		INSERT INTO user_characters(user_id, box_index, class, name)
		SELECT id, 0, @class, @name FROM ins_user
		ON CONFLICT (user_id, box_index) DO NOTHING
	), existing AS (
		SELECT id, email, team_size
		FROM users
		WHERE id = @id
	)
	SELECT id, email, team_size FROM ins_user UNION SELECT id, email, team_size FROM existing;`

	args := pgx.NamedArgs{"id": user.ID, "email": user.Email, "class": NONE, "name": "Perso 1"}
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
	NONE Class = iota
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
	NONE:       "TOUS",
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

type Character struct {
	userID   string
	BoxIndex int
	Class    Class
	Name     string
}

func tableUserTeam(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS user_characters(
			user_id TEXT REFERENCES users(id) ON DELETE CASCADE,
			box_index INTEGER NOT NULL,
			class INTEGER NOT NULL,
			name TEXT NOT NULL,
			PRIMARY KEY(user_id, box_index)
		);`,
	)
	return err
}

// Update the character class, and if first time update, set the name to the name of the class
func UpdateCharacterClass(db *pgxpool.Pool, userID string, boxIndex int, class Class) error {
	query :=
		`INSERT INTO user_characters(user_id, box_index, class, name)
		VALUES (@user_id, @box_index, @class, @name)
		ON CONFLICT (user_id, box_index)
		DO UPDATE SET class = @class`
	_, err := db.Exec(context.Background(), query, pgx.NamedArgs{
		"user_id":   userID,
		"box_index": boxIndex,
		"class":     class,
		"name":      ClassToName[class],
	})
	return err
}

// Update the character class, and if first time update, set the class to NONE which is the current displayed
func UpdateCharacterName(db *pgxpool.Pool, userID string, boxIndex int, name string) error {
	query :=
		`INSERT INTO user_characters(user_id, box_index, class, name)
		VALUES (@user_id, @box_index, @class, @name)
		ON CONFLICT (user_id, box_index)
		DO UPDATE SET name = @name`
	_, err := db.Exec(context.Background(), query, pgx.NamedArgs{
		"user_id":   userID,
		"box_index": boxIndex,
		"class":     NONE,
		"name":      name,
	})
	return err
}

func GetTeam(db *pgxpool.Pool, userID string) ([]Character, error) {
	query :=
		`SELECT user_id, box_index, class, name
		FROM user_characters
		JOIN users ON id = user_id
		WHERE user_id = @user_id AND box_index < team_size
		ORDER BY box_index;`
	args := pgx.NamedArgs{"user_id": userID}
	rows, err := db.Query(context.Background(), query, args)
	if err != nil {
		return nil, err
	}

	boxes := make([]Character, 0)
	for rows.Next() {
		box := Character{}
		err := rows.Scan(&box.userID, &box.BoxIndex, &box.Class, &box.Name)
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

// Merge between Progress and TeamBox to get the state (done/!done) for
// each box of each card BoxesState is map[boxIndex]isDone := map[int]bool
// We only need the done status of a box because we only need the class &
// name once, when we get the whole team
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
