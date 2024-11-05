package handlers

import (
	"fmt"
	"net/http"
	"optimax/internal/auth"
	"optimax/internal/db"
	"strconv"
)

func Toggle(w http.ResponseWriter, r *http.Request) {
	userAuth, err := auth.GetUser(r)
	if err != nil {
		msg := fmt.Sprintf("Could not get user: %v", err)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	card := r.URL.Query().Get("id")
	if card == "" {
		msg := fmt.Sprintf("No step ID: %s", card)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	cardID, err := strconv.Atoi(card)
	if err != nil {
		msg := fmt.Sprintf("Invalid step ID: %s", card)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	u := db.User{ID: userAuth.UserID}
	u.ToggleProgress(cardID)
	w.WriteHeader(http.StatusNoContent)
}
