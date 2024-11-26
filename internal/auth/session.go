package auth

import (
	"fmt"
	"net/http"
	"optiguide/internal/db"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/markbates/goth"
)

// TODO : move to db
var store *sessions.CookieStore

func GetUser(r *http.Request) (goth.User, error) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		return goth.User{}, err
	}
	userData, ok := session.Values["user"]
	if !ok {
		return goth.User{}, fmt.Errorf("User not found in session")
	}
	user, ok := userData.(goth.User)
	if !ok {
		return user, fmt.Errorf("User is not a goth.User")
	}
	return user, nil
}

func SaveUser(dbpool *pgxpool.Pool, user goth.User, w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "user-session")
	if err != nil && !strings.Contains(err.Error(), "securecookie: the value is not valid") {
		return err
	}
	session.Values["user"] = user
	err = session.Save(r, w)
	if err != nil {
		return err
	}
	_, err = db.GetUserFromProvider(dbpool, user)
	return err
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	fmt.Println("Clearing session...")
	session, err := store.Get(r, "user-session")
	if err != nil {
		fmt.Printf("Error getting session: %v\n", err)
		return err
	}
	session.Options.MaxAge = -1
	err = session.Save(r, w)
	if err != nil {
		fmt.Printf("Error saving session: %v\n", err)
		return err
	}
	http.SetCookie(w, &http.Cookie{
		Name:   "user-session",
		MaxAge: -1,
		Path:   "/",
	})
	fmt.Println("Session cleared successfully")
	return nil
}
