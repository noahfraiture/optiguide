package guild

import (
	"context"
	"fmt"
	"html/template"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"optiguide/internal/handlers"
	topbar "optiguide/internal/handlers/page"
)

type GuildData struct {
	LoggedIn bool
	Guilds   []db.Guild
}

var funcsGuild template.FuncMap = template.FuncMap{
	"emptyArr": func() []any { return []any{} },
	"map":      handlers.RenderMap,
}

func Guild(w http.ResponseWriter, r *http.Request) {
	funcs := funcsGuild
	for k, v := range topbar.FuncsTopbar {
		funcs[k] = v
	}

	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return
	}

	userAuth, err := auth.GetUser(r)
	loggedIn := err == nil
	var guilds []db.Guild
	if loggedIn {
		guilds, err = db.GetGuild(dbPool, userAuth.UserID)
		if err != nil && err != db.ErrNoGuild {
			fmt.Println(err)
			http.Error(w, "Error fetching guild", http.StatusInternalServerError)
			return
		}
	}

	tmpl, err := template.
		New("base.html").
		Funcs(funcs).
		ParseFiles(
			"templates/base.html",
			"templates/topbar.html",
			"templates/guild/guild.html",
		)
	if err != nil {
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "base.html", GuildData{
		LoggedIn: loggedIn,
		Guilds:   guilds,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		return
	}
}

func GuildSearch(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get db in guild search", http.StatusBadRequest)
		return
	}

	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusUnauthorized)
		return
	}

	guilds, err := db.SearchGuilds(dbPool, userAuth.UserID, name)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}

	tmpl, err := template.New("guild.html").Funcs(funcsGuild).ParseFiles("templates/guild/guild.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "search-results", map[string]any{
		"Guilds": guilds,
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		return
	}
}

func GuildCreate(w http.ResponseWriter, r *http.Request) {
	name := r.FormValue("name")
	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user in guild create", http.StatusUnauthorized)
		return
	}
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get db in guild create", http.StatusBadRequest)
		return
	}
	user, err := db.GetUser(dbPool, userAuth.UserID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't start transaction", http.StatusBadRequest)
		return
	}
	guildID, err := db.CreateGuild(tx, ctx, name)
	if err != nil {
		fmt.Println(err)
		fmt.Println(tx.Rollback(ctx))
		http.Error(w, "can't create guild", http.StatusBadRequest)
		return
	}
	err = db.JoinGuild(tx, ctx, guildID, userAuth.UserID)
	if err != nil {
		fmt.Println(err)
		fmt.Println(tx.Rollback(ctx))
		http.Error(w, "can't join guild", http.StatusBadRequest)
		return
	}
	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println(err)
		fmt.Println(tx.Rollback(ctx))
		http.Error(w, "can't commit transaction", http.StatusBadRequest)
		return
	}
	tmpl, err := template.
		New("guild").
		Funcs(funcsGuild).
		ParseFiles(
			"templates/guild/guild.html",
		)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to load template", http.StatusInternalServerError)
		return
	}
	err = tmpl.ExecuteTemplate(w, "guild", db.Guild{
		Name: name,
		ID:   guildID,
		Users: []db.GuildUser{{
			Email:    user.Email,
			TeamSize: user.TeamSize,
			Progress: 0,
		}},
	})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Unable to execute template", http.StatusInternalServerError)
		return
	}
}

func GuildJoin(w http.ResponseWriter, r *http.Request) {
	guildUUID, err := handlers.GetParameterUUID(w, r, "guild")
	if err != nil {
		return
	}
	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user in guild create", http.StatusUnauthorized)
		return
	}
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get db in guild create", http.StatusBadRequest)
		return
	}
	err = db.JoinGuild(dbPool, context.Background(), guildUUID, userAuth.UserID)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't join guild", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func GuildLeave(w http.ResponseWriter, r *http.Request) {
	guildUUID, err := handlers.GetParameterUUID(w, r, "id")
	if err != nil {
		return
	}
	userAuth, err := auth.GetUser(r)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get user in guild create", http.StatusUnauthorized)
		return
	}
	dbPool, err := db.GetPool()
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't get db in guild create", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	tx, err := dbPool.Begin(ctx)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "can't start transaction", http.StatusBadRequest)
		return
	}
	err = db.LeaveGuild(tx, ctx, guildUUID, userAuth.UserID)
	if err != nil {
		fmt.Println(err)
		fmt.Println(tx.Rollback(ctx))
		http.Error(w, "can't join guild", http.StatusBadRequest)
		return
	}
	err = db.DeleteGuildIfEmpty(tx, ctx, guildUUID)
	// TODO : ignore "nothing changed"
	if err != nil {
		fmt.Println(err)
		fmt.Println(tx.Rollback(ctx))
		http.Error(w, "can't join guild", http.StatusBadRequest)
		return
	}
	err = tx.Commit(ctx)
	if err != nil {
		fmt.Println(err)
		fmt.Println(tx.Rollback(ctx))
		http.Error(w, "can't commit transaction", http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}
