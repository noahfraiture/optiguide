package db

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markbates/goth"
)

type User struct {
	ID         uuid.UUID
	ProviderID string
	Provider   string
	Username   string
	Email      string
	TeamSize   int
	// percentage of progress of the user.
	// This is used when create a new guild to display correct information without
	// make another request
	Progress float64
}

func GetUserFromProvider(dbPool *pgxpool.Pool, userAuth goth.User) (User, error) {
	user := User{
		ProviderID: userAuth.UserID,
		Provider:   userAuth.Provider,
		Email:      userAuth.Email,
		Username:   userAuth.NickName,
	}
	err := setUserFromProvider(dbPool, &user)
	return user, err
}

// Insert the user and set its email, team_size and username
func setUserFromProvider(dbPool *pgxpool.Pool, user *User) error {
	query := `WITH ins_user AS (
		INSERT INTO users(id, provider_id, provider, username, email, team_size, progress)
		VALUES (@id, @provider_id, @provider, @username, @email, 1, 0.0)
		ON CONFLICT (provider, provider_id) DO NOTHING
		RETURNING id, provider_id, provider, username, email, team_size, progress
	), ins_team AS (
		INSERT INTO user_characters(user_id, box_index, class, name)
		SELECT id, 0, @class, @name FROM ins_user
		ON CONFLICT (user_id, box_index) DO NOTHING
	), existing AS (
		SELECT id, provider_id, provider, username, email, team_size, progress
		FROM users
		WHERE provider_id = @provider_id AND provider = @provider
	)
	SELECT id, username, email, team_size, ROUND(progress::numeric, 2)
	FROM ins_user
	UNION
	SELECT id, username, email, team_size, ROUND(progress::numeric, 2)
	FROM existing;`

	args := pgx.NamedArgs{
		"id":          uuid.New(),
		"provider_id": user.ProviderID,
		"provider":    user.Provider,
		"email":       user.Email,
		"username":    user.Email,
		"class":       NONE,
		"name":        "Perso 1",
	}
	row := dbPool.QueryRow(context.Background(), query, args)
	err := row.Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.TeamSize,
		&user.Progress,
	)
	return err
}

func UpdateUsername(dbPool *pgxpool.Pool, user User, username string) error {
	_, err := dbPool.Exec(
		context.Background(),
		`UPDATE users
		SET username = @username
		WHERE id = @id;`,
		pgx.NamedArgs{
			"username": username,
			"id":       user.ID,
		},
	)
	return err
}

func PlusTeamSize(dbPool *pgxpool.Pool, user User, value int) error {
	_, err := dbPool.Exec(
		context.Background(),
		`UPDATE users
		SET
		  progress = users.progress * users.team_size / (users.team_size + @value),
		  team_size = users.team_size + @value
		WHERE
		  id = @id;`,
		pgx.NamedArgs{
			"id":    user.ID,
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

// Update the character class, and if first time update, set the name to the name of the class
func UpdateCharacterClass(dbPool *pgxpool.Pool, user User, boxIndex int, class Class) error {
	query :=
		`INSERT INTO user_characters(user_id, box_index, class, name)
		VALUES (@user_id, @box_index, @class, @name)
		ON CONFLICT (user_id, box_index)
		DO UPDATE SET class = @class`
	_, err := dbPool.Exec(context.Background(), query, pgx.NamedArgs{
		"user_id":   user.ID,
		"box_index": boxIndex,
		"class":     class,
		"name":      fmt.Sprintf("Perso %d", boxIndex+1),
	})
	return err
}

// Update the character class, and if first time update, set the class to NONE which is the current displayed
func UpdateCharacterName(dbPool *pgxpool.Pool, user User, boxIndex int, name string) error {
	query :=
		`INSERT INTO user_characters(user_id, box_index, class, name)
		VALUES (@user_id, @box_index, @class, @name)
		ON CONFLICT (user_id, box_index)
		DO UPDATE SET name = @name`
	_, err := dbPool.Exec(context.Background(), query, pgx.NamedArgs{
		"user_id":   user.ID,
		"box_index": boxIndex,
		"class":     NONE,
		"name":      name,
	})
	return err
}

func GetTeam(dbPoll *pgxpool.Pool, user User) ([]Character, error) {
	query :=
		`SELECT user_id, box_index, class, name
		FROM user_characters
		JOIN users ON id = user_id
		WHERE user_id = @user_id AND box_index < team_size
		ORDER BY box_index;`
	args := pgx.NamedArgs{"user_id": user.ID}
	rows, err := dbPoll.Query(context.Background(), query, args)
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
func GetRenderBoxByCards(dbPool *pgxpool.Pool, user User) (map[int]BoxesState, error) {
	// We can't make a LEFT JOIN to have `done` as false for a default value
	// Because we can't know `card_id`
	query :=
		`SELECT card_id, box_index, done
		FROM progress
		WHERE user_id = @user_id;`
	args := pgx.NamedArgs{
		"user_id": user.ID,
	}
	rows, err := dbPool.Query(context.Background(), query, args)
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
