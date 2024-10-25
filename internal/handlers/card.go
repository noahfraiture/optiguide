package handlers

import (
	"fmt"
	"net/http"
	"optimax/internal/parser"
	"strconv"

	"html/template"
)

func RenderCard(w http.ResponseWriter, r *http.Request) {
	idParam := r.URL.Query().Get("id")
	if idParam == "" {
		http.Error(w, "No card ID", http.StatusBadRequest)
		return
	}
	id, err := strconv.Atoi(idParam)
	if err != nil {
		http.Error(w, "Invalid card ID", http.StatusBadRequest)
	}
	card, err := parser.GetCard(id)
	if err != nil {
		http.Error(w, "Error in parsing", http.StatusInternalServerError)
		return
	}
	tmpl, err := template.
		New("card.html").
		Funcs(template.FuncMap{
			"inc": func(i int) int {
				return i + 1
			},
		}).
		ParseFiles("templates/card.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.Execute(w, card)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
}
