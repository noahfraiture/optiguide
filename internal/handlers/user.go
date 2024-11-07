package handlers

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
	box, err := strconv.Atoi(boxStr)
	if err != nil {
		http.Error(w, "box value", http.StatusBadRequest)
		return
	}

	// NOTE: could remove this call since we only need userid
	user, err := db.QueryUser(userAuth.UserID)
	fmt.Println(box)
	if err != nil {
		http.Error(w, "set progress", http.StatusBadRequest)
		return
	}
	err = user.SetProgress(cardID, box)
	if err != nil {
		http.Error(w, "set progress", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
