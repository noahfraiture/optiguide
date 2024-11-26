package user

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"strconv"
)

func Toggle(w http.ResponseWriter, r *http.Request) {
	userAuth, err := auth.GetUser(r)
	if err != nil {
		msg := fmt.Sprintf("Could not get user: %v", err)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusUnauthorized)
		return
	}

	card := r.URL.Query().Get("card")
	if card == "" {
		msg := fmt.Sprintf("No card ID: %s", card)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	cardID, err := strconv.Atoi(card)
	if err != nil {
		msg := fmt.Sprintf("Invalid card ID: %s", card)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	boxStr := r.URL.Query().Get("box")
	if boxStr == "" {
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

	user, err := db.GetUserFromProvider(dbPool, userAuth.Provider, userAuth.UserID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusInternalServerError)
		return
	}

	err = db.ToggleProgress(dbPool, user, cardID, boxIndex)
	if err != nil {
		fmt.Println("toggle error")
		http.Error(w, "toggle progress", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
