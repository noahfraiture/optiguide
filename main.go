package main

import (
	"fmt"
	"log"
	"net/http"

	"optiguide/internal/db"
	"optiguide/internal/handlers/page/about"
	"optiguide/internal/handlers/page/characters"
	"optiguide/internal/handlers/page/guild"
	"optiguide/internal/handlers/page/home"
	"optiguide/internal/handlers/user"
)

func main() {

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
	http.HandleFunc("/card/search", home.SearchCards)
	http.HandleFunc("/card/toggle", user.Toggle)
	http.HandleFunc("/card/toggle-achievement", user.ToggleAchievement)

	// Characters
	http.HandleFunc("/characters", characters.Characters)

	// About
	http.HandleFunc("/about", about.About)
	http.HandleFunc("/about/submit-feedback", about.SubmitFeedback)

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

	http.HandleFunc("/user/edit-name", user.EditName)
	http.HandleFunc("/user/save-name", user.SaveName)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Start server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
