package handlers

import (
	"fmt"
	"net/http"

	"github.com/markbates/goth"
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
		http.Error(w, "error", http.StatusInternalServerError)
		return
	}
	fmt.Println(user)

	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

// Home function to check if the user is connected
func Home(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.GetFromSession("google", r)
	if err != nil || user == "" {
		http.Error(w, "User not logged in", http.StatusUnauthorized)
		return
	}

	// Display a connected message if user data is found
	fmt.Fprintf(w, "Welcome, you are logged in!")
}

// IsConnected function to check if user is connected
func IsConnected(r *http.Request) (bool, *goth.User) {
	user, err := gothic.CompleteUserAuth(nil, r)
	if err != nil {
		return false, nil
	}
	return true, &user
}
