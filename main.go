package main

import (
	"fmt"
	"log"
	"net/http"

	"optiguide/internal/auth"
	"optiguide/internal/db"
	"optiguide/internal/handlers/page/about"
	"optiguide/internal/handlers/page/guild"
	"optiguide/internal/handlers/page/home"
	"optiguide/internal/handlers/user"
	"optiguide/internal/parser"
)

func main() {

	var err error
	cards, err := parser.Parse("guide.xlsx")
	if err != nil {
		log.Fatal(err)
	}

	auth.Init()
	err = db.Init()
	if err != nil {
		log.Fatal(err)
	}

	dbPool, err := db.GetPool()
	if err != nil {
		log.Fatal(err)
	}
	existingCards, err := db.GetCards(dbPool, 0)
	if err != nil {
		log.Fatal(err)
	}
	if len(existingCards) == 0 {
		err = db.InsertCards(dbPool, cards)
		if err != nil {
			log.Fatal(err)
		}
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

	// FIXME : needed to have tailwindcss
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	fmt.Println("Start server...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
