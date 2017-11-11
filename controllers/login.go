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
	stateCookie := CreateCookie(authStateCookieName, state, expiry, true)
	http.SetCookie(w, stateCookie)
	authURL := fmt.Sprintf(
		"%s?client_id=%s&response_type=%s&redirect_uri=%s&scope=%s&state=%s",
		authorizeURL, clientID, "code", url.PathEscape(redirectURI), strings.Join(scope[:], "%20"), state,
	)
	http.Redirect(w, r, authURL, 302)
}
