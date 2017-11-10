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
	appURL                 = "http://localhost:3000/authenticate" //AppConfig.AppURL
	authorizeURL           = AppConfig.Auth.AuthorizeURL
	tokenURL               = AppConfig.Auth.TokenURL
	redirectURI            = AppConfig.Auth.Redirect
	scope                  = AppConfig.Auth.Scope
	authStateCookieName    = AppConfig.Cookie.SpotifyAuthState
	accessTokenCookieName  = AppConfig.Cookie.SpotifyAccessToken
	refreshTokenCookieName = AppConfig.Cookie.SpotifyRefreshToken
	clientID               = os.Getenv(AppConfig.Env.SpotifyID)
	clientSecret           = os.Getenv(AppConfig.Env.SpotifySecret)
)

type token struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	Scope        string `json:"scope"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token"`
}

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

// CreateCookie : save cookie
func CreateCookie(name string, value string, expiry time.Time) *http.Cookie {
	cookie := http.Cookie{
		Name:     name,
		Value:    value,
		Path:     "/",
		Expires:  expiry,
		HttpOnly: true,
	}
	return &cookie
}

// WriteCookies : save cookie(s)
func WriteCookies(w http.ResponseWriter, cookies ...*http.Cookie) {
	for i := 0; i < len(cookies); i++ {
		http.SetCookie(w, cookies[i])
	}
}

// ClearCookies : set cookie(s) value empty
func ClearCookies(w http.ResponseWriter, cookieNames ...string) {
	expiry := time.Now().Add(365 * 24 * time.Hour)
	for i := 0; i < len(cookieNames); i++ {
		cookie := CreateCookie(cookieNames[i], "", expiry)
		http.SetCookie(w, cookie)
	}
}
