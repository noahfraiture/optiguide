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
	Team     []db.TeamBox
	LoggedIn bool
	CardData CardData // Data for the first card to be display
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
	"doneAt": func(boxes db.BoxesState, boxIndex int) bool {
		if box, ok := boxes[boxIndex]; ok {
			return box
		}
		return false
	},
	"classAt": func(boxes []db.TeamBox, boxIndex int) db.Class {
		for _, box := range boxes {
			if box.BoxIndex == boxIndex {
				return box.Class
			}
		}
		return db.TOUS
	},
	"boxAt": func(boxes map[int]db.BoxesState, cardID int) db.BoxesState {
		if box, ok := boxes[cardID]; ok {
			return box
		}
		return db.BoxesState{}
	},
	"renderIcon": renderIcon,
	"map": func(args ...any) map[string]any {
		dict := make(map[string]any)
		for i := range len(args) / 2 {
			if v, ok := args[i*2].(string); ok {
				dict[v] = args[i*2+1]
			}
		}
		return dict
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
		ParseFiles(
			"templates/base.html",
			"templates/topbar.html",
			"templates/home.html",
			"templates/team.html",
			"templates/card.html",
		)
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

	// We use empty cards and page=-1 to display nothing but the loader of the first page
	data := HomeData{CardData: CardData{
		Page: -1,
	}}

	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		data.LoggedIn = false
	} else {
		data.LoggedIn = true
		data.Team, err = db.GetClasses(dbPool, userAuth.UserID)
		sort.Slice(
			data.Team,
			func(i, j int) bool { return data.Team[i].BoxIndex < data.Team[j].BoxIndex },
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
