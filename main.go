package main

import (
	"log"
	"net/http"
	"text/template"

	"optimax/internal/auth"
	"optimax/internal/handlers"
)

func main() {
	auth.Init()

	http.HandleFunc("/", serveIndex)
	http.HandleFunc("/card", handlers.RenderCard)
	http.HandleFunc("/home", handlers.Home)
	http.HandleFunc("/auth/google/", handlers.GoogleLogin)
	http.HandleFunc("/auth/google/callback", handlers.GoogleCallback)

	// FIXME : needed to have tailwindcss
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	log.Println("Server is running on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// serveIndex serves the index.html file
func serveIndex(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	// Execute the template without any data, since index.html does not require it
	if err := tmpl.Execute(w, nil); err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
}
