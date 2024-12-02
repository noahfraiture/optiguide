package auth

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/sessions"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func Init() error {
	// Load environment variables
	clientID, clientSecret, err := getGoogle()
	if err != nil {
		return err
	}
	clientCallbackURL := os.Getenv("GOOGLE_CALLBACK_URL")
	sessionSecret, err := getSession()
	if err != nil {
		return err
	}

	// Check that all necessary variables are set
	if clientID == "" || clientSecret == "" || clientCallbackURL == "" || sessionSecret == "" {
		log.Fatalf("Environment variables (GOOGLE_ID, GOOGLE_SECRET, GOOGLE_CALLBACK_URL, SESSION_SECRET) are required but got (%s, %s, %s, %s)", clientID, clientSecret, clientCallbackURL, sessionSecret)
	}

	store = sessions.NewCookieStore([]byte(sessionSecret))

	// Set up Google provider for Goth
	goth.UseProviders(google.New(clientID, clientSecret, clientCallbackURL))

	// Configure the session store with the session secret
	store := sessions.NewCookieStore([]byte(sessionSecret))
	gothic.Store = store
	return nil
}

func getGoogle() (string, string, error) {
	id, ok := os.LookupEnv("GOOGLE_ID")
	if !ok {
		idFile, ok := os.LookupEnv("GOOGLE_ID_FILE")
		if !ok {
			return "", "", fmt.Errorf("No password set")
		}
		data, err := os.ReadFile(idFile)
		if err != nil {
			return "", "", err
		}
		id = strings.TrimSpace(string(data))
	}
	secret, ok := os.LookupEnv("GOOGLE_SECRET")
	if !ok {
		secretFile, ok := os.LookupEnv("GOOGLE_SECRET_FILE")
		if !ok {
			return "", "", fmt.Errorf("No password set")
		}
		data, err := os.ReadFile(secretFile)
		if err != nil {
			return "", "", err
		}
		secret = strings.TrimSpace(string(data))
	}
	return id, secret, nil
}

func getSession() (string, error) {
	secret, ok := os.LookupEnv("SESSION_SECRET")
	if !ok {
		secretFile, ok := os.LookupEnv("SESSION_SECRET_FILE")
		if !ok {
			return "", fmt.Errorf("No password set")
		}
		data, err := os.ReadFile(secretFile)
		if err != nil {
			return "", err
		}
		secret = strings.TrimSpace(string(data))
	}
	return secret, nil

}
