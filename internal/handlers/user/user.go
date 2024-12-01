package user

import (
	"fmt"
	"html/template"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	topbar "optiguide/internal/handlers/page"
	"strconv"
)

func EditName(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.
		New("edit-name").
		Funcs(topbar.FuncsTopbar).
		ParseFiles("templates/topbar.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "edit-name", nil)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
		return
	}
}
func SaveName(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get db", http.StatusBadRequest)
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
		http.Error(w, "can't get user", http.StatusInternalServerError)
		return
	}
	err = db.UpdateUsername(dbPool, user, name)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "cna't update username", http.StatusInternalServerError)
		return
	}
	user.Username = name
	tmpl, err := template.
		New("name").
		Funcs(topbar.FuncsTopbar).
		ParseFiles("templates/topbar.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "name", user)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
		return
	}

}

func Toggle(w http.ResponseWriter, r *http.Request) {
	userAuth, err := auth.GetUser(r)
	if err != nil {
		msg := fmt.Sprintf("Could not get user: %v", err)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	cardStr := r.URL.Query().Get("card")
	if cardStr == "" {
		msg := fmt.Sprintf("No card index: %s", cardStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	cardIndex, err := strconv.Atoi(cardStr)
	if err != nil {
		msg := fmt.Sprintf("Invalid card ID: %s", cardStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	boxStr := r.URL.Query().Get("box")
	if boxStr == "" {
		fmt.Println("Can't get box")
		http.Error(w, "box", http.StatusBadRequest)
		return
	}
	boxIndex, err := strconv.Atoi(boxStr)
	if err != nil {
		fmt.Println("No box value")
		http.Error(w, "box value", http.StatusBadRequest)
		return
	}

	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println("Can't get db to toggle box")
		http.Error(w, "Can't get db to toggle box", http.StatusInternalServerError)
		return
	}

	user, err := db.GetUserFromProvider(dbPool, userAuth)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusInternalServerError)
		return
	}

	err = db.ToggleProgress(dbPool, user, cardIndex, boxIndex)
	if err != nil {
		fmt.Println("toggle error")
		http.Error(w, "toggle progress", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func ToggleAchievement(w http.ResponseWriter, r *http.Request) {
	userAuth, err := auth.GetUser(r)
	if err != nil {
		msg := fmt.Sprintf("Could not get user: %v", err)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	cardStr := r.URL.Query().Get("card")
	if cardStr == "" {
		msg := fmt.Sprintf("No card index: %s", cardStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	cardIndex, err := strconv.Atoi(cardStr)
	if err != nil {
		msg := fmt.Sprintf("Invalid card ID: %s", cardStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	achievementStr := r.URL.Query().Get("achievement")
	if achievementStr == "" {
		fmt.Println("Can't get box")
		http.Error(w, "box", http.StatusBadRequest)
		return
	}

	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println("Can't get db to toggle box")
		http.Error(w, "Can't get db to toggle box", http.StatusInternalServerError)
		return
	}

	user, err := db.GetUserFromProvider(dbPool, userAuth)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusInternalServerError)
		return
	}

	err = db.ToggleAchievement(dbPool, user, cardIndex, achievementStr)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "toggle achievement", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
