package home

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"strconv"

	"html/template"
)

func RenderCard(w http.ResponseWriter, r *http.Request) {
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println("can't get db in Render")
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}
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
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	user, err := db.QueryUser(dbPool, userSesssion.UserID)
	if err != nil {
		msg := fmt.Sprintf("Card not found: %d", page)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
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
	boxes, err := db.GetRenderBoxByCards(dbPool, user.ID)
	if err != nil {
		fmt.Printf("Can't get boxes %s\n", err)
		http.Error(w, "Can't get boxes", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.
		New("card.html").
		Funcs(funcsHome).
		ParseFiles("templates/card.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]any{
		"Cards": cards,
		"Boxes": boxes,
		"Page":  page,
		"Size":  user.TeamSize,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
}
