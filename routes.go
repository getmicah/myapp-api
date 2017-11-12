package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LoginHandler : /auth/login
type LoginHandler struct {
	clientID        string
	redirectURI     string
	scope           []string
	authStateCookie CookieID
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	dur := 3600 * time.Second
	state := GenerateRandomString(16)
	expiry := time.Now().Add(dur)
	if err := WriteCookie(w, h.authStateCookie, state, expiry); err != nil {
		// set error header
		fmt.Println(err)
		return
	}
	api := "https://accounts.spotify.com/authorize/"
	authURL := fmt.Sprintf(
		"%s?client_id=%s&response_type=%s&redirect_uri=%s&scope=%s&state=%s",
		api, h.clientID, "code", url.PathEscape(h.redirectURI), strings.Join(h.scope[:], "%20"), state,
	)
	http.Redirect(w, r, authURL, 302)
}

// LogoutHandler : /auth/logout
type LogoutHandler struct {
	cookies []CookieID
	appURL  string
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	for i := 0; i < len(h.cookies); i++ {
		if err := ClearCookie(w, h.cookies[i]); err != nil {
			SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	http.Redirect(w, r, h.appURL, 302)
}

// CallbackHandler : /auth/callback
type CallbackHandler struct {
	authStateCookie    CookieID
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
	timeLayout         string
	redirectURI        string
	appURL             string
}

func (h *CallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	originalState, err := ReadCookie(r, h.authStateCookie)
	if err != nil {
		// error header
		fmt.Println(err)
		return
	}
	newState := r.URL.Query().Get("state")
	if newState != originalState {
		// error header
		fmt.Println("new state != original state")
		return
	}
	token, err := GetAuthToken(r, h.redirectURI, h.clientID, h.clientSecret)
	if err != nil {
		// error header
		fmt.Println(err)
		return
	}
	ClearCookie(w, h.authStateCookie)
	accessTokenExpiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	yearExpiry := time.Now().Add(365 * 24 * time.Hour)
	tokenExpiryValue := accessTokenExpiry.Format(h.timeLayout)
	WriteCookie(w, h.accessTokenCookie, token.AccessToken, accessTokenExpiry)
	WriteCookie(w, h.refreshTokenCookie, token.RefreshToken, yearExpiry)
	WriteCookie(w, h.tokenExpiryCookie, tokenExpiryValue, yearExpiry)
	http.Redirect(w, r, h.appURL, 302)
}

// MeHandler : /me
type MeHandler struct {
	spotifyEndpoint   string
	accessTokenCookie CookieID
}

func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	GetSpotify(w, r, h.spotifyEndpoint, h.accessTokenCookie)
}
