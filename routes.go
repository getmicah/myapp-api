package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/boltdb/bolt"
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
		SendMethodError(w, r.Method)
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
		SendMethodError(w, r.Method)
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
	redirectURI        string
	appURL             string
}

func (h *CallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		callbackGet(w, r, h)
	default:
		SendMethodError(w, r.Method)
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
	tokenExpiryValue := accessTokenExpiry.Format(TimeLayout)
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
}

func (h *MeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		meGet(w, r, h)
	default:
		SendMethodError(w, r.Method)
	}

}

func meGet(w http.ResponseWriter, r *http.Request, h *MeHandler) {
	accessToken, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	res, err := SpotifyGet(r, h.spotifyEndpoint, accessToken)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

// StationHandler : POST /station
type StationHandler struct {
	authStateCookie    CookieID
	accessTokenCookie  CookieID
	refreshTokenCookie CookieID
	tokenExpiryCookie  CookieID
	clientID           string
	clientSecret       string
	redirectURI        string
	appURL             string
	db                 *bolt.DB
}

func (h *StationHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		stationGet(w, r, h)
	case "POST":
		stationPost(w, r, h)
	case "DELETE":
		stationDelete(w, r, h)
	default:
		SendMethodError(w, r.Method)
	}
}

func stationGet(w http.ResponseWriter, r *http.Request, h *StationHandler) {
	userID := r.URL.Query().Get("id")
	status, dbErr := GetStationStatus(h.db, userID)
	if dbErr != nil {
		SendError(w, http.StatusInternalServerError, dbErr.Error())
		return
	}
	var s Station
	s.Active = status
	body, err := json.Marshal(s)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func stationPost(w http.ResponseWriter, r *http.Request, h *StationHandler) {
	accessToken, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	res, err := SpotifyGet(r, "/me", accessToken)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var u User
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	dbErr := TurnOnStation(h.db, u.ID)
	if dbErr != nil {
		SendError(w, http.StatusInternalServerError, dbErr.Error())
		return
	}
	var s Station
	s.Active = true
	body, err := json.Marshal(s)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}

func stationDelete(w http.ResponseWriter, r *http.Request, h *StationHandler) {
	accessToken, err := LoadAccessToken(w, r, h.accessTokenCookie, h.refreshTokenCookie, h.tokenExpiryCookie, h.clientID, h.clientSecret)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	res, err := SpotifyGet(r, "/me", accessToken)
	if err != nil {
		SendError(w, http.StatusUnauthorized, err.Error())
		return
	}
	var u User
	if err := json.NewDecoder(res.Body).Decode(&u); err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	dbErr := TurnOffStation(h.db, u.ID)
	if dbErr != nil {
		SendError(w, http.StatusInternalServerError, dbErr.Error())
		return
	}
	var s Station
	s.Active = false
	body, err := json.Marshal(s)
	if err != nil {
		SendError(w, http.StatusInternalServerError, err.Error())
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	w.Write(body)
}
