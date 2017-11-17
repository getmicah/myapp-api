package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/rs/cors"
)

const (
	// ClientTimeout : timeout for http.Client
	ClientTimeout = time.Second * 10
	// TimeLayout : format for converting time to and from string
	TimeLayout = "2006-01-02 15:04:05.999999999 -0700 MST"
)

func main() {
	// config
	config := getConfig("~/config.json")
	clientID := os.Getenv("SPOTIFY_CLIENT_ID")
	clientSecret := os.Getenv("SPOTIFY_CLIENT_SECRET")
	scope := []string{
		"playlist-modify-public",
	}

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
	mux.Handle("/auth", &AuthHandler{
		accessTokenCookie:  accessTokenCookie,
		refreshTokenCookie: refreshTokenCookie,
		tokenExpiryCookie:  tokenExpiryCookie,
		clientID:           clientID,
		clientSecret:       clientSecret,
	})
	mux.Handle("/search", &SearchHandler{
		accessTokenCookie:  accessTokenCookie,
		refreshTokenCookie: refreshTokenCookie,
		tokenExpiryCookie:  tokenExpiryCookie,
		clientID:           clientID,
		clientSecret:       clientSecret,
	})
	mux.Handle("/artist", &ArtistHandler{
		accessTokenCookie:  accessTokenCookie,
		refreshTokenCookie: refreshTokenCookie,
		tokenExpiryCookie:  tokenExpiryCookie,
		clientID:           clientID,
		clientSecret:       clientSecret,
	})
	mux.Handle("/track", &TrackHandler{
		accessTokenCookie:  accessTokenCookie,
		refreshTokenCookie: refreshTokenCookie,
		tokenExpiryCookie:  tokenExpiryCookie,
		clientID:           clientID,
		clientSecret:       clientSecret,
	})
	mux.Handle("/rec", &RecHandler{
		accessTokenCookie:  accessTokenCookie,
		refreshTokenCookie: refreshTokenCookie,
		tokenExpiryCookie:  tokenExpiryCookie,
		clientID:           clientID,
		clientSecret:       clientSecret,
	})
	mux.Handle("/playlist", &PlaylistHandler{
		accessTokenCookie:  accessTokenCookie,
		refreshTokenCookie: refreshTokenCookie,
		tokenExpiryCookie:  tokenExpiryCookie,
		clientID:           clientID,
		clientSecret:       clientSecret,
	})

	// middleware
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{config.AppURL},
		AllowCredentials: true,
		AllowedMethods:   []string{"GET", "POST", "DELETE"},
	})
	app := c.Handler(mux)

	// Go!
	if config.Production {
		http.ListenAndServeTLS(
			":443",
			"/etc/letsencrypt/live/api.micahcowell.com/fullchain.pem",
			"/etc/letsencrypt/live/api.micahcowell.com/privkey.pem",
			app,
		)
	} else {
		http.ListenAndServe(":3000", app)
	}
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
	Production  bool   `json:"production"`
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
