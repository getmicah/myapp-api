package controllers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Login : redirect user to Spotify login
func Login(w http.ResponseWriter, r *http.Request) {
	dur := 3600 * time.Second
	expiry := time.Now().Add(dur)
	state := GenerateRandomString(16)
	stateCookie, err := CreateEncryptedCookie(SecureStateCookie, AppConfig.Cookie.State, state, expiry)
	if err != nil {
		// set error header
		fmt.Println(err)
		return
	}
	http.SetCookie(w, stateCookie)
	api := "https://accounts.spotify.com/authorize/"
	redirect := AppConfig.Auth.Redirect
	scope := AppConfig.Auth.Scope
	authURL := fmt.Sprintf(
		"%s?client_id=%s&response_type=%s&redirect_uri=%s&scope=%s&state=%s",
		api, ClientID, "code", url.PathEscape(redirect), strings.Join(scope[:], "%20"), state,
	)
	http.Redirect(w, r, authURL, 302)
}
