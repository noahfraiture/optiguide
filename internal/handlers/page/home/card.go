package home

import (
	"fmt"
	"net/http"
	"optiguide/internal/db"
	"optiguide/internal/handlers"
	"strconv"

	"html/template"
)

type CardData struct {
	Cards      []db.Card
	Team       []db.Character
	BoxesState map[int]db.BoxesState
	Page       int
	Size       int
}

func RenderCard(w http.ResponseWriter, r *http.Request) {
	pageParam := r.URL.Query().Get("page")
	if pageParam == "" {
		msg := fmt.Sprintf("No card ID: %s", pageParam)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	page, err := strconv.Atoi(pageParam)
	if err != nil {
		msg := fmt.Sprintf("Invalid card ID: %s", pageParam)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	user, err := handlers.GetUser(w, r)
	if err != nil {
		return
	}
	renderCard(w, page, user)
}

func renderCard(w http.ResponseWriter, page int, user db.User) {
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println("can't get db in render")
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}

	cards, err := db.GetCards(dbPool, page)
	if err != nil {
		fmt.Printf("Can't get cards %s\n", err)
		http.Error(w, "Can't get cards", http.StatusInternalServerError)
		return
	}
	boxes, err := db.GetRenderBoxByCards(dbPool, user)
	if err != nil {
		fmt.Printf("Can't get boxes %s\n", err)
		http.Error(w, "Can't get boxes", http.StatusInternalServerError)
		return
	}

	team, err := db.GetTeam(dbPool, user)
	if err != nil {
		fmt.Println("Can't get team", err)
		http.Error(w, "Can't get team", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.
		New("card.html").
		Funcs(funcsHome).
		ParseFiles(
			"templates/home/card.html",
			"templates/home/team.html",
		)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "cards", CardData{
		Cards:      cards,
		Team:       team,
		BoxesState: boxes,
		Page:       page,
		Size:       user.TeamSize,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
}
