package main

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"

	"github.com/rs/cors"
)

func main() {
	// env
	config := GetConfig("./config/config.json")
	clientID := os.Getenv(config.Env.SpotifyID)
	clientSecret := os.Getenv(config.Env.SpotifySecret)
	timeLayout := "2006-01-02 15:04:05.999999999 -0700 MST"

	// cookies
	authStateCookie := GenerateCookie("auth_state")
	accessTokenCookie := GenerateCookie("access_token")
	refreshTokenCookie := GenerateCookie("refresh_token")
	tokenExpiryCookie := GenerateCookie("token_expiry")

	// router
	mux := http.NewServeMux()
	mux.Handle("/auth/login", &LoginHandler{
		clientID:        clientID,
		redirect:        config.Auth.Redirect,
		scope:           config.Auth.Scope,
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
		redirect:           config.Auth.Redirect,
		appURL:             config.AppURL,
	})
	mux.Handle("/me", &MeHandler{
		spotifyEndpoint:   "/me",
		accessTokenCookie: accessTokenCookie,
	})
	c := cors.New(cors.Options{
		AllowedOrigins:   []string{"http://localhost:8080"},
		AllowCredentials: true,
	})

	// Go!
	app := c.Handler(mux)
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
