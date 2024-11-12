package home

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"strconv"
	"text/template"
)

var funcsTeam = template.FuncMap{
	"nameFromClass": nameFromClass,
	"renderIcon":    renderIcon,
	"nbClass":       func() int { return int(db.NB_CLASS) },
	"add":           func(i, j int) int { return i + j },
	"iterate": func(max int) []int {
		r := make([]int, max)
		for i := range max {
			r[i] = i
		}
		return r
	},
}

func renderIcon(class db.Class, boxIndex int) string {
	return fmt.Sprintf(`<img id="icon-%d" src="/static/images/%s.avif"</button>`, boxIndex, nameFromClass(int(class)))
}

func nameFromClass(class int) string {
	return db.ClassToName[db.Class(class)]
}

func PickClass(w http.ResponseWriter, r *http.Request) {

	classStr := r.URL.Query().Get("class")
	if classStr == "" {
		msg := fmt.Sprintf("No class: %s", classStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	class, err := strconv.Atoi(classStr)
	if err != nil {
		msg := fmt.Sprintf("Invalid class value: %s", classStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	indexStr := r.URL.Query().Get("index")
	if indexStr == "" {
		msg := fmt.Sprintf("No index: %s", indexStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		msg := fmt.Sprintf("Invalid index value: %s", indexStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusBadRequest)
		return
	}
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get db", http.StatusBadRequest)
		return
	}
	err = db.UpdateClass(dbPool, userAuth.UserID, index, db.Class(class))
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't update class", http.StatusBadRequest)
		return
	}

	iconHTML := renderIcon(db.Class(class), index)
	tmpl, err := template.New("iconHTML").Parse(iconHTML)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "bad render", http.StatusBadRequest)
		return
	}
	err = tmpl.Execute(w, nil)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "bad render", http.StatusBadRequest)
		return
	}
}

func Plus(w http.ResponseWriter, r *http.Request) {
	dbPool, err := db.GetPool()
	if err != nil {
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	user, err := db.QueryUser(dbPool, userSesssion.UserID)
	if err != nil {
		http.Error(w, "error for user", http.StatusBadRequest)
		return
	}
	// TODO: block on more that 32
	err = db.PlusTeamSize(dbPool, user.ID, 1)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "user plus", http.StatusBadRequest)
		return
	}
	userBox, err := db.InsertClass(dbPool, user.ID, user.TeamSize, db.NONE)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "set class plus", http.StatusBadRequest)
		return
	}
	user.TeamSize += 1

	// We must parse team and picker to have the context for picker
	tmpl, err := template.New("class-picker").
		Funcs(funcsTeam).
		ParseFiles("templates/team.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template parsing error", http.StatusInternalServerError)
		return
	}

	// Then we execute only the template picker
	err = tmpl.ExecuteTemplate(w, "picker", userBox)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "template execution error", http.StatusInternalServerError)
		return
	}
}

func Minus(w http.ResponseWriter, r *http.Request) {
	dbPool, err := db.GetPool()
	if err != nil {
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		msg := "User not found"
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}
	user, err := db.QueryUser(dbPool, userSesssion.UserID)
	if err != nil {
		http.Error(w, "error for user", http.StatusBadRequest)
		return
	}
	if user.TeamSize == 1 {
		renderCard(w, 0, user)
		return
	}
	err = db.PlusTeamSize(dbPool, user.ID, -1)
	if err != nil {
		http.Error(w, "user minus", http.StatusBadRequest)
		return
	}
	user.TeamSize -= 1
	w.WriteHeader(http.StatusOK)
}
