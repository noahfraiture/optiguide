package about

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"optiguide/internal/handlers"
	topbar "optiguide/internal/handlers/page"
	"strings"
)

type aboutData struct {
	LoggedIn bool
}

var funcsAbout = template.FuncMap{
	"emptyArr": func() []any { return []any{} },
	"map":      handlers.RenderMap,
}

func About(w http.ResponseWriter, r *http.Request) {
	funcs := funcsAbout
	for k, v := range topbar.FuncsTopbar {
		funcs[k] = v
	}

	_, err := auth.GetUser(r)
	loggedIn := err == nil

	tmpl, err := template.New("base.html").
		Funcs(funcs).
		ParseFiles(
			"templates/base.html",
			"templates/topbar.html",
			"templates/about/about.html",
		)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", aboutData{
		LoggedIn: loggedIn,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		return
	}
}

func SubmitFeedback(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, "Can't parse form", http.StatusBadRequest)
		return
	}

	feedback := r.FormValue("feedback")
	if strings.TrimSpace(feedback) == "" {
		tmpl, err := template.New("feedback-fail").ParseFiles("templates/about/about.html")
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Unable to load template", http.StatusInternalServerError)
			return
		}

		err = tmpl.ExecuteTemplate(w, "feedback-fail", nil)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		}
		return
	}

	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}

	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Failed to get user", http.StatusUnauthorized)
		return
	}

	if err := db.StoreFeedback(dbPool, context.Background(), feedback, userAuth.UserID); err != nil {
		http.Error(w, "Failed to store feedback", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("feedback-success").ParseFiles("templates/about/about.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "feedback-success", nil)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
	}
}
