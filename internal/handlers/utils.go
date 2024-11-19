package handlers

import (
	"fmt"
	"net/http"
	"optiguide/internal/auth"
	"optiguide/internal/db"
	"strconv"

	"github.com/google/uuid"
)

func RenderMap(args ...any) map[string]any {
	dict := make(map[string]any)
	for i := range len(args) / 2 {
		if v, ok := args[i*2].(string); ok {
			dict[v] = args[i*2+1]
		}
	}
	return dict
}

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

// Return the parameter and write error status in http method if error
func GetParameterUUID(w http.ResponseWriter, r *http.Request, name string) (uuid.UUID, error) {
	valueStr := r.URL.Query().Get(name)
	if valueStr == "" {
		msg := fmt.Sprintf("No %s: %s", name, valueStr)
		fmt.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return [16]byte{}, fmt.Errorf("%s", msg)
	}
	value, err := uuid.Parse(valueStr)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Can't parse uuid", http.StatusBadRequest)
		return [16]byte{}, err
	}
	return value, nil
}

// Get the db, the session and the user from the db and write http error if there's an error
func GetUser(w http.ResponseWriter, r *http.Request) (db.User, error) {
	dbPool, err := db.GetPool()
	if err != nil {
		http.Error(w, "Can't get db to get user", http.StatusInternalServerError)
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
