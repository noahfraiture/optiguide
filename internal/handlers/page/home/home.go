package home

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"optiguide/internal/handlers"
	topbar "optiguide/internal/handlers/page"
	"sort"
	"text/template"
)

type HomeData struct {
	Team     []db.Character
	TeamSize int
	topbar.TopbarData
	CardData CardData // Data for the first card to be display
}

var funcsHome = template.FuncMap{
	"add": func(i, j int) int {
		return i + j
	},
	"iterate": func(max int) []int {
		r := make([]int, max)
		for i := range max {
			r[i] = i
		}
		return r
	},
	// Functions instead of `index . .` in html template, help to provide default value
	"doneAtIndex": func(boxes db.BoxesState, boxIndex int) bool {
		if box, ok := boxes[boxIndex]; ok {
			return box
		}
		return false
	},
	"characterAtIndex": func(boxes []db.Character, boxIndex int) db.Character {
		for _, box := range boxes {
			if box.BoxIndex == boxIndex {
				return box
			}
		}
		return db.Character{
			Class: db.NONE,
			Name:  fmt.Sprintf("Perso %d", boxIndex+1),
		}
	},
	"boxAtCard": func(boxes map[int]db.BoxesState, cardID int) db.BoxesState {
		if box, ok := boxes[cardID]; ok {
			return box
		}
		return db.BoxesState{}
	},
	"renderIcon": renderIcon,
	"map":        handlers.RenderMap,
}

func Home(w http.ResponseWriter, r *http.Request) {
	funcs := funcsHome
	for k, v := range topbar.FuncsTopbar {
		funcs[k] = v
	}
	for k, v := range funcsTeam {
		funcs[k] = v
	}

	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}

	userAuth, err := auth.GetUser(r)
	loggedIn := err == nil
	team := []db.Character{}
	var user db.User
	if loggedIn {
		// Get user
		user, err := db.GetUserFromProvider(dbPool, userAuth.Provider, userAuth.UserID)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Can't get user", http.StatusInternalServerError)
		}
		// Get team
		team, err = db.GetTeam(dbPool, user)
		sort.Slice(
			team,
			func(i, j int) bool { return team[i].BoxIndex < team[j].BoxIndex },
		)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Can't get boxes", http.StatusInternalServerError)
			return
		}
	}

	tmpl, err := template.
		New("base.html").
		Funcs(funcs).
		ParseFiles(
			"templates/base.html",
			"templates/topbar.html",
			"templates/home/home.html",
			"templates/home/team.html",
			"templates/home/card.html",
		)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base.html",
		HomeData{
			// We use empty cards and page=-1 to display nothing but the loader of the first page
			CardData:   CardData{Page: -1},
			Team:       team,
			TeamSize:   user.TeamSize,
			TopbarData: topbar.TopbarData{LoggedIn: loggedIn, Username: user.Username},
		},
	)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}
