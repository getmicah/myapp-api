package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/boltdb/bolt"
	"github.com/rs/cors"
)

const (
	// ClientTimeout : timeout for http.Client
	ClientTimeout = time.Second * 10
	// TimeLayout : format for converting time to and from string
	TimeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
	// StationsBucket : name of boltdb "Stations" bucket
	StationsBucket = "Stations"
	// StationActive : station active value
	StationActive = "on"
)

func main() {
	// config
	config := getConfig("./config.json")
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	scope := []string{
		"user-modify-playback-state",
		"user-read-currently-playing",
		"user-read-playback-state",
		"user-read-recently-played",
		"user-follow-read",
	}

	// database
	db, err := bolt.Open("bolt.db", 0600, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(StationsBucket))
		if err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

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
	})
	mux.Handle("/station", &StationHandler{
		authStateCookie:    authStateCookie,
		accessTokenCookie:  accessTokenCookie,
		refreshTokenCookie: refreshTokenCookie,
		tokenExpiryCookie:  tokenExpiryCookie,
		clientID:           clientID,
		clientSecret:       clientSecret,
		redirectURI:        config.RedirectURI,
		appURL:             config.AppURL,
		db:                 db,
	})

	// middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
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
