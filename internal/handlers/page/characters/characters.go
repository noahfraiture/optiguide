package characters

import (
	"fmt"
	"html/template"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"optiguide/internal/handlers"
	topbar "optiguide/internal/handlers/page"
	"sort"
)

type CharactersData struct {
	Characters []db.Character
	topbar.TopbarData
}

var funcsCharacters = template.FuncMap{
	"emptyArr": func() []any { return []any{} },
	"map":      handlers.RenderMap,
	"className": func(class db.Class) string {
		return db.ClassToName[class]
	},
	"add": func(a int, b int) int { return a + b },
}

func Characters(w http.ResponseWriter, r *http.Request) {
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}

	userAuth, err := auth.GetUser(r)
	loggedIn := err == nil
	var characters []db.Character
	var user db.User
	if loggedIn {
		user, err = db.GetUserFromProvider(dbPool, userAuth)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Can't get user", http.StatusInternalServerError)
			return
		}
		characters, err = db.GetTeam(dbPool, user)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "Can't get characters", http.StatusInternalServerError)
			return
		}
		sort.Slice(
			characters,
			func(i, j int) bool { return characters[i].BoxIndex < characters[j].BoxIndex },
		)
	}

	tmpl, err := template.
		New("base.html").
		Funcs(funcsCharacters).
		ParseFiles(
			"templates/base.html",
			"templates/topbar.html",
			"templates/characters/characters.html",
		)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	err = tmpl.ExecuteTemplate(w, "base.html", CharactersData{
		Characters: characters,
		TopbarData: topbar.TopbarData{LoggedIn: loggedIn, Username: user.Username},
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Error executing template", http.StatusInternalServerError)
		return
	}
} 