package handlers

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"sort"
	"text/template"
)

type HomeData struct {
	Boxes    []db.UserBox
	LoggedIn bool
	NbClass  db.Class
}

func Home(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.
		New("base.html").
		Funcs(template.FuncMap{
			"renderAuthButton": renderAuthButton,
			"nameFromClass":    nameFromClass,
			"inc": func(i int) int {
				return i + 1
			},
			"iterate": func(max db.Class) []db.Class {
				r := make([]db.Class, max)
				for i := range max {
					r[i] = i
				}
				return r
			},
		}).
		ParseFiles("templates/base.html", "templates/topbar.html", "templates/home.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}

	data := HomeData{NbClass: db.NB_CLASS}

	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		data.LoggedIn = false
	} else {
		data.LoggedIn = true
		data.Boxes, err = db.GetClasses(dbPool, userAuth.UserID)
		sort.Slice(
			data.Boxes,
			func(i, j int) bool { return data.Boxes[i].BoxIndex < data.Boxes[j].BoxIndex },
		)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Can't get boxes", http.StatusInternalServerError)
			return
		}
	}
	if err := tmpl.ExecuteTemplate(w, "base.html", data); err != nil {
		fmt.Println(err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
}

func renderAuthButton(isLoggedIn bool) string {
	if isLoggedIn {
		return `<a href="/logout" class="bg-red-500 hover:bg-red-600 text-white font-bold py-2 px-4 rounded inline-block text-center">Logout</a>`
	}
	return `<a href="/auth/google" class="bg-blue-500 hover:bg-blue-600 text-white font-bold py-2 px-4 rounded inline-block text-center">Login with Google</a>`
}

func nameFromClass(class db.Class) string {
	return db.ClassToName[db.Class(class)]
}
