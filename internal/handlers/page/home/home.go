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
	Cards []*db.Card
}

func Home(w http.ResponseWriter, r *http.Request) {

	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}

	userAuth, err := auth.GetUser(r)
	loggedIn := err == nil
	var team []db.Character
	var user db.User
	var cards []*db.Card
	if loggedIn {
		// Get user
		user, err = db.GetUserFromProvider(dbPool, userAuth)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Can't get user", http.StatusInternalServerError)
		}
		// Get team
		team, err = db.GetTeam(dbPool, user)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Can't get boxes", http.StatusInternalServerError)
			return
		}
		sort.Slice(
			team,
			func(i, j int) bool { return team[i].BoxIndex < team[j].BoxIndex },
		)

		cards, err = db.GetCards(dbPool, user, "")
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Can't get cards", http.StatusInternalServerError)
			return
		}
	}

	tmpl, err := template.
		New("base.html").
		Funcs(handlers.HtmlFuncs).
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
			Cards:      cards,
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

func SearchCards(w http.ResponseWriter, r *http.Request) {
	search := r.FormValue("name")

	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get db", http.StatusInternalServerError)
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
		http.Error(w, "Can't get user", http.StatusInternalServerError)
	}

	cards, err := db.GetCards(dbPool, user, search)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get cards", http.StatusInternalServerError)
		return
	}

	team, err := db.GetTeam(dbPool, user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get boxes", http.StatusInternalServerError)
		return
	}
	sort.Slice(
		team,
		func(i, j int) bool { return team[i].BoxIndex < team[j].BoxIndex },
	)

	tmpl, err := template.
		New("card.html").
		Funcs(handlers.HtmlFuncs).
		ParseFiles(
			"templates/home/card.html",
			"templates/home/team.html", // needed for character icon
		)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "cards",
		// Here we can ignore other parameter that won't be used
		HomeData{
			Cards: cards,
			Team:  team,
		},
	)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}
