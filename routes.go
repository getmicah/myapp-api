package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// LoginHandler : GET /auth/login
type LoginHandler struct {
	clientID        string
	redirectURI     string
	scope           []string
	authStateCookie CookieID
}

func (h *LoginHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		loginGet(w, r, h)
	default:
		msg := fmt.Sprintf("Endpoint doesn't support %s request", r.Method)
		SendError(w, http.StatusBadRequest, msg)
	}
}

func loginGet(w http.ResponseWriter, r *http.Request, h *LoginHandler) {
	dur := 3600 * time.Second
	state := GenerateRandomString(16)
	expiry := time.Now().Add(dur)
	if err := WriteCookie(w, h.authStateCookie, state, expiry); err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
	}
	api := "https://accounts.spotify.com/authorize/"
	authURL := fmt.Sprintf(
		"%s?client_id=%s&response_type=%s&redirect_uri=%s&scope=%s&state=%s",
		api, h.clientID, "code", url.PathEscape(h.redirectURI), strings.Join(h.scope[:], "%20"), state,
	)
	http.Redirect(w, r, authURL, 302)
}

// LogoutHandler : GET /auth/logout
type LogoutHandler struct {
	cookies []CookieID
	appURL  string
}

func (h *LogoutHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		logoutGet(w, r, h)
	default:
		msg := fmt.Sprintf("Endpoint doesn't support %s request", r.Method)
		SendError(w, http.StatusBadRequest, msg)
	}
}

func logoutGet(w http.ResponseWriter, r *http.Request, h *LogoutHandler) {
	for i := 0; i < len(h.cookies); i++ {
		if err := ClearCookie(w, h.cookies[i]); err != nil {
			SendError(w, http.StatusInternalServerError, err.Error())
			return
		}
	}
	http.Redirect(w, r, h.appURL, 302)
}

// CallbackHandler : GET /auth/callback
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
	switch r.Method {
	case "GET":
		callbackGet(w, r, h)
	default:
		msg := fmt.Sprintf("Endpoint doesn't support %s request", r.Method)
		SendError(w, http.StatusBadRequest, msg)
	}
}

func callbackGet(w http.ResponseWriter, r *http.Request, h *CallbackHandler) {
	originalState, err := ReadCookie(r, h.authStateCookie)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	newState := r.URL.Query().Get("state")
	if newState != originalState {
		SendError(w, http.StatusUnauthorized, "Auth state compormised")
		return
	}
	callbackErr := r.URL.Query().Get("error")
	if callbackErr != "" {
		SendError(w, http.StatusUnauthorized, callbackErr)
		return
	}
	code := r.URL.Query().Get("code")
	token, err := RequestOAuthToken(r, code, h.redirectURI, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	accessTokenExpiry := time.Now().Add(time.Duration(token.ExpiresIn) * time.Second)
	yearExpiry := time.Now().Add(365 * 24 * time.Hour)
	tokenExpiryValue := accessTokenExpiry.Format(h.timeLayout)
	WriteCookie(w, h.accessTokenCookie, token.AccessToken, accessTokenExpiry)
	WriteCookie(w, h.refreshTokenCookie, token.RefreshToken, yearExpiry)
	WriteCookie(w, h.tokenExpiryCookie, tokenExpiryValue, yearExpiry)
	ClearCookie(w, h.authStateCookie)
	http.Redirect(w, r, h.appURL, 302)
}

// MeHandler : GET /me
type MeHandler struct {
	spotifyEndpoint    string
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
	timeLayout         string
}

func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		meGet(w, r, h)
	default:
		msg := fmt.Sprintf("Endpoint doesn't support %s request", r.Method)
		SendError(w, http.StatusBadRequest, msg)
	}

}

func meGet(w http.ResponseWriter, r *http.Request, h *MeHandler) {
	accessToken, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret, h.timeLayout)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	SpotifyGet(w, r, h.spotifyEndpoint, accessToken)
}

// StationHandler : POST /station
type StationHandler struct {
}

func (h *StationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		// Serve the resource.
	case "POST":
		// Create a new record.
	case "PUT":
		// Update an existing record.
	case "DELETE":
		// Remove the record.
	default:
		msg := fmt.Sprintf("Endpoint doesn't support %s request", r.Method)
		SendError(w, http.StatusBadRequest, msg)
	}
}
