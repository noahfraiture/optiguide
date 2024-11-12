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
		http.Error(w, msg, http.StatusBadRequest)
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
		fmt.Println("Can't get db")
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}

	err = db.ToggleProgress(dbPool, userAuth.UserID, cardID, boxIndex)
	if err != nil {
		fmt.Println("toggle error")
		http.Error(w, "toggle progress", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
