package user

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"

	"github.com/markbates/goth/gothic"
)

func GoogleLogin(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	q.Add("provider", "google")
	r.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(w, r)
}

func GoogleCallback(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	q.Add("provider", "google")
	r.URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(w, r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error during authentication", http.StatusInternalServerError)
		return
	}
	dbPool, err := db.GetPool()
	if err != nil {
		http.Error(w, "Can't get db in callback", http.StatusInternalServerError)
		return
	}
	err = auth.SaveUser(dbPool, user, w, r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error during saving user", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func Logout(w http.ResponseWriter, r *http.Request) {
	// Clear the session
	err := auth.ClearSession(w, r)
	if err != nil {
		http.Error(w, "Error during logout", http.StatusInternalServerError)
		return
	}

	// Perform logout using gothic
	err = gothic.Logout(w, r)
	if err != nil {
		http.Error(w, "Error during logout", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}
