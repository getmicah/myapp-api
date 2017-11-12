package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/boltdb/bolt"
	"github.com/rs/cors"
)

func main() {
	// config
	config := getConfig("./config.json")
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	timeLayout := "2006-01-02 15:04:05.999999999 -0700 MST"
	scope := []string{
		"user-modify-playback-state",
		"user-read-currently-playing",
		"user-read-playback-state",
		"user-read-recently-played",
	}

	// database
	db, err := bolt.Open("bolt.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// cookies
	authStateCookie := GenerateCookie("auth_state")
	accessTokenCookie := GenerateCookie("access_token")
	refreshTokenCookie := GenerateCookie("refresh_token")
	tokenExpiryCookie := GenerateCookie("token_expiry")

	// router
	mux := http.NewServeMux()
	mux.Handle("/auth/login", &LoginHandler{
		clientID:        clientID,
		redirectURI:     config.RedirectURI,
		scope:           scope,
		authStateCookie: authStateCookie,
	})
	mux.Handle("/auth/logout", &LogoutHandler{
		cookies: []CookieID{
			authStateCookie,
			accessTokenCookie,
			refreshTokenCookie,
			tokenExpiryCookie,
		},
		appURL: config.AppURL,
	})
	mux.Handle("/auth/callback", &CallbackHandler{
		authStateCookie:    authStateCookie,
		accessTokenCookie:  accessTokenCookie,
		refreshTokenCookie: refreshTokenCookie,
		tokenExpiryCookie:  tokenExpiryCookie,
		clientID:           clientID,
		clientSecret:       clientSecret,
		timeLayout:         timeLayout,
		redirectURI:        config.RedirectURI,
		appURL:             config.AppURL,
	})
	mux.Handle("/me", &MeHandler{
		spotifyEndpoint:    "/me",
		accessTokenCookie:  accessTokenCookie,
		refreshTokenCookie: refreshTokenCookie,
		tokenExpiryCookie:  tokenExpiryCookie,
		clientID:           clientID,
		clientSecret:       clientSecret,
		timeLayout:         timeLayout,
	})
	mux.Handle("/station", &StationHandler{})

	// middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowCredentials: true,
	})
	app := c.Handler(mux)

	// Go!
	http.ListenAndServe(":3000", app)
}

// GenerateRandomString : create random string with n length
func GenerateRandomString(n int) string {
	b := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b)
}

// GenerateRandomBytes : create random []byte with n length
func GenerateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

type config struct {
	APIURL      string `json:"apiURL"`
	AppURL      string `json:"appURL"`
	RedirectURI string `json:"redirectURI"`
}

func getConfig(path string) config {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		panic(err)
	}
	var c config
	json.Unmarshal(file, &c)
	return c
}
