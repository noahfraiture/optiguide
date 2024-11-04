package auth

import (
	"fmt"
	"net/http"
	"optimax/internal/db"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
)

var store *sessions.CookieStore

func GetUser(r *http.Request) (goth.User, error) {
	session, err := store.Get(r, "user-session")
	if err != nil {
		return goth.User{}, err
	}
	userAny, ok := session.Values["user"]
	if !ok {
		return goth.User{}, fmt.Errorf("User not found in session")
	}
	user, ok := userAny.(goth.User)
	if !ok {
		return user, fmt.Errorf("User is not a goth.User")
	}
	return user, nil
}

func SaveUser(user goth.User, w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "user-session")
	if err != nil && !strings.Contains(err.Error(), "securecookie: the value is not valid") {
		return err
	}
	session.Values["user"] = user
	err = session.Save(r, w)
	if err != nil {
		return err
	}
	return db.InsertUser(db.User{ID: user.UserID, Email: user.Email})
}

func ClearSession(w http.ResponseWriter, r *http.Request) error {
	session, err := store.Get(r, "user-session")
	if err != nil {
		return err
	}
	session.Options.MaxAge = -1
	return session.Save(r, w)
}
