package handlers

import (
	"fmt"
	"html/template"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"strconv"

	"github.com/google/uuid"
)

var HtmlFuncs = template.FuncMap{
	// home.html,team.html,card.html
	"iterate": func(max int) []int {
		r := make([]int, max)
		for i := range max {
			r[i] = i
		}
		return r
	},
	// home.html. Checkbox status
	"doneAtIndex": func(boxes db.BoxesState, boxIndex int) bool {
		if box, ok := boxes[boxIndex]; ok {
			return box
		}
		return false
	},
	// home.html,team.html,card.html
	"characterAtIndex": func(boxes []db.Character, boxIndex int) db.Character {
		for _, box := range boxes {
			if box.BoxIndex == boxIndex {
				return box
			}
		}
		return db.Character{
			Class:    db.NONE,
			Name:     fmt.Sprintf("Perso %d", boxIndex+1),
			BoxIndex: boxIndex,
		}
	},
	// home.html
	"boxAtCard": func(boxes map[int]db.BoxesState, cardID int) db.BoxesState {
		if box, ok := boxes[cardID]; ok {
			return box
		}
		return db.BoxesState{}
	},
	"map": renderMap,

	// team.html
	"className": className,
	"nbClass":   func() int { return int(db.NB_CLASS) },

	// topbar.html. Used to have a default name to avoid having unclickable name
	"displayName": func(name string) string {
		if name == "" {
			return "Inconnu"
		}
		return name
	},

	// card.html,team.html,character.html. Used to have a default name to avoid having unclickable name
	"displayCharacterName": func(name string, index int) string {
		if name == "" {
			return fmt.Sprintf("Perso %d", index+1)
		}
		return name
	},

	// guild.html. Used to render the research result when there's no research yet
	"emptyArr": func() []any { return []any{} },
}

func className(class any) template.HTML {
	var i int
	switch v := class.(type) {
	case int:
		i = v
	case db.Class:
		i = int(v)
	default:
		return ""
	}
	return template.HTML(db.ClassToName[db.Class(i)])
}
func renderMap(args ...any) map[string]any {
	dict := make(map[string]any)
	for i := range len(args) / 2 {
		if v, ok := args[i*2].(string); ok {
			dict[v] = args[i*2+1]
		}
	}
	return dict
}

// Return the parameter and write error status in http method if error
func GetParameterInt(w http.ResponseWriter, r *http.Request, name string) (int, error) {
	valueStr := r.URL.Query().Get(name)
	if valueStr == "" {
		msg := fmt.Sprintf("No %s: %s", name, valueStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return 0, fmt.Errorf("%s", msg)
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		msg := fmt.Sprintf("No %s: %s", name, valueStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return 0, fmt.Errorf("%s", msg)
	}
	return value, nil
}

// Return the parameter and write error status in http method if error
func GetParameterString(w http.ResponseWriter, r *http.Request, name string) string {
	return r.URL.Query().Get(name)
}

// Return the parameter and write error status in http method if error
func GetParameterUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, error) {
	valueStr := r.URL.Query().Get(name)
	if valueStr == "" {
		msg := fmt.Sprintf("No %s: %s", name, valueStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return [16]byte{}, fmt.Errorf("%s", msg)
	}
	value, err := uuid.Parse(valueStr)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't parse uuid", http.StatusBadRequest)
		return [16]byte{}, err
	}
	return value, nil
}

// Get the db, the session and the user from the db and write http error if there's an error
func GetUser(w http.ResponseWriter, r *http.Request) (db.User, error) {
	dbPool, err := db.GetPool()
	if err != nil {
		http.Error(w, "Can't get db to get user", http.StatusInternalServerError)
		return db.User{}, err
	}
	userAuth, err := auth.GetUser(r)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return db.User{}, nil
	}
	user, err := db.GetUserFromProvider(dbPool, userAuth)
	if err != nil {
		http.Error(w, "error for user", http.StatusBadRequest)
		return db.User{}, nil
	}
	return user, err
}
