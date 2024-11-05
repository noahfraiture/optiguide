package handlers

import (
	"fmt"
	"net/http"
	"optimax/internal/auth"
	"optimax/internal/db"
	"strconv"

	"html/template"
)

// TODO
func RenderCard(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		msg := fmt.Sprintf("No card ID: %s", idParam)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idParam)
	if err != nil {
		msg := fmt.Sprintf("Invalid card ID: %s", idParam)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	card, err := db.GetCard(id)
	if err != nil {
		msg := fmt.Sprintf("Card not found: %d", id)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := fmt.Sprintf("User not found: %s", idParam)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	user := db.User{ID: userSesssion.UserID}
	done, err := user.IsStepDone(card.ID)
	if err != nil {
		msg := "Card done error"
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

	err = tmpl.Execute(w, map[string]any{"Card": card, "Checked": done})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
}
