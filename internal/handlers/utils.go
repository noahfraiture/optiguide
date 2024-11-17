package handlers

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"strconv"
)

// Return the parameter and write error status in http method if error
func GetParameterInt(w http.ResponseWriter, r *http.Request, name string) (int, error) {
	valueStr := r.URL.Query().Get(name)
	if valueStr == "" {
		msg := fmt.Sprintf("No %s: %s", name, valueStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return 0, fmt.Errorf("%s", msg)
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		msg := fmt.Sprintf("No %s: %s", name, valueStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return 0, fmt.Errorf("%s", msg)
	}
	return value, nil
}

// Return the parameter and write error status in http method if error
func GetParameterString(w http.ResponseWriter, r *http.Request, name string) (string, error) {
	valueStr := r.URL.Query().Get(name)
	if valueStr == "" {
		msg := fmt.Sprintf("No %s: %s", name, valueStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return "", fmt.Errorf("%s", msg)
	}
	return valueStr, nil
}

// Get the db, the session and the user from the db and write http error if there's an error
func GetUser(w http.ResponseWriter, r *http.Request) (db.User, error) {
	dbPool, err := db.GetPool()
	if err != nil {
		http.Error(w, "Can't get db", http.StatusInternalServerError)
		return db.User{}, err
	}
	userSesssion, err := auth.GetUser(r)
	if err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return db.User{}, nil
	}
	user, err := db.GetUser(dbPool, userSesssion.UserID)
	if err != nil {
		http.Error(w, "error for user", http.StatusBadRequest)
		return db.User{}, nil
	}
	return user, err
}
