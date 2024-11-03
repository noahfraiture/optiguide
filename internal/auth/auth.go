package auth

import (
	"fmt"
	"log"
	"os"

	"github.com/gorilla/sessions"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func Init() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal(".env file failed to load!")
	}

	clientID := os.Getenv("GOOGLE_ID")
	clientSecret := os.Getenv("GOOGLE_SECRET")
	clientCallbackURL := os.Getenv("GOOGLE_CALLBACK_URL")
	sessionSecret := os.Getenv("SESSION_SECRET")

	// Check that all necessary variables are set
	if clientID == "" || clientSecret == "" || clientCallbackURL == "" || sessionSecret == "" {
		log.Fatal("Environment variables (GOOGLE_ID, GOOGLE_SECRET, GOOGLE_CALLBACK_URL, SESSION_SECRET) are required")
	}

	store = sessions.NewCookieStore([]byte(sessionSecret))

	// Set up Google provider for Goth
	goth.UseProviders(google.New(clientID, clientSecret, clientCallbackURL))

	// Configure the session store with the session secret
	store := sessions.NewCookieStore([]byte(sessionSecret))
	gothic.Store = store

	fmt.Println("Google OAuth initialized with session secret")
}
