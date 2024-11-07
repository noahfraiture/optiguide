package handlers

import (
	"fmt"
	"net/http"
	"optimax/internal/auth"
	"optimax/internal/db"
	"strconv"

	"html/template"
)

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
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := fmt.Sprintf("User not found: %s", pageParam)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	user := db.User{ID: userSesssion.UserID}
	cardsDone, err := user.GetPage(page)
	if err != nil {
		msg := fmt.Sprintf("Card not found: %d", page)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	tmpl, err := template.
		New("card.html").
		Funcs(template.FuncMap{
			"inc": func(i int) int {
				return i + 1
			},
			"minus": func(i, j int) int {
				return i - j
			},
			"greater": func(i, j int) bool {
				return i > j
			},
		}).
		ParseFiles("templates/card.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]any{"Cards": cardsDone, "Page": page})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
}
