package home

import (
	"fmt"
	"html/template"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"optiguide/internal/handlers"
)

var funcsTeam = template.FuncMap{
	"className": className,
	"nbClass":   func() int { return int(db.NB_CLASS) },
	"add":       func(i, j int) int { return i + j },
	"iterate": func(max int) []int {
		r := make([]int, max)
		for i := range max {
			r[i] = i
		}
		return r
	},
}

func SaveName(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	index, err := handlers.GetParameterInt(w, r, "index")
	if err != nil {
		return
	}
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get db", http.StatusBadRequest)
		return
	}
	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusUnauthorized)
		return
	}
	user, err := db.GetUserFromProvider(dbPool, userAuth)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusInternalServerError)
		return
	}
	err = db.UpdateCharacterName(dbPool, user, index, name)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't update char", http.StatusBadRequest)
	}
	tmpl, err := template.
		New("swap-name").
		Funcs(funcsHome).
		ParseFiles("templates/home/team.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "swap-name", map[string]any{
		"Index": index,
		"Name":  name,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
		return
	}
}

func RenderEditableName(w http.ResponseWriter, r *http.Request) {
	name, err := handlers.GetParameterString(w, r, "name")
	if err != nil {
		return
	}
	index, err := handlers.GetParameterInt(w, r, "index")
	if err != nil {
		return
	}
	tmpl, err := template.
		New("editable-name").
		Funcs(funcsHome).
		ParseFiles("templates/home/team.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "editable-name", map[string]any{
		"Index": index,
		"Name":  name,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
		return
	}
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

func PickCharacter(w http.ResponseWriter, r *http.Request) {
	class, err := handlers.GetParameterInt(w, r, "class")
	if err != nil {
		return
	}
	index, err := handlers.GetParameterInt(w, r, "index")
	if err != nil {
		return
	}
	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusUnauthorized)
		return
	}
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get db", http.StatusBadRequest)
		return
	}
	user, err := db.GetUserFromProvider(dbPool, userAuth)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusInternalServerError)
		return
	}
	err = db.UpdateCharacterClass(dbPool, user, index, db.Class(class))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't update class", http.StatusBadRequest)
		return
	}

	tmpl, err := template.
		New("swap-icon").
		Funcs(funcsHome).
		ParseFiles("templates/home/team.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "swap-icon", map[string]any{
		"Class": db.Class(class),
		"Index": index,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
		return
	}
}

type PlusData struct {
	MaxCardID int
	Team      []db.Character
	Boxes     map[int]db.BoxesState
	BoxIndex  int
}

func Plus(w http.ResponseWriter, r *http.Request) {
	dbPool, err := db.GetPool()
	if err != nil {
		http.Error(w, "Can't get db to plus", http.StatusInternalServerError)
		return
	}
	userAuth, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}
	user, err := db.GetUserFromProvider(dbPool, userAuth)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusInternalServerError)
		return
	}
	if user.TeamSize >= 32 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	err = db.PlusTeamSize(dbPool, user, 1)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "user plus", http.StatusBadRequest)
		return
	}

	team, err := db.GetTeam(dbPool, user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "user plus", http.StatusBadRequest)
		return
	}

	user.TeamSize += 1
	boxes, err := db.GetRenderBoxByCards(dbPool, user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "boxes", http.StatusBadRequest)
		return
	}

	tmpl, err := template.
		New("swap").
		Funcs(funcsHome).
		ParseFiles(
			"templates/home/home.html",
			"templates/home/team.html",
			"templates/home/card.html",
		)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "swap", PlusData{
		MaxCardID: 10,
		Team:      team,
		Boxes:     boxes,
		BoxIndex:  user.TeamSize - 1,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
		return
	}
}

func Minus(w http.ResponseWriter, r *http.Request) {
	dbPool, err := db.GetPool()
	if err != nil {
		http.Error(w, "Can't get db to minus", http.StatusInternalServerError)
		return
	}
	userAuth, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}
	user, err := db.GetUserFromProvider(dbPool, userAuth)
	if err != nil {
		http.Error(w, "error for user", http.StatusBadRequest)
		return
	}
	if user.TeamSize <= 1 {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	err = db.PlusTeamSize(dbPool, user, -1)
	if err != nil {
		http.Error(w, "user minus", http.StatusBadRequest)
		return
	}
	user.TeamSize -= 1
	tmpl, err := template.
		New("delete-character").
		Funcs(funcsHome).
		ParseFiles("templates/home/team.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "delete-character", user.TeamSize)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
		return
	}
}
