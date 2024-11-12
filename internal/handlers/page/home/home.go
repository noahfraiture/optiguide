package home

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	topbar "optiguide/internal/handlers/page"
	"sort"
	"text/template"
)

type HomeData struct {
	Boxes    []db.UserBox
	LoggedIn bool
}

var funcsHome = template.FuncMap{
	"add": func(i, j int) int {
		return i + j
	},
	"minus": func(i, j int) int {
		return i - j
	},
	"iterate": func(max int) []int {
		r := make([]int, max)
		for i := range max {
			r[i] = i
		}
		return r
	},
	// Functions instead of `index . .` in html template, help to provide default value
	"doneAt": func(boxes map[int]db.BoxState, index int) bool {
		if box, ok := boxes[index]; ok {
			return box.Done
		}
		return false
	},
	"classAt": func(boxes map[int]db.BoxState, i int) db.Class {
		if box, ok := boxes[i]; ok {
			return box.Class
		}
		return db.NONE
	},
}

func Home(w http.ResponseWriter, r *http.Request) {
	funcs := funcsHome
	for k, v := range topbar.FuncsTopbar {
		funcs[k] = v
	}
	for k, v := range funcsTeam {
		funcs[k] = v
	}
	tmpl, err := template.
		New("base.html").
		Funcs(funcs).
		ParseFiles("templates/base.html", "templates/topbar.html", "templates/home.html", "templates/team.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}

	data := HomeData{}

	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		data.LoggedIn = false
	} else {
		data.LoggedIn = true
		data.Boxes, err = db.GetClasses(dbPool, userAuth.UserID)
		sort.Slice(
			data.Boxes,
			func(i, j int) bool { return data.Boxes[i].BoxIndex < data.Boxes[j].BoxIndex },
		)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Can't get boxes", http.StatusInternalServerError)
			return
		}
	}
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		fmt.Println(err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}
