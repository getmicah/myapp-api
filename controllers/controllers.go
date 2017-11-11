package controllers

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"
	"os"
	"time"

	"github.com/getmicah/myapp-api/config"
)

// AppConfig : global app config
var AppConfig = config.Get("./config/config.json")

var (
	appURL                 = AppConfig.AppURL
	authorizeURL           = AppConfig.Auth.AuthorizeURL
	tokenURL               = AppConfig.Auth.TokenURL
	redirectURI            = AppConfig.Auth.Redirect
	scope                  = AppConfig.Auth.Scope
	authStateCookieName    = AppConfig.Cookie.SpotifyAuthState
	accessTokenCookieName  = AppConfig.Cookie.SpotifyAccessToken
	refreshTokenCookieName = AppConfig.Cookie.SpotifyRefreshToken
	tokenExpiryCookieName  = AppConfig.Cookie.SpotifyTokenExpiry
	clientID               = os.Getenv(AppConfig.Env.SpotifyID)
	clientSecret           = os.Getenv(AppConfig.Env.SpotifySecret)
)

// GenerateRandomString : create a random string with n length
func GenerateRandomString(n int) string {
	b := generateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b)
}

func generateRandomBytes(n int) []byte {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

// CreateCookie : create a cookie
func CreateCookie(name string, value string, expiry time.Time, httpOnly bool) *http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  expiry,
		HttpOnly: httpOnly,
	}
	return &cookie
}

// ClearCookie : set cookie(s) value empty
func ClearCookie(w http.ResponseWriter, name string) {
	expiry := time.Now().Add(-100 * time.Hour)
	cookie := CreateCookie(name, "", expiry, true)
	http.SetCookie(w, cookie)
}
