package handlers

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"

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
		http.Error(w, "Error during authentication", http.StatusInternalServerError)
		return
	}
	err = auth.SaveUser(user, w, r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error during saving user", http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
}

func RenderAuthButton(isLoggedIn bool) string {
	if isLoggedIn {
		return `<a href="/logout" class="bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded inline-block text-center">Logout</a>`
	}
	return `<a href="/auth/google" class="bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded inline-block text-center">Login with Google</a>`
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
