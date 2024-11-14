package home

import (
	"fmt"
	"html/template"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"strconv"
)

var funcsTeam = template.FuncMap{
	"renderIcon": renderIcon,
	"nbClass":    func() int { return int(db.NB_CLASS) },
	"add":        func(i, j int) int { return i + j },
	"iterate": func(max int) []int {
		r := make([]int, max)
		for i := range max {
			r[i] = i
		}
		return r
	},
}

func renderIcon(class any, boxIndex int) template.HTML {
	var i int
	switch v := class.(type) {
	case int:
		i = v
	case db.Class:
		i = int(v)
	default:
		return ""
	}
	return template.HTML(fmt.Sprintf(`
		<img src="/static/images/%[2]s.avif" alt=%[2]s class="inline-block h-6 w-6 mr-2">
		%[2]s
		`,
		boxIndex,
		nameFromClass(i),
	))
}

func nameFromClass(class int) string {
	return db.ClassToName[db.Class(class)]
}

func PickClass(w http.ResponseWriter, r *http.Request) {

	classStr := r.URL.Query().Get("class")
	if classStr == "" {
		msg := fmt.Sprintf("No class: %s", classStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	class, err := strconv.Atoi(classStr)
	if err != nil {
		msg := fmt.Sprintf("Invalid class value: %s", classStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	indexStr := r.URL.Query().Get("index")
	if indexStr == "" {
		msg := fmt.Sprintf("No index: %s", indexStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		msg := fmt.Sprintf("Invalid index value: %s", indexStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusBadRequest)
		return
	}
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get db", http.StatusBadRequest)
		return
	}
	err = db.UpdateClass(dbPool, userAuth.UserID, index, db.Class(class))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't update class", http.StatusBadRequest)
		return
	}

	iconHTML := renderIcon(db.Class(class), index)
	tmpl := fmt.Sprintf(
		`<div hx-swap-oob="outerHTML" id="icon-%[1]d">%[2]s</div>`,
		index,
		iconHTML,
	)
	_, err = w.Write([]byte(tmpl))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}

func Plus(w http.ResponseWriter, r *http.Request) {
	dbPool, err := db.GetPool()
	if err != nil {
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	user, err := db.QueryUser(dbPool, userSesssion.UserID)
	if err != nil {
		http.Error(w, "error for user", http.StatusBadRequest)
		return
	}
	// TODO: block on more that 32
	err = db.PlusTeamSize(dbPool, user.ID, 1)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "user plus", http.StatusBadRequest)
		return
	}
	team, err := db.GetClasses(dbPool, user.ID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "user plus", http.StatusBadRequest)
		return
	}
	user.TeamSize += 1

	tmpl, err := template.
		New("home.html").
		Funcs(funcsHome).
		ParseFiles("templates/home.html", "templates/card.html", "templates/team.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "body", HomeData{
		Team:     team,
		LoggedIn: true,
		CardData: CardData{Page: -1},
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
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	user, err := db.QueryUser(dbPool, userSesssion.UserID)
	if err != nil {
		http.Error(w, "error for user", http.StatusBadRequest)
		return
	}
	if user.TeamSize == 1 {
		renderCard(w, 0, user)
		return
	}
	err = db.PlusTeamSize(dbPool, user.ID, -1)
	if err != nil {
		http.Error(w, "user minus", http.StatusBadRequest)
		return
	}
	user.TeamSize -= 1
	tmpl := fmt.Sprintf(`<div hx-swap-oob="delete" id="character-picker-%d"></div>`, user.TeamSize)
	_, err = w.Write([]byte(tmpl))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
		return
	}
}
