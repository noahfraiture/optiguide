package main

import (
	"log"
	"net/http"

	"optiguide/internal/auth"
	"optiguide/internal/db"
	"optiguide/internal/handlers"
)

func main() {

	auth.Init()
	err := db.Init()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", handlers.Home)
	http.HandleFunc("/card", handlers.RenderCard)
	http.HandleFunc("/minus", handlers.Minus)
	http.HandleFunc("/plus", handlers.Plus)
	http.HandleFunc("/toggle", handlers.Toggle)
	http.HandleFunc("/auth/google/", handlers.GoogleLogin)
	http.HandleFunc("/auth/google/callback", handlers.GoogleCallback)
	http.HandleFunc("/logout", handlers.Logout)

	// FIXME : needed to have tailwindcss
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
