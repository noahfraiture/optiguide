package main

import (
	"log"
	"net/http"

	"optiguide/internal/auth"
	"optiguide/internal/db"
	"optiguide/internal/handlers/page/home"
	"optiguide/internal/handlers/user"
)

func main() {

	auth.Init()
	err := db.Init()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", home.Home)
	http.HandleFunc("/pick-class", home.PickClass)
	http.HandleFunc("/card", home.RenderCard)
	http.HandleFunc("/card/minus", home.Minus)
	http.HandleFunc("/card/plus", home.Plus)
	http.HandleFunc("/card/toggle", user.Toggle)
	http.HandleFunc("/auth/google/", user.GoogleLogin)
	http.HandleFunc("/auth/google/callback", user.GoogleCallback)
	http.HandleFunc("/logout", user.Logout)

	// FIXME : needed to have tailwindcss
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Fatal(http.ListenAndServe(":8080", nil))
}
