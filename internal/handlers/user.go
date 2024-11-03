package handlers

import (
	"fmt"
	"net/http"
	"optimax/internal/auth"
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

	step := r.URL.Query().Get("step")
	if step == "" {
		msg := fmt.Sprintf("No step ID: %s", step)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	stepID, err := strconv.Atoi(step)
	if err != nil {
		msg := fmt.Sprintf("Invalid step ID: %s", step)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	var userProgress auth.UserProgress
	var ok bool
	if userProgress, ok = auth.Progress[userAuth.UserID]; !ok {
		userProgress = auth.UserProgress{
			UserID: userAuth.UserID,
			Steps:  make(map[int]bool),
		}
	}
	progress, ok := userProgress.Steps[stepID]
	userProgress.Steps[stepID] = !ok || !progress
	auth.Progress[userProgress.UserID] = userProgress
	w.WriteHeader(http.StatusNoContent)
}
