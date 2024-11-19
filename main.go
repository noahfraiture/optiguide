package main

import (
	"fmt"
	"log"
	"net/http"

	"optiguide/internal/auth"
	"optiguide/internal/db"
	"optiguide/internal/handlers/page/guild"
	"optiguide/internal/handlers/page/home"
	"optiguide/internal/handlers/user"
)

func main() {

	auth.Init()
	err := db.Init()
	if err != nil {
		log.Fatal(err)
	}

	// Home
	http.HandleFunc("/", home.Home)
	http.HandleFunc("/team/pick", home.PickCharacter)
	http.HandleFunc("/team/minus", home.Minus)
	http.HandleFunc("/team/plus", home.Plus)
	http.HandleFunc("/team/editable-name", home.RenderEditableName)
	http.HandleFunc("/team/save-name", home.SaveName)
	http.HandleFunc("/card", home.RenderCard)
	http.HandleFunc("/card/toggle", user.Toggle)

	// Guild
	http.HandleFunc("/guild", guild.Guild)
	http.HandleFunc("/guild/create", guild.GuildCreate)
	http.HandleFunc("/guild/search", guild.GuildSearch)
	http.HandleFunc("/guild/join", guild.GuildJoin)
	http.HandleFunc("/guild/leave", guild.GuildLeave)

	// TopBar
	http.HandleFunc("/auth/google/", user.GoogleLogin)
	http.HandleFunc("/auth/google/callback", user.GoogleCallback)
	http.HandleFunc("/logout", user.Logout)

	// FIXME : needed to have tailwindcss
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Start server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
