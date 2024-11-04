package main

import (
	"log"
	"net/http"
	"text/template"

	"optimax/internal/auth"
	"optimax/internal/db"
	"optimax/internal/handlers"
)

func main() {

	auth.Init()
	err := db.Init()
	if err != nil {
		log.Fatal(err)
	}

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/card", handlers.RenderCard)
	http.HandleFunc("/toggle", handlers.Toggle)
	http.HandleFunc("/auth/google/", handlers.GoogleLogin)
	http.HandleFunc("/auth/google/callback", handlers.GoogleCallback)
	http.HandleFunc("/logout", handlers.Logout)

	// FIXME : needed to have tailwindcss
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func serveIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.
		New("base.html").
		Funcs(template.FuncMap{
			"renderAuthButton": handlers.RenderAuthButton,
			"isLoggedIn": func() bool {
				_, err := auth.GetUser(r)
				return err == nil
			},
		}).
		ParseFiles("templates/base.html", "templates/topbar.html", "templates/home.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	_, err = auth.GetUser(r)
	data := map[string]interface{}{
		"IsLoggedIn": err == nil,
	}

	// Execute the template without any data, since index.html does not require it
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}
