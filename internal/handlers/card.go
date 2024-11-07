package handlers

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"strconv"

	"html/template"
)

func Plus(w http.ResponseWriter, r *http.Request) {
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	user, err := db.QueryUser(userSesssion.UserID)
	if err != nil {
		http.Error(w, "error for user", http.StatusBadRequest)
		return
	}
	// TODO: error message or block case
	if user.TeamSize == 32 {
		renderCard(w, 0, user)
		return
	}
	err = user.PlusTeamSize(1)
	if err != nil {
		http.Error(w, "user plus", http.StatusBadRequest)
		return
	}
	user.TeamSize += 1
	renderCard(w, 0, user)
}
func Minus(w http.ResponseWriter, r *http.Request) {
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	user, err := db.QueryUser(userSesssion.UserID)
	if err != nil {
		http.Error(w, "error for user", http.StatusBadRequest)
		return
	}
	if user.TeamSize == 1 {
		renderCard(w, 0, user)
		return
	}
	err = user.PlusTeamSize(-1)
	if err != nil {
		http.Error(w, "user minus", http.StatusBadRequest)
		return
	}
	user.TeamSize -= 1
	renderCard(w, 0, user)
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
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	user, err := db.QueryUser(userSesssion.UserID)
	if err != nil {
		msg := fmt.Sprintf("Card not found: %d", page)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	renderCard(w, page, user)
}

func renderCard(w http.ResponseWriter, page int, user db.User) {
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
			"iterate": func(max int) []int {
				r := make([]int, max)
				for i := range max {
					r[i] = i
				}
				return r
			},
			"and": func(value, i int) bool {
				return value&(1<<i) != 0
			},
		}).
		ParseFiles("templates/card.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, map[string]any{"Cards": cardsDone, "Page": page, "Size": user.TeamSize})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

}
